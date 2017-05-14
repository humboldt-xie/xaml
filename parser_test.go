package xaml

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"unicode"
)

const (
	T_ELEMENT = iota
	T_TEXT
)

type XamlArg struct {
	Key   string
	Value string
}

type XamlEle struct {
	Type   int
	Name   string
	Level  int
	Text   string
	Parent *XamlEle
	Args   []XamlArg
	Child  []*XamlEle
}

func (x *XamlEle) AddChild(c *XamlEle) {
	x.Child = append(x.Child, c)
	c.Parent = x
}

type Parser struct {
	Reader io.Reader
	Buf    []byte
	C      byte
	Has    bool
	Error  error
}

func (p *Parser) SetError(s string) {
	p.Error = fmt.Errorf("%s", s)
}

func (p *Parser) Next() (rune, bool) {
	if p.Error != nil {
		return 0, false
	}
	buf := make([]byte, 1)
	_, err := p.Reader.Read(buf)
	if err != nil {
		p.Error = err
		p.Has = false
		return 0, false
	} else {
		p.C = buf[0]
		p.Has = true
		return rune(p.C), true
	}

}

func (p *Parser) Cur() (rune, bool) {
	if p.Error != nil {
		return 0, false
	}
	if p.Has {
		return rune(p.C), true
	}
	p.Next()
	return p.Cur()

}
func (p *Parser) ParseEleString() string {
	c, ok := p.Cur()
	if !ok {
		return ""
	}
	if c != '"' {
		p.SetError(fmt.Sprintf("string start not \" %s", string(c)))
		return ""
	}

	c, ok = p.Next()
	if !ok {
		return ""
	}
	str := ""
	for c != '"' {
		str += string(c)
		c, ok = p.Next()
		if !ok {
			//error
			return ""
		}
	}
	//skip last "
	p.Next()
	return str
}
func (p *Parser) ParseTextEle(level int) *XamlEle {
	text := p.ParseEleString()
	child := XamlEle{Type: T_TEXT, Level: level, Text: text}
	return &child
}
func (p *Parser) ParseSkip(s rune) int {
	count := 0
	for {
		c, ok := p.Cur()
		if !ok {
			break
		}
		if c == s {
			count += 1
			p.Next()
		} else {
			break
		}
	}
	return count
}
func (p *Parser) ParseEleBody(ele *XamlEle) {
	for {
		p.ParseSkip(' ')
		c, ok := p.Cur()
		if !ok {
			return
		}
		switch c {
		case ':':
			p.Next()
			key := p.ParseEleName()
			p.ParseSkip(' ')
			value := p.ParseEleString()
			ele.Args = append(ele.Args, XamlArg{Key: key, Value: value})
		case '"':
			child := p.ParseTextEle(ele.Level + 1)
			ele.AddChild(child)
		default:
			return
		}
	}
	return
}
func (p *Parser) ParseEleName() string {
	c, ok := p.Cur()
	if !ok {
		return ""
	}
	eleName := ""
	for unicode.IsLetter(c) {
		eleName += string(c)
		p.Next()
		c, ok = p.Cur()
	}
	return eleName
}

func (p *Parser) ParseEle() *XamlEle {
	for {
		spaceCount := p.ParseSkip(' ')
		c, ok := p.Cur()
		if !ok {
			return nil
		}
		switch rune(c) {
		case '"':
			// parse text ele
			ele := p.ParseTextEle(spaceCount / 2)
			return ele
		default:
			eleName := p.ParseEleName()
			if p.Error != nil {
				return nil
			}
			ele := XamlEle{Type: T_ELEMENT, Level: spaceCount / 2, Name: eleName}
			p.ParseEleBody(&ele)
			if p.Error != nil {
				return nil
			}
			return &ele
		}
	}
}

func (p *Parser) Parse() *XamlEle {
	root := &XamlEle{Level: -1}
	for {
		ele := p.ParseEle()
		if p.Error != nil {
			break
			//return root
		}
		for ele.Level <= root.Level {
			root = root.Parent
		}
		if ele.Level > root.Level+1 {
			p.SetError("level great then 1")
			break
			//return root
		}

		if ele.Level == root.Level+1 {
			root.AddChild(ele)
			root = ele
		}
		end := p.ParseSkip('\n')
		if end <= 0 {
			p.SetError("element must end of \\n")
			break
		}
	}
	for root.Level >= 0 {
		root = root.Parent
	}
	return root
}

func (x *XamlEle) Render(w io.Writer) {
	if x.Level < 0 {
		for _, c := range x.Child {
			c.Render(w)
		}
		return
	}
	if x.Type == T_TEXT {
		w.Write([]byte(x.Text))
		return
	}
	w.Write([]byte("<" + x.Name + " "))
	for _, t := range x.Args {
		w.Write([]byte(t.Key + "=" + "\"" + t.Value + "\" "))
	}
	if x.Child == nil {
		w.Write([]byte("/>"))
		return
	} else {
		w.Write([]byte(">"))
	}
	for _, c := range x.Child {
		c.Render(w)
	}
	w.Write([]byte("</" + x.Name + ">"))
}

func TestParseEle(t *testing.T) {
	src := "div :id \"first div\" \"hello\"\n  a:href\"http://www.baidu.com\"\"baidu\"\n"
	buffer := &bytes.Buffer{}
	p := Parser{Reader: strings.NewReader(src)}
	ele := p.Parse()

	ele.Render(buffer)
	// output <div><li/>
	t.Fatal(ele, p.Error, string(buffer.Bytes()))
}

func TestParse(t *testing.T) {
	src := "div\nli\n"
	p := Parser{Reader: strings.NewReader(src)}
	target := ""
	for ; ; p.Next() {
		c, f := p.Cur()
		if !f {
			break
		}
		target += string(c)
		t.Log(string(c), f)
	}
	if src != string(target) {
		t.Fatal(len(src), "!=", len(string(target)), src[6], target[7])
	}
}

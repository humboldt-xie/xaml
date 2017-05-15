package xaml

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogin(t *testing.T) {
	ele, _ := ParseFile("testdata/login.html.xaml")
	t.Fatal(string(ele.Bytes()))
}

func TestParseEle(t *testing.T) {
	src := "div :id \"first div\" \"hello\"\n  a:href \"http://www.baidu.com\" \"baidu\"\n"
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
		t.Fatal(len(src), "!=", len(string(target)), target)
	}
}

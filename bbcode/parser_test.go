package bbcode

import (
	// "fmt"
	"bytes"
	"testing"

	"code.google.com/p/go.net/html"
)

// func TestParser(t *testing.T) {
// 	tok := TokenizeString("prefix [b]bold[/b] suffix [URL=http://www.google.com]blah[/url][i][/i]abc", -1)
// 	for t := tok.Next(); t != nil; t = tok.Next() {
// 		fmt.Printf("%+v\n", t)
// 	}
// }

func TestInvalidHTML(t *testing.T) {
	r := bytes.NewReader([]byte("<html><body><b><i>testing</b></i></body></html>"))
	h, err := html.Parse(r)
	if err != nil {
		t.Fatal(err)
	}
	w := &bytes.Buffer{}
	if err := html.Render(w, h); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s\n", w.String())
}

func TestToHTML(t *testing.T) {
	html, errs := BBCodeToHTML("prefix [i][b]bold talic[/i] just bold[/b] suffix & [url=http://www.google.com][img]http://example.com/some.jpg[/img] & xx[/url] abc")
	if errs != nil {
		t.Fatalf("BBCodeToHTML failed with errors: %+v", errs)
	}
	expected := `prefix <em><strong>bold talic</strong></em><strong> just bold</strong> suffix &amp; <a href="http://www.google.com"><img src="http://example.com/some.jpg"/> &amp; xx</a> abc`
	if html != expected {
		t.Fatalf("BBCodeToHTML returned wrong html:\n%s\ninstead of\n%s", html, expected)
	}
}

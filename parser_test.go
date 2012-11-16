package bbcode

import (
	// "fmt"
	"testing"
)

// func TestParser(t *testing.T) {
// 	tok := TokenizeString("prefix [b]bold[/b] suffix [URL=http://www.google.com]blah[/url][i][/i]abc", -1)
// 	for t := tok.Next(); t != nil; t = tok.Next() {
// 		fmt.Printf("%+v\n", t)
// 	}
// }

func TestToHTML(t *testing.T) {
	html, errs := BBCodeToHTML("prefix [b]bold[/b] suffix [url=http://www.google.com][img]http://example.com/some.jpg[/img][/url] abc")
	if errs != nil {
		t.Fatalf("BBCodeToHTML failed with errors: %+v", errs)
	}
	if html != `prefix <strong>bold</strong> suffix <a href="http://www.google.com"><img src="http://example.com/some.jpg"></a> abc` {
		t.Fatal("BBCodeToHTML returned wrong html")
	}
}

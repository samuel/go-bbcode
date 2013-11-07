package bbcode

import (
	"testing"
)

func testSimple(t *testing.T, bbcode, expectedHtml string) {
	html, errs := BBCodeToHTML(bbcode)
	if errs != nil {
		t.Fatalf("BBCodeToHTML failed on '%s' with errors: %+v", bbcode, errs)
	}
	if html != expectedHtml {
		t.Fatalf("BBCodeToHTML returned wrong html:\n%s\ninstead of the expected\n%s", html, expectedHtml)
	}
}

func testError(t *testing.T, bbcode string) {
	_, errs := BBCodeToHTML(bbcode)
	if errs == nil {
		t.Fatalf("BBCodeToHTML expected error for '%s'", bbcode)
	}
}

func TestToHTML(t *testing.T) {
	testSimple(t, "prefix [i][b]bold talic[/i] just bold[/b] suffix & [url=http://www.google.com][img]http://example.com/some.jpg[/img] & xx[/url] abc", `prefix <em><strong>bold talic</strong></em><strong> just bold</strong> suffix &amp; <a href="http://www.google.com"><img src="http://example.com/some.jpg"/> &amp; xx</a> abc`)
}

func TestB(t *testing.T) {
	testSimple(t, "[b]italic[/b]", "<strong>italic</strong>")
}

func TestI(t *testing.T) {
	testSimple(t, "[i]italic[/i]", "<em>italic</em>")
}

func TestUrl(t *testing.T) {
	testSimple(t, "[url=http://www.google.com]google[/url]", `<a href="http://www.google.com">google</a>`)
	testSimple(t, "[url]http://www.google.com/<foo>[/url]", `<a href="http://www.google.com/%3Cfoo%3E">http://www.google.com/%3Cfoo%3E</a>`)
	testError(t, "[url=www.google.com]google[/url]")
}

func TestImg(t *testing.T) {
	testSimple(t, "[img]http://www.google.com/logo.pcx[/img]", `<img src="http://www.google.com/logo.pcx"/>`)
	testSimple(t, "[IMG]http://www.google.com/logo.pcx[/IMG]", `<img src="http://www.google.com/logo.pcx"/>`)
}

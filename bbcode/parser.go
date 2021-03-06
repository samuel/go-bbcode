package bbcode

import (
	"bytes"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	html5 "code.google.com/p/go.net/html"
)

// [color=#ff0000]...[/color]
// [size=20%]...[/size]
// [:)]
// [list][*]...[*]...[/list]
//   [ul][li]{text}[/li][/ul]
// [table][tr][td]...[/td][/tr][/table]

// [XXX]
// [XXX=YYY]
// [XXX YYY=ZZZ UUU=VVV]
// [/XXX]

var (
	re = regexp.MustCompile(`(?Ui)\[(?:(/?)([a-z|\*]+)|([A-Za-z]+)=([^\]\s]+))\s*\]`)
)

type ErrUnknownTag string

func (e ErrUnknownTag) Error() string {
	return fmt.Sprintf("bbcode: unknown tag '%s'", string(e))
}

type ErrInvalidUrl string

func (e ErrInvalidUrl) Error() string {
	return fmt.Sprintf("bbcode: invalid url '%s'", string(e))
}

type ErrIncompleteTag string

func (e ErrIncompleteTag) Error() string {
	return fmt.Sprintf("bbcode: incomplete tag '%s'", string(e))
}

type Token struct {
	Text string

	Tag   string
	End   bool
	Value string
}

type Tokenizer struct {
	bbcode      string
	hits        [][]int
	checkpoints []int
	index       int
}

func TokenizeString(bbcode string, maxTags int) *Tokenizer {
	return &Tokenizer{
		bbcode:      bbcode,
		hits:        re.FindAllStringSubmatchIndex(bbcode, maxTags),
		checkpoints: make([]int, 0),
	}
}

func (t *Tokenizer) Begin() {
	t.checkpoints = append(t.checkpoints, t.index)
}

func (t *Tokenizer) Commit() {
	t.checkpoints = t.checkpoints[:len(t.checkpoints)-1]
}

func (t *Tokenizer) Rollback() {
	t.index = t.checkpoints[len(t.checkpoints)-1]
	t.Commit()
}

func (t *Tokenizer) Next() *Token {
	for t.index < len(t.hits)*2+1 {
		i := t.index / 2
		t.index++

		var idx []int
		if i < len(t.hits) {
			idx = t.hits[i]
		} else {
			idx = []int{len(t.bbcode), -1}
		}
		if t.index&1 == 1 {
			// text
			o := 0
			if i > 0 {
				o = t.hits[i-1][1]
			}
			txt := t.bbcode[o:idx[0]]
			if txt != "" {
				return &Token{Text: txt}
			}
		} else {
			tok := Token{}
			tok.Text = t.bbcode[idx[0]:idx[1]]
			if idx[2] >= 0 {
				// [tag] or [/tag]
				if idx[2] == idx[3]-1 {
					tok.End = true
				}
				tok.Tag = strings.ToLower(t.bbcode[idx[4]:idx[5]])
			} else {
				// [tag=value]
				tok.Tag = strings.ToLower(t.bbcode[idx[6]:idx[7]])
				tok.Value = t.bbcode[idx[8]:idx[9]]
			}
			return &tok
		}
	}
	return nil
}

func validateUrl(u string) (*url.URL, error) {
	ur, err := url.Parse(u)
	if err != nil {
		return nil, ErrInvalidUrl(u)
	}
	if !ur.IsAbs() || (ur.Scheme != "http" && ur.Scheme != "https") || ur.Host == "" {
		return nil, ErrInvalidUrl(u)
	}
	// TODO: Validate ur.Host is a public internet domain
	return ur, nil
}

func tokensToHTML(tok *Tokenizer) ([]string, []error) {
	bits := make([]string, 0, 32)
	var errors []error = nil
	// blockStack := make([]Token, 0, 32)
	inLink := false
	for t := tok.Next(); t != nil; t = tok.Next() {
		if t.Tag == "" {
			bits = append(bits, html.EscapeString(t.Text))
		} else {
			switch t.Tag {
			case "center":
				if t.End {
					bits = append(bits, "</span>")
				} else {
					bits = append(bits, `<span style="text-align:center;">`)
				}
			case "b":
				if t.End {
					bits = append(bits, "</strong>")
				} else {
					bits = append(bits, "<strong>")
				}
			case "i":
				if t.End {
					bits = append(bits, "</em>")
				} else {
					bits = append(bits, "<em>")
				}
			case "url":
				if t.End {
					if inLink {
						bits = append(bits, "</a>")
					} else {
						errors = append(errors, ErrIncompleteTag("url"))
					}
				} else {
					noLabel := false
					if t.Value == "" {
						t = tok.Next()
						if t == nil || t.Tag != "" {
							errors = append(errors, ErrIncompleteTag("url"))
							break
						}
						t.Value = t.Text
						noLabel = true
					}
					ur, err := validateUrl(t.Value)
					if err != nil {
						errors = append(errors, err)
					} else {
						inLink = true
						escapedUrl := html.EscapeString(ur.String())
						bits = append(bits, "<a href=\"", escapedUrl, "\">")
						if noLabel {
							bits = append(bits, escapedUrl)
						}
					}
				}
			case "img":
				if !t.End {
					tok.Begin()

					t = tok.Next()
					if t == nil {
						tok.Commit()
						errors = append(errors, ErrIncompleteTag("img"))
						return bits, errors
					}

					if t.Tag != "" {
						tok.Rollback()
						errors = append(errors, ErrInvalidUrl(t.Tag))
					} else {
						tok.Commit()
						ur, err := validateUrl(t.Text)
						if err != nil {
							errors = append(errors, err)
						} else {
							bits = append(bits, "<img src=\"", html.EscapeString(ur.String()), "\">")
						}
					}
				}
			}
		}
	}
	return bits, errors
}

func sanitizeHtml(ht string) (string, error) {
	node, err := html5.Parse(strings.NewReader(ht))
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := html5.Render(buf, node); err != nil {
		return "", err
	}
	str := buf.String()
	return str[25 : len(str)-14], nil
}

func BBCodeToHTML(bbcode string) (string, []error) {
	tok := TokenizeString(bbcode, 200)
	bits, errs := tokensToHTML(tok)
	html := strings.Join(bits, "")
	htmlS, err := sanitizeHtml(html)
	if err != nil {
		errs = append(errs, err)
	}
	return htmlS, errs
}

package bbcode

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
)

// [b]...[/b]
// [color=#ff0000]...[/color]
// [:)]
// [list][*]...[*]...[/list]
//   [ul][li]{text}[/li][/ul]
// [table][tr][td]...[/td][/tr][/table]
// [img width={width} height={height}]{url}[/img]

// [XXX]
// [XXX=YYY]
// [XXX YYY=ZZZ UUU=VVV]
// [/XXX]

var (
	re = regexp.MustCompile(`(?Ui)\[(?:(/?)([a-z|\*]+)|([A-Za-z]+)=([^\]\s]+))\s*\]`)
)

type ErrUnknownTag string

func (e ErrUnknownTag) Error() string {
	return fmt.Sprintf("{ErrUnknownTag %s}", string(e))
}

type ErrInvalidUrl string

func (e ErrInvalidUrl) Error() string {
	return fmt.Sprintf("{ErrInvalidUrl %s}", string(e))
}

type ErrIncompleteTag string

func (e ErrIncompleteTag) Error() string {
	return fmt.Sprintf("{ErrIncompleteTag %s}", string(e))
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
					ur, err := validateUrl(t.Value)
					if err != nil {
						errors = append(errors, err)
					} else {
						inLink = true
						bits = append(bits, "<a href=\"", ur.String(), "\">")
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
							bits = append(bits, "<img src=\"", ur.String(), "\">")
						}
					}
				}
			}
		}
	}
	return bits, errors
}

func BBCodeToHTML(bbcode string) (string, []error) {
	tok := TokenizeString(bbcode, 200)
	bits, errs := tokensToHTML(tok)
	return strings.Join(bits, ""), errs
}

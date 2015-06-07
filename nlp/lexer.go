package nlp

import (
	"regexp"
)

type Lexer struct {
	rules    []*Pair
	buffer   string
	beg, end int
	line     int
	text     string
	rem      []string
}

func NewLexer(rs []*Pair) *Lexer {
	this := Lexer{
		line: 0,
		text: "",
		beg:  0,
		end:  0,
	}

	this.rules = rs

	return &this
}

func (this *Lexer) getToken(stream string) int {
	token := 0
	for token == 0 {
		for this.beg == this.end {
			if stream != "" {
				this.line++
				this.beg = 0
				this.end = len(stream)
			} else {
				return 0
			}
		}

		found := false
		for i := 0; i < len(this.rules) && !found; i++ {
			rule := this.rules[i].first.(*regexp.Regexp)
			line := stream[this.beg:this.end]

			this.rem = RegExHasSuffix(rule, line)
			if len(this.rem) > 0 {
				if this.rem[0] != "" {
					token = this.rules[i].second.(int)
					this.text = this.rem[0]
					this.beg += len(this.rem[0])
					found = true
				} else {
					found = true
					this.beg++
				}
			}
		}

		if !found || this.beg >= this.end {
			token = -1
			this.text = stream[this.beg:this.end]
		}
	}
	return token
}

func (this *Lexer) getText() string { return this.text }
func (this *Lexer) lineno() int     { return this.line }

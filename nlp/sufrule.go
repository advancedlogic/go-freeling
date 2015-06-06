package nlp

import "regexp"

type sufrule struct {
	term, output, retok, lema, expression string
	cond                                  *regexp.Regexp
	acc, enc, always, nomore              int
}

func NewEmptySufRule() *sufrule {
	return &sufrule{}
}

func NewSufRuleFromRexEx(c string) *sufrule {
	return &sufrule{
		expression: c,
		cond:       regexp.MustCompile(c),
	}
}

func NewSufRuleFromSufRule(c *sufrule) *sufrule {
	return &sufrule{
		term:       c.term,
		output:     c.output,
		retok:      c.retok,
		lema:       c.lema,
		expression: c.expression,
		cond:       c.cond,
		acc:        c.acc,
		enc:        c.enc,
		always:     c.always,
		nomore:     c.nomore,
	}
}

package nlp

import (
	"github.com/fatih/set"
	"regexp"
)

type AccentsModule interface {
	FixAccentuation(*set.Set, *sufrule)
}

type AccentsDefault struct {
}

func NewAccetsDefault() *AccentsDefault {
	LOG.Trace("Create default accent handler")
	return &AccentsDefault{}
}

func (this *AccentsDefault) FixAccentuation(candidates *set.Set, suf *sufrule) {
	LOG.Trace("Default accentuation. Candidates " + candidates.String())
}

type AccentsES struct {
	llanaAcc        *regexp.Regexp
	agudaMal        *regexp.Regexp
	monosil         *regexp.Regexp
	lastVowelPutAcc *regexp.Regexp
	lastVowelNotAcc *regexp.Regexp
	anyVowelAcc     *regexp.Regexp

	withAcc    map[string]string
	withoutAcc map[string]string
}

func NewAccentsES() *AccentsES {
	return &AccentsES{
		withAcc:    make(map[string]string),
		withoutAcc: make(map[string]string),
	}
}

func (this *AccentsES) FixAccentuation(candidates *set.Set, suf *sufrule) {
	LOG.Trace("ES accentuation. Candidates " + candidates.String())
}

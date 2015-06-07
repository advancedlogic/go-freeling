package nlp

import (
	"container/list"
	set "gopkg.in/fatih/set.v0"
	"regexp"
	"strconv"
	"strings"
)

const (
	TOKENIZER_MACROS = 1 + iota
	TOKENIZER_REGEXPS
	TOKENIZER_ABBREV
)

type Tokenizer struct {
	abrevs  *set.Set
	rules   *list.List
	matches map[string]int
}

func NewTokenizer(tokenizerFile string) *Tokenizer {
	this := Tokenizer{
		abrevs:  set.New(),
		rules:   list.New(),
		matches: make(map[string]int),
	}

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("Macros", TOKENIZER_MACROS)
	cfg.AddSection("RegExps", TOKENIZER_REGEXPS)
	cfg.AddSection("Abbreviations", TOKENIZER_ABBREV)

	if !cfg.Open(tokenizerFile) {
		LOG.Panic("Error opening file " + tokenizerFile)
	}

	macros := list.New()
	rul := false
	var ci string
	line := ""
	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
		case TOKENIZER_MACROS:
			{
				if rul {
					LOG.Panic("Error reading tokenizer configuration. Macros must be defined before rules.")
				}
				mname := items[0]
				mvalue := items[1]
				macros.PushBack(Pair{mname, mvalue})
				LOG.Trace("Read macro " + mname + ": " + mvalue)
				break
			}
		case TOKENIZER_REGEXPS:
			{
				var substr int
				comm := items[0]
				substr, _ = strconv.Atoi(items[1])
				re := items[2]
				rul = true

				for i := macros.Front(); i != nil; i = i.Next() {
					mname := "{" + i.Value.(Pair).first.(string) + "}"
					mvalue := i.Value.(Pair).second.(string)
					p := strings.Index(re, mname)
					for p > -1 {
						re = strings.Replace(re, mname, mvalue, -1)
						p = strings.Index(re[p:], mname)
					}
				}

				if len(items) > 3 {
					ci = items[3]
				}

				if ci == "CI" {
					newre := "(?i)" + re
					x, err := regexp.Compile(newre)
					if err == nil {
						this.rules.PushBack(Pair{comm, x})
					} else {
						LOG.Warn("Rule " + comm + " [" + newre + "] failed to be compiled")
					}
				} else {
					x, err := regexp.Compile(re)
					if err == nil {
						this.rules.PushBack(Pair{comm, x})
					} else {
						LOG.Warn("Rule " + comm + " [" + re + "] failed to be compiled")
					}
				}

				this.matches[comm] = substr
				LOG.Trace("Stored rule " + comm + " " + re + " " + strconv.Itoa(substr))
				break

			}
		case TOKENIZER_ABBREV:
			{
				this.abrevs.Add(line)
				break
			}
		default:
			break
		}
	}

	LOG.Trace("analyzer succesfully created")
	return &this
}

func (this *Tokenizer) Tokenize(p string, offset int, v *list.List) {
	var t [10]string
	var i *list.Element
	var match bool
	substr := 0
	ln := 0

	v = v.Init()

	cont := 0
	for cont < len(p) {
		for WhiteSpace(p[cont]) {
			cont++
			offset++
		}
		LOG.Trace("Tokenizing [" + p[cont:] + "]")
		match = false

		for i = this.rules.Front(); i != nil && !match; i = i.Next() {
			LOG.Trace("Checking rule " + i.Value.(Pair).first.(string))
			ps := strings.Index(p[cont:], " ")
			delta := cont + ps
			if ps == -1 {
				delta = cont + len(p) - cont
			}
			results := RegExHasSuffix(i.Value.(Pair).second.(*regexp.Regexp), p[cont:delta])
			if len(results) > 0 {
				match = true
				ln = 0
				substr = this.matches[i.Value.(Pair).first.(string)]
				for j := If(substr == 0, 0, 1).(int); j <= substr && match; j++ {
					t[j] = results[j]
					ln += len(t[j])
					LOG.Trace("Found match " + strconv.Itoa(j) + " [" + t[j] + "] for rule " + i.Value.(Pair).first.(string))
					if string(i.Value.(Pair).first.(string)[0]) == "*" {
						lower := strings.ToLower(t[j])
						if !this.abrevs.Has(lower) {
							match = false
							LOG.Trace("Special rule and found match not in abbrev list. Rule not satisfied")
						}
					}
				}
			}

		}

		if match {
			if i == nil {
				i = this.rules.Back()
			} else {
				i = i.Prev()
			}
			substr = this.matches[i.Value.(Pair).first.(string)]
			for j := If(substr == 0, 0, 1).(int); j <= substr && match; j++ {
				if len(t[j]) > 0 {
					LOG.Trace("Accepting matched substring [" + t[j] + "]")
					w := NewWordFromLemma(t[j])
					w.setSpan(offset, offset+len(t[j]))
					offset += len(t[j])
					v.PushBack(w)
				} else {
					LOG.Trace("Skipping matched null substring [" + t[j] + "]")
				}
			}
			cont += ln
		} else if cont < len(p) {
			LOG.Warn("No rule matched input substring" + p[cont:] + " . Character " + string(p[cont:][0]) + " skipped . Check your tokenization rules")
			cont++
		}
	}

	offset++
}

package nlp

import (
	"container/list"
	"github.com/fatih/set"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	SUF  = 0
	PREF = 1
)

type Affixes struct {
	accen          *Accent
	affix          [2]map[string]*set.Set
	affixAlways    [2]map[string]*set.Set
	ExistingLength [2]*set.Set
	Longest        [2]int
}

func NewAffixes(sufFile string) *Affixes {
	this := Affixes{}

	filestr, err := ioutil.ReadFile(sufFile)
	if err != nil {
		CRASH("Error opening file "+sufFile, MOD_AFFIX)
		return nil
	}
	lines := strings.Split(string(filestr), "\n")

	this.Longest[SUF] = 0
	this.Longest[PREF] = 0

	kind := -1
	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, "#") {
			items := Split(line, "\t")
			if line == "<Suffixes>" {
				kind = SUF
			} else if line == "<Prefixes>" {
				kind = PREF
			} else if line == "</Suffixes>" {
				kind = -1
			} else if line == "</Prefixes>" {
				kind = -1
			} else if kind == SUF || kind == PREF {
				key := items[0]
				term := items[1]
				cond := items[2]
				output := items[3]
				acc := items[4]
				enc := items[5]
				nomore := items[6]
				lema := items[7]
				always := items[8]
				retok := items[9]

				suf := NewSufRuleFromRexEx(cond)
				suf.term = term
				suf.output = output
				suf.acc, _ = strconv.Atoi(acc)
				suf.enc, _ = strconv.Atoi(enc)
				suf.nomore, _ = strconv.Atoi(nomore)
				suf.lema = lema
				suf.always, _ = strconv.Atoi(always)
				suf.retok = retok

				if suf.retok == "-" {
					suf.retok = ""
				}

				if this.affix[kind] == nil {
					this.affix[kind] = make(map[string]*set.Set)
				}

				if this.affix[kind][key] == nil {
					this.affix[kind][key] = set.New(set.ThreadSafe).(*set.Set)
				}

				this.affix[kind][key].Add(suf)
				if suf.always == 1 {
					if this.affixAlways[kind] == nil {
						this.affixAlways[kind] = make(map[string]*set.Set)
					}
					if this.affixAlways[kind][key] == nil {
						this.affixAlways[kind][key] = set.New(set.ThreadSafe).(*set.Set)
					}
					this.affixAlways[kind][key].Add(suf)
				}

				if this.ExistingLength[kind] == nil {
					this.ExistingLength[kind] = set.New(set.ThreadSafe).(*set.Set)
				}

				this.ExistingLength[kind].Add(len(key))

				if len(key) > this.Longest[kind] {
					this.Longest[kind] = len(key)
				}
			}

		}
	}

	TRACE(3, "analyzer succesfully created", MOD_AFFIX)

	return &this
}

func (this *Affixes) lookFowAffixes(w *Word, dic *Dictionary) {
	if w.getNAnalysis() > 0 {
		TRACE(2, "=== Known word "+w.getForm()+", with "+strconv.Itoa(w.getNAnalysis())+" analysis. Looking only for 'always' affixes", MOD_AFFIX)
		TRACE(3, "--- Checkin SUF ---", MOD_AFFIX)
		this.lookForAffixesInList(SUF, this.affixAlways[SUF], w, dic)
		TRACE(3, "--- Checkin PREF ---", MOD_AFFIX)
		this.lookForAffixesInList(PREF, this.affixAlways[PREF], w, dic)
		TRACE(3, "--- Checkin SUF+PREF ---", MOD_AFFIX)
		this.lookForCombinedAffixes(this.affixAlways[SUF], this.affixAlways[PREF], w, dic)
	} else {
		TRACE(2, "=== Unknown word "+w.getForm()+", Looking only for any affixes", MOD_AFFIX)
		TRACE(3, "--- Checkin SUF ---", MOD_AFFIX)
		this.lookForAffixesInList(SUF, this.affix[SUF], w, dic)
		TRACE(3, "--- Checkin PREF ---", MOD_AFFIX)
		this.lookForAffixesInList(PREF, this.affix[PREF], w, dic)
		TRACE(3, "--- Checkin SUF+PREF ---", MOD_AFFIX)
		this.lookForCombinedAffixes(this.affix[SUF], this.affix[PREF], w, dic)
	}
}

func (this *Affixes) lookForAffixesInList(kind int, suff map[string]*set.Set, w *Word, dic *Dictionary) {
	var i, j, ln int
	var lws, formTerm, formRoot string
	lws = w.getLCForm()
	ln = len(lws)

	var rules *set.Set

	for i = 1; i <= this.Longest[kind] && i < ln; i++ {
		j = ln - i
		if this.ExistingLength[kind] == nil {
			TRACE(4, "No affixes  of size "+strconv.Itoa(i), MOD_AFFIX)
		} else {
			if kind == SUF {
				formTerm = lws[j:]
			} else if kind == PREF {
				formTerm = lws[0:i]
			}

			rules = suff[formTerm]
			if rules == nil || rules.Size() == 0 {
				TRACE(3, "No rules for affix "+formTerm+" (size "+strconv.Itoa(i), MOD_AFFIX)
			} else {
				TRACE(3, "Found "+strconv.Itoa(rules.Size())+" rules for affix "+formTerm+" (size "+strconv.Itoa(i), MOD_AFFIX)
				lrules := rules.List()
				for s := 0; s < rules.Size(); s++ {
					sufit := lrules[s].(*sufrule)
					if kind == SUF {
						formRoot = lws[0:j]
					} else if kind == PREF {
						formRoot = lws[i:]
					}
					TRACE(3, "Trying rule ["+formTerm+" "+sufit.term+" "+sufit.expression+" "+sufit.output+"] on root "+formRoot, MOD_AFFIX)
					candidates := this.GenerateRoots(kind, sufit, formRoot)
					this.accen.FixAccentutation(candidates, sufit)
					this.SearchRootsList(candidates, formTerm, sufit, w, dic)
				}
			}
		}
	}
}

func (this *Affixes) lookForCombinedAffixes(suff map[string]*set.Set, pref map[string]*set.Set, w *Word, dic *Dictionary) {
	var i, j, ln int
	var lws, formSuf, formPref, formRoot string
	lws = w.getLCForm()
	ln = len(lws)

	var rulesS *set.Set
	var rulesP *set.Set

	var candidates, cand1 *set.Set

	for i = 1; i <= this.Longest[SUF] && i < ln; i++ {
		if this.ExistingLength[SUF].Has(i) == false {
			TRACE(4, "No suffixes  of size "+strconv.Itoa(i), MOD_AFFIX)
			continue
		}

		for j = 1; j <= this.Longest[PREF] && j <= ln-i; j++ {
			if this.ExistingLength[PREF].Has(j) == false {
				TRACE(4, "No prefixes  of size "+strconv.Itoa(i), MOD_AFFIX)
				continue
			}

			formSuf = lws[ln-i:]
			formPref = lws[0:j]

			rulesS = suff[formSuf]
			if rulesS.Size() == 0 || rulesS.List()[0].(string) != formSuf {
				TRACE(3, "No rules for suffix "+formSuf+" (size "+strconv.Itoa(i), MOD_AFFIX)
				continue
			}

			rulesP = suff[formPref]
			if rulesP.Size() == 0 || rulesP.List()[0].(string) != formPref {
				TRACE(3, "No rules for prefix "+formPref+" (size "+strconv.Itoa(i), MOD_AFFIX)
				continue
			}

			formRoot = lws[0 : ln-i][j:]
			TRACE(3, "Trying a decomposition: "+formPref+"+"+formRoot+"+"+formSuf, MOD_AFFIX)

			TRACE(3, "Found "+strconv.Itoa(rulesS.Size())+" rules for suffix "+formSuf+" (size "+strconv.Itoa(i), MOD_AFFIX)
			TRACE(3, "Found "+strconv.Itoa(rulesP.Size())+" rules for prefix "+formPref+" (size "+strconv.Itoa(i), MOD_AFFIX)

			//wfid := w.foundInDict()

			lrulesS := rulesS.List()
			lrulesP := rulesP.List()

			for s := 0; s < rulesS.Size(); s++ {
				sufit := lrulesS[s].(*sufrule)
				for p := 0; p < rulesP.Size(); p++ {
					prefit := lrulesP[p].(*sufrule)
					candidates = set.New(set.ThreadSafe).(*set.Set)
					cand1 = this.GenerateRoots(SUF, sufit, formRoot)
					this.accen.FixAccentutation(cand1, sufit)
					lcand1 := cand1.List()
					for _, c := range lcand1 {
						cand2 := this.GenerateRoots(PREF, prefit, c.(string))
						this.accen.FixAccentutation(cand2, prefit)
						candidates.Add(cand2)
					}
				}
			}
		}
	}
}

func (this *Affixes) GenerateRoots(kind int, suf *sufrule, rt string) *set.Set {
	cand := set.New(set.ThreadSafe).(*set.Set)
	var term, r string
	var pe int

	cand.Clear()
	term = suf.term
	TRACE(3, "Possible terminations/beginnings: "+term, MOD_AFFIX)
	pe = strings.Index(term, "|")
	for pe > -1 {
		r = term[0:pe]
		if r == "*" {
			r = ""
		}

		if kind == SUF {
			TRACE(3, "Adding to t_roots the element "+rt+r, MOD_AFFIX)
			cand.Add(rt + r)
		} else if kind == PREF {
			TRACE(3, "Adding to t_roots the element "+r+rt, MOD_AFFIX)
			cand.Add(r + rt)
		}

		term = term[pe+1:]
		pe = strings.Index(term, "|")
	}

	if term == "*" {
		term = ""
	}

	if kind == SUF {
		TRACE(3, "Adding to t_roots the element "+rt+term, MOD_AFFIX)
		cand.Add(rt + term)
	} else if kind == PREF {
		TRACE(3, "Adding to t_roots the element "+term+rt, MOD_AFFIX)
		cand.Add(term + rt)
	}

	return cand
}

func (this *Affixes) SearchRootsList(roots *set.Set, aff string, suf *sufrule, wd *Word, dic *Dictionary) {
	TRACE(3, "Checking a list of "+strconv.Itoa(roots.Size())+" roots", MOD_AFFIX)

	remain := roots.List()
	for len(remain) > 0 {
		r := remain[0]
		la := list.New()
		dic.SearchForm(r.(string), la)

		if la.Len() == 0 {
			TRACE(3, "Root "+r.(string)+" not found", MOD_AFFIX)
			roots.Remove(r)
		} else {
			TRACE(3, "Root "+r.(string)+" found in dictionary", MOD_AFFIX)
			this.ApplyRule(r.(string), la, aff, suf, wd, dic)
		}

		remain = remain[1:]
	}

}

func (this *Affixes) ApplyRule(r string, la *list.List, aff string, suf *sufrule, wd *Word, dic *Dictionary) {
	var tag, lem string
	var a *Analysis
	for pos := la.Front(); pos != nil; pos = pos.Next() {
		if suf.cond.FindString(pos.Value.(*Analysis).getTag()) == "" {
			TRACE(3, "Tag "+pos.Value.(*Analysis).getTag()+"fails input condition "+suf.expression, MOD_AFFIX)
		} else {
			TRACE(3, "Tag "+pos.Value.(*Analysis).getTag()+"satisfies input condition "+suf.expression, MOD_AFFIX)
			if suf.nomore != 0 {
				wd.setFoundInDict(true)
			}

			if suf.output == "*" {
				TRACE(3, "Output tag as found in dictionary", MOD_AFFIX)
				tag = pos.Value.(*Analysis).getTag()
			} else {
				tag = suf.output
			}

			suflem := *list.New()
			tmpItems := Split(suf.lema, "+")
			for _, tmpItem := range tmpItems {
				suflem.PushBack(tmpItem)
			}
			lem = ""
			for s := suflem.Front(); s != nil; s = s.Next() {
				if string(s.Value.(string)[0]) == "F" {
					TRACE(3, "Output lemma: add original word form", MOD_AFFIX)
					lem = lem + wd.getLCForm()
				} else if string(s.Value.(string)[0]) == "R" {
					TRACE(3, "Output lemma: add root found in dictionary", MOD_AFFIX)
					lem = lem + r
				} else if string(s.Value.(string)[0]) == "L" {
					TRACE(3, "Output lemma: add lemma found in dictionary", MOD_AFFIX)
					lem = lem + pos.Value.(*Analysis).getLemma()
				} else if string(s.Value.(string)[0]) == "A" {
					TRACE(3, "Output lemma: add affix", MOD_AFFIX)
					lem = lem + aff
				} else {
					TRACE(3, "Output lemma: add string "+s.Value.(string), MOD_AFFIX)
					lem = lem + s.Value.(string)
				}
			}

			TRACE(3, "Analysis for the affixed form "+r+" ("+lem+","+tag+")", MOD_AFFIX)

			var rtk *list.List

			this.CheckRetokenizable(suf, r, lem, tag, dic, rtk, Capitalization(wd.getForm()))

			var p *list.Element
			for p = wd.Front(); p != nil && !(p.Value.(*Analysis).getLemma() == lem && p.Value.(*Analysis).getTag() == tag); p = p.Next() {
			}

			if p == nil {
				a.init(lem, tag)
				a.setRetokenizable(rtk)
				wd.addAnalysis(a)
			} else {
				TRACE(3, "Analysis was already three, adding RTK info", MOD_AFFIX)
				if rtk.Len() > 0 && p.Value.(*Analysis).isRetokenizable() {
					p.Value.(*Analysis).setRetokenizable(rtk)
				}
			}
		}
	}
}

func (this *Affixes) CheckRetokenizable(suf *sufrule, form string, lem string, tag string, dic *Dictionary, rtk *list.List, caps int) {
	TRACE(3, "Check retokenizable.", MOD_AFFIX)
	if suf.retok != "" {
		TRACE(3, " - sufrule has RTK: "+suf.retok, MOD_AFFIX)
		i := strings.Index(suf.retok, ":")

		forms := list.New()
		tmpItems := Split(suf.retok[0:i], "+")
		for _, tmpItem := range tmpItems {
			forms.PushBack(tmpItem)
		}
		tags := list.New()
		tmpItems = Split(suf.retok[i+1:], "+")
		for _, tmpItem := range tmpItems {
			forms.PushBack(tmpItem)
		}

		a := &Analysis{}
		first := true
		var k, j *list.Element

		for k = forms.Front(); k != nil; k = k.Next() {
			for j = tags.Front(); j != nil; j = j.Next() {
				w := NewWordFromLemma("")
				if k.Value.(string) == "$$" {
					w.setForm(Capitalize(form, caps, first))
					a.init(lem, tag)
					w.addAnalysis(a)
				} else {
					var la *list.List
					w.setForm(Capitalize(k.Value.(string), caps, first))
					dic.SearchForm(k.Value.(string), la)
					for a := la.Front(); a != nil; a = a.Next() {
						if strings.Index(a.Value.(*Analysis).getTag(), j.Value.(string)) > -1 {
							w.addAnalysis(a.Value.(*Analysis))
						}
					}
				}

				rtk.PushBack(w)
				TRACE(3, "    word "+w.getForm()+" ("+w.getLemma(0)+","+w.getTag(0)+") added to decomposition list", MOD_AFFIX)
				first = false
			}
		}
	}
}

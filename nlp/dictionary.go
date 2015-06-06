package nlp

import (
	"container/list"
	"strings"
	"strconv"
	set "gopkg.in/fatih/set.v0"
	"math"
	//"os"
)

const (
	DICTIONARY_INDEX = 1 + iota
	DICTIONARY_LEMMA_PREF
	DICTIONARY_POS_PREF
	DICTIONARY_ENTRIES
)

const (
	TAG_DIVIDER = "|"
	LEMMA_DIVIDER = " "
)

type Dictionary struct {
	InverseDic             bool
	RetokenizeContractions bool
	AffixAnalysis          bool
	CompoundAnalysis       bool
	suf                    *Affixes
	comp                   *Compound

	morfodb *Database
	inverdb *Database

	lemmaPrefs map[string]string
	posPrefs   map[string]string
}

func NewDictionary(Lang string, dicFile string, sufFile string, compFile string, invDic bool, retok bool) *Dictionary {
	this := Dictionary{}

	this.InverseDic = invDic
	this.RetokenizeContractions = retok

	this.suf = nil

	if sufFile != "" {
		this.suf = NewAffixes(sufFile)
	}

	this.AffixAnalysis = (this.suf != nil)

	this.comp = nil

	if compFile != "" {
		this.comp = &Compound{} //TODO
	}
	this.CompoundAnalysis = (this.comp != nil)

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("IndexType", DICTIONARY_INDEX)
	cfg.AddSection("LemmaPreferences", DICTIONARY_LEMMA_PREF)
	cfg.AddSection("PosPreferences", DICTIONARY_POS_PREF)
	cfg.AddSection("Entries", DICTIONARY_ENTRIES)

	if !cfg.Open(dicFile) {
		LOG.Panic("Error opening file "+dicFile)
	}

	this.morfodb = nil
	this.inverdb = nil

	line := ""

	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
			case DICTIONARY_INDEX: {
				tpe := -1
				if line == "DB_PREFTREE" {
					tpe = DB_PREFTREE
				} else if line == "DB_MAP" {
					tpe = DB_MAP
				} else {
					LOG.Panic("Invalid IndexType '" + line + "' specified in dictionary file " + dicFile)
				}

				this.morfodb = NewDatabase(tpe)
				if this.InverseDic {
					this.inverdb = NewDatabase(DB_MAP)
				}
				break
			}
			case DICTIONARY_LEMMA_PREF: {
				lem1 := items[0]
				lem2 := items[1]
				_, exists := this.lemmaPrefs[lem1]
				if ! exists {
					this.lemmaPrefs[lem1] = lem2
				}
				break
			}
			case DICTIONARY_POS_PREF: {
				pos1 := items[0]
				pos2 := items[1]
				_, exists := this.posPrefs[pos1]
				if !exists {
					this.posPrefs[pos1] = pos2
				}
				break
			}
			case DICTIONARY_ENTRIES: {
				if this.morfodb == nil {
					LOG.Panic("No IndexType specified in dictionary file " + dicFile)
				}

				pos := strings.Index(line, " ")
				key := line[0:pos]
				data := line[pos + 1:]

				if key == "" {
					LOG.Panic("Invalid format. Unexpected blank line in " + dicFile)
				}

				lems := list.New()

				if !this.ParseDictEntry(data, lems) {
					LOG.Panic("Invalid pair lemma-tag in dictionary line " + key + " " + data)
				}

				data = this.CompactData(lems)

				this.morfodb.addDatabase(key, data)

				if this.InverseDic {
					for p := lems.Front(); p != nil; p = p.Next() {
						for t := p.Value.(Pair).second.(*list.List).Front(); t != nil; t = t.Next() {
							this.inverdb.addDatabase(p.Value.(Pair).first.(string) + "#" + t.Value.(string), key)
						}
					}
				}
				break
			}
		default:
			break
		}
	}

	LOG.Trace("Analyzer successfully created")
	return &this
}

func (this *Dictionary) less(s1 string, s2 string, pref map[string]string) bool {
	var p string
	var exists bool

	p, exists = pref[s1]
	if exists && p == s2 { return true}
	p, exists = pref[s2]
	if exists && p == s1 { return false}

	return s1 < s2
}

func (this *Dictionary) SortList(ls *list.List, pref map[string]string) {
	tmpLs := make([]string, ls.Len())
	count := 0
	for i := ls.Front(); i != nil; i = i.Next() {
		tmpLs[count] = i.Value.(string)
		count++
	}

	for i := 0; i < len(tmpLs); i++ {
		for j := i + 1; j < len(tmpLs); j++ {
			if this.less(tmpLs[j], tmpLs[i], pref) {
				tmp := tmpLs[j]
				tmpLs[j] = tmpLs[i]
				tmpLs[i] = tmp
			}
		}
	}

	ls = ls.Init()
	for i := 0; i < len(tmpLs); i++ {
		ls.PushBack(tmpLs[i])
	}
}

func (this *Dictionary) ParseDictEntry(data string, lems *list.List) bool {
	aux := make(map[string]*set.Set)
	dataItems := Split(data, " ")
	sl := set.New()

	for i := 0; i < len(dataItems) - 1; i = i + 2{
		lemma := dataItems[i]
		sl.Add(lemma)
		if i == len(dataItems) {
			return false
		}
		tag := dataItems[i + 1]

		l := aux[lemma]
		if l == nil {
			l = set.New()
			aux[lemma] = l
		}
		l.Add(tag)
	}

	ll := list.New()
	for _, l := range sl.List() { ll.PushBack(l.(string))}
	this.SortList(ll, this.lemmaPrefs)

	for k := ll.Front(); k != nil; k = k.Next() {

		l := aux[k.Value.(string)]

		lt := list.New()
		for _, s := range l.List() { lt.PushBack(s.(string))}
		this.SortList(lt, this.posPrefs)
		lems.PushBack(Pair{k.Value.(string), lt})

	}

	return true
}

func (this *Dictionary) CompactData(lems *list.List) string {
	cdata := ""
	for p := lems.Front(); p != nil; p = p.Next() {
		pair := p.Value.(Pair)
		lemma := pair.first.(string)
		values := pair.second.(*list.List)
		cdata += lemma + LEMMA_DIVIDER
		for v := values.Front(); v != nil; v = v.Next() {
			cdata += v.Value.(string) + TAG_DIVIDER
		}
		cdata = strings.Trim(cdata, "|") + " "
	}
	cdata = strings.Trim(cdata, " ")
	return cdata
}

func (this *Dictionary) SearchForm(s string, la *list.List) {
	key := strings.ToLower(s)
	data := this.morfodb.accessDatabase(key)
	if data != "" {
		p := 0
		q := 0
		for p > -1 {
			LOG.Trace("word '" + s + "'. remaining data: [" + data[p:] + "]")
			q = strings.Index(data[p:], LEMMA_DIVIDER)
			lem := data[p:]
			if q > -1 {
				lem = data[p:p+q]
			}
			LOG.Trace("   got lemma " + lem + " p=" + strconv.Itoa(p) + " q=" + strconv.Itoa(q))
			p = p + q + 1
			q = strings.Index(data[p:], LEMMA_DIVIDER)
			tmpString := data[p:]
			if q > -1 {
				tmpString = data[p: p+q]
			}
			tags := Split(tmpString, TAG_DIVIDER)
			for _, tag := range tags {
				LOG.Trace("Adding (" + lem + "," + tag + ") to analysis list")
				a := NewAnalysis(lem, tag)
				la.PushBack(a)
			}

			p = If(q == -1 , -1, p + q + 1).(int)
		}
	}
}

func (this *Dictionary) tagCombination(p *list.Element, last *list.Element) *list.List {
	output := list.New()
	if p == last {
		tmpItems := Split(p.Value.(string), "/")
		for _, tmpItem := range tmpItems {
			output.PushBack(tmpItem)
		}
		return output
	} else {
		tmpItems := Split(p.Value.(string), "/")
		curr := list.New()
		for _, tmpItem := range tmpItems {
			curr.PushBack(tmpItem)
		}
		c := this.tagCombination(p.Next(),last)
		for i := curr.Front(); i != nil; i = i.Next() {
			for j := c.Front(); j != nil; j = j.Next() {
				output.PushBack(i.Value.(string) + "+" + j.Value.(string))
			}
		}

		return output
	}
}

func (this *Dictionary) CheckContracted(form string, lem string, tag string, lw *list.List) bool {
	LOG.Trace("Check contracted word " + form)
	caps := Capitalization(form)
	la := list.New()
	contr := false

	pl := strings.Index(lem, "+")
	pt := strings.Index(tag, "+")

	for pl > -1 && pt > -1 {
		contr = true

		cl := Substr(lem, 0, pl)
		ct := Split(Substr(tag, 0, pt), "/")
		lem = Substr(lem, pl + 1, -1)
		tag = Substr(tag, pt + 1, -1)

		LOG.Trace("Searching contraction component " + cl + "_" + strings.Join(ct, "/"))
		la = la.Init()
		this.SearchForm(cl, la)
		cl = Capitalize(cl, caps, lw.Len() == 0)
		c := NewWordFromLemma(cl)
		for a := la.Front(); a != nil; a = a.Next() {
			for _, t := range ct {
				if strings.Index(a.Value.(*Analysis).getTag(), t) == 0 || t == "*" {
					c.addAnalysis(a.Value.(*Analysis))
					LOG.Trace("   Matching analysis: " + a.Value.(*Analysis).getTag())
				}
			}
		}

		lw.PushBack(c)


		if c.getNAnalysis() == 0 {
			LOG.Panic("Tag not found for contraction component. Check dictionary entries for '" + form + "' and '" + cl + "'")
		}

		pl = strings.Index(lem, "+")
		pt = strings.Index(tag, "+")
	}

	if contr {
		cl := Substr(lem,0,pl)
		ct := Split(Substr(tag, 0, pt), "/")
		lem = Substr(lem, pl + 1, -1)
		tag = Substr(tag, pt + 1, -1)

		LOG.Trace("Searching contraction component... " + cl + "_" + strings.Join(ct, "/"))

		la = list.New()
		this.SearchForm(cl, la)

		if caps == 2 {
			cl = strings.ToUpper(cl)
		} else if (caps == 1 && lw.Len() == 0) {
			cl = strings.Title(cl)
		}
		LOG.Tracef("Found %d analysis", la.Len())
		c := NewWordFromLemma(cl)
		for a := la.Front(); a != nil; a = a.Next() {
			for _, t := range ct {
				if strings.Index(a.Value.(*Analysis).getTag(), t) == 0 || t == "*" {
					c.addAnalysis(a.Value.(*Analysis))
					LOG.Trace("   Matching analysis: " + a.Value.(*Analysis).getTag())
				}
			}
		}

		lw.PushBack(c)
		if c.getNAnalysis() == 0 {
			LOG.Panic("Tag not found for contraction component. Check dictionary entries for '" + form + "' and '" + cl + "'")
		}
	}

	return contr
}

func (this *Dictionary) AnnotateWord(w *Word, lw *list.List, override bool) bool {
	LOG.Trace("Searching in dictionary for word " + w.getForm())
	la := list.New()
	this.SearchForm(w.getForm(), la)
	w.setFoundInDict(la.Len() > 0)
	LOG.Trace("   Found " + strconv.Itoa(la.Len()) + " analysis.")
	for a:= la.Front(); a != nil; a = a.Next() {
		w.addAnalysis(a.Value.(*Analysis))
		LOG.Trace("   added analysis " + a.Value.(*Analysis).getLemma())
	}

	if (this.CompoundAnalysis) {
		//TODO
	}

	contr := false

	if !this.RetokenizeContractions || override {
		newLa := list.New()
		na := &Analysis{}

		for a := w.Front(); a != nil; a = a.Next() {
			tgs := list.New()
			tmpItems := Split(a.Value.(*Analysis).getTag(), "+")
			for _, tmpItem := range tmpItems { tgs.PushBack(tmpItem) }
			tc := this.tagCombination(tgs.Front(),tgs.Back().Prev())

			if tc.Len() > 1 {
				newLa = newLa.Init()
				for tag := tc.Front(); tag != nil; tag = tag.Next() {
					na.init(a.Value.(*Analysis).getLemma(), tag.Value.(string))
					newLa.PushBack(na)
				}

				ta := a
				for t := newLa.Front(); t != nil; t = t.Next() {
					ta = w.InsertAfter(ta, t)
				}
				w.Remove(a)
				a = ta
			}
		}

		for a := w.Front(); a != nil; a = a.Next() {
			lw = lw.Init()
			if this.CheckContracted(w.getForm(), a.Value.(*Analysis).getLemma(), a.Value.(*Analysis).getTag(), lw) {
				a.Value.(*Analysis).setRetokenizable(lw)
			}
		}
	} else {
		ca := w.Front()
		for ca != nil && (strings.Index(ca.Value.(*Analysis).getLemma(), "+") == -1 || strings.Index(ca.Value.(*Analysis).getTag(), "+") == -1) {
			ca = ca.Next()
		}

		if ca != nil && w.getNAnalysis() > 1 {
			LOG.Warn("Contraction " + w.getForm() + " has several analysis in dictionary. All ignored except (" + ca.Value.(*Analysis).getLemma() + "," + ca.Value.(*Analysis).getTag() + "). Set RetokenizeContraction=false to keep all analysis.")
		} else {
			ca = w.Front()
		}
		if ca != nil && this.CheckContracted(w.getForm(), ca.Value.(*Analysis).getLemma(), ca.Value.(*Analysis).getTag(), lw) {
				contr = true
		}
	}

	return contr
}

func (this *Dictionary) Analyze(se *Sentence) {
	contr := false

	for pos := se.Front(); pos != nil; pos = pos.Next() {
		LOG.Tracef("Processing: %s - %d %s", pos.Value.(*Word).getForm(), pos.Value.(*Word).getNAnalysis(), string(pos.Value.(*Word).getTag(0)))
		if pos.Value.(*Word).getNAnalysis()	== 0 || (pos.Value.(*Word).getNAnalysis() > 0 && string(pos.Value.(*Word).getTag(0)[0]) == "Z") {
			LOG.Trace("Annotating word:" + pos.Value.(*Word).getForm())

			lw := list.New()

			if this.AnnotateWord(pos.Value.(*Word), lw, false) {
				st := pos.Value.(*Word).getSpanStart()
				fin := pos.Value.(*Word).getSpanFinish()
				LOG.Trace("Contraction found, replacing... " + pos.Value.(*Word).getForm() + ". span=(" + strconv.Itoa(int(st)) + "," + strconv.Itoa(int(fin)) + ")")

				step := (float64(fin) - float64(st) + 1.0) / float64(lw.Len())
				step = math.Max(1, step)
				ln := math.Max(1, step - 1.0)

				var n int
				var i *list.Element
				n = 1
				for i = lw.Front(); i != nil; i = i.Next() {
					f := If( n == lw.Len(), fin, st + int(ln)).(int)
					i.Value.(*Word).setSpan(st, f)
					i.Value.(*Word).user = pos.Value.(*Word).user

					LOG.Trace("   Inserting " + i.Value.(*Word).getForm() + ". span=(" + strconv.Itoa(int(st)) + "," + strconv.Itoa(int(fin)) + ")")
					pos = se.InsertBefore(i.Value.(*Word), pos)
					pos = pos.Next()
					st = st + int(step)

					contr = true
				}

				LOG.Trace("   Erasing " + pos.Value.(*Word).getForm())
				q := pos
				q = q.Prev()
				se.Remove(pos)
				pos = q
			}
		}
	}

	if contr {
		se.rebuildWordIndex()
	}
}




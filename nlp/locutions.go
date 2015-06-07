package nlp

import (
	"container/list"
	set "gopkg.in/fatih/set.v0"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	LOCUTIONS_ST_P = 1 + iota
	LOCUTIONS_ST_M
	LOCUTIONS_ST_STOP
)

const (
	LOCUTIONS_TK_pref = 1 + iota
	LOCUTIONS_TK_mw
	LOCUTIONS_TK_prefL
	LOCUTIONS_TK_mwL
	LOCUTIONS_TK_prefP
	LOCUTIONS_TK_mwP
	LOCUTIONS_TK_other
)

const (
	LOCUTIONS_TAGSET = 1 + iota
	LOCUTIONS_MULTIWORDS
	LOCUTIONS_ONLYSELECTED
)

type LocutionStatus struct {
	AutomatStatus
	accMW, longestMW *set.Set
	components       []*Word
	overLongest      int
	mwAnalysis       *list.List
	mwAmbiguous      bool
}

func NewLocutionStatus() *LocutionStatus {
	return &LocutionStatus{
		accMW:      set.New(),
		longestMW:  set.New(),
		mwAnalysis: list.New(),
		components: make([]*Word, 0),
	}
}

type Locutions struct {
	Automat
	locut        map[string]string
	prefixes     *set.Set
	Tags         *TagSet
	onlySelected bool
}

func NewLocutions(locFile string) *Locutions {
	this := Locutions{
		locut:    make(map[string]string),
		prefixes: set.New(),
	}

	/*
		cfg := NewConfigFile(false, "##")
		cfg.AddSection("TagSetFile", LOCUTIONS_TAGSET)
		cfg.AddSection("Multiwords", LOCUTIONS_MULTIWORDS)
		cfg.AddSection("OnlySelected", LOCUTIONS_ONLYSELECTED)
	*/
	filestr, err := ioutil.ReadFile(locFile)
	if err != nil {
		LOG.Panic("Error opening file " + locFile)
	}
	lines := strings.Split(string(filestr), "\n")

	for _, line := range lines {
		this.addLocution(line)
	}

	/*
		if !cfg.Open(locFile) {
			CRASH("Error opening file " + locFile, MOD_LOCUTIONS)
		}

		line := ""
		for cfg.GetContentLine(&line) {
			switch cfg.GetSection() {
				case LOCUTIONS_MULTIWORDS: {
					this.addLocution(line)
					break
				}
				case LOCUTIONS_TAGSET: {
					path := locFile[0:strings.LastIndex(locFile, "/")]
					this.Tags = NewTagset(path + "/" + strings.Replace(line, "./", "", -1))
					break
				}
				case LOCUTIONS_ONLYSELECTED: {
					this.onlySelected = (line == "yes" || line == "true")
					break
				}
			default:
				break
			}
		}
	*/

	this.initialState = LOCUTIONS_ST_P
	this.stopState = LOCUTIONS_ST_STOP
	if this.final == nil {
		this.final = set.New()
	}
	this.final.Add(LOCUTIONS_ST_M)
	var s, t int
	for s = 0; s < AUTOMAT_MAX_STATES; s++ {
		for t = 0; t < AUTOMAT_MAX_TOKENS; t++ {
			this.trans[s][t] = LOCUTIONS_ST_STOP
		}
	}

	this.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_pref] = LOCUTIONS_ST_P
	this.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_prefL] = LOCUTIONS_ST_P
	this.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_prefP] = LOCUTIONS_ST_P
	this.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_mw] = LOCUTIONS_ST_M
	this.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_mwL] = LOCUTIONS_ST_M
	this.trans[LOCUTIONS_ST_P][LOCUTIONS_TK_mwP] = LOCUTIONS_ST_M

	this.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_pref] = LOCUTIONS_ST_P
	this.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_prefL] = LOCUTIONS_ST_P
	this.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_prefP] = LOCUTIONS_ST_P
	this.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_mw] = LOCUTIONS_ST_M
	this.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_mwL] = LOCUTIONS_ST_M
	this.trans[LOCUTIONS_ST_M][LOCUTIONS_TK_mwP] = LOCUTIONS_ST_M

	LOG.Trace("analyzer succesfully created")

	return &this
}

func (this *Locutions) BuildMultiword(se *Sentence, start *list.Element, end *list.Element, fs int, built *bool, st *LocutionStatus) *list.Element {
	mw := list.New()
	var form string
	for i := 0; i < st.shiftBegin && start != nil; i++ {
		start = start.Next()
	}
	var i *list.Element
	for i = start; i != end; i = i.Next() {
		mw.PushBack(i.Value.(*Word))
		form += i.Value.(*Word).getForm() + "_"
		LOG.Trace("added last [" + form + "]")
	}

	mw.PushBack(i.Value.(*Word))
	form += i.Value.(*Word).getForm()
	LOG.Trace("added last [" + form + "]")

	w := NewMultiword(form, mw)
	if this.ValidMultiWord(w, st) {
		LOG.Trace("Valid Multiword. Modifying the sentence")
		end = end.Next()
		se.InsertBefore(w, start)
		for i = start; i != end; i = i.Next() {
			i.Value.(*Word).expired = true
			LOG.Trace("Word " + i.Value.(*Word).getForm() + " expired")
		}
		i = end
		LOG.Trace("New word inserted")
		this.SetMultiwordAnalysis(w, fs, st)
		*built = true
	} else {
		LOG.Trace("Multiword found but rejected. Sentence untouched")
		this.ResetActions(st)
		i = start
		*built = false
	}

	return i
}

func (this *Locutions) addLocution(line string) {
	if line == "" {
		return
	}
	var prefix, key, lemma, tag string
	var p int
	items := Split(line, " ")
	key = items[0]

	lemma = items[1]
	tag = items[2]

	data := lemma + " " + tag
	var t [2]string
	i := 0

	for k := 3; k < len(items); k++ {
		t[i] = items[k]
		if i == 1 {
			data += "#" + t[0] + " " + t[1]
			t[0] = ""
			t[1] = ""
		}
		i = 1 - i
	}

	if t[0] == "" {
		t[0] = "I"
	}
	data += "|" + t[0]

	this.locut[key] = data

	prefix = ""
	p = strings.Index(key, "_")
	for p > -1 {
		prefix += key[0 : p+1]
		this.prefixes.Add(prefix)
		key = key[p+1:]
		p = strings.Index(key, "_")
	}
}

func (this *Locutions) setOnlySelected(b bool) {
	this.onlySelected = b
}

func (this *Locutions) check(s string, acc *set.Set, mw *bool, pref *bool, st *LocutionStatus) bool {
	if this.locut[s] != "" {
		acc.Add(s)
		st.longestMW = acc
		st.overLongest = 0
		LOG.Trace("   Added MW:" + s)
		*mw = true
	} else if this.prefixes.Has(s + "_") {
		acc.Add(s)
		LOG.Trace("   Added PRF:" + s)
		*pref = true
	}

	return *mw || *pref
}

func (this *Locutions) ComputeToken(state int, j *list.Element, se *Sentence) int {
	st := se.getProcessingStatus().(*LocutionStatus)
	if st.components == nil {
		st.components = make([]*Word, 0)
	}
	st.components = append(st.components, j.Value.(*Word))
	var form, lem, tag string
	form = j.Value.(*Word).getLCForm()

	token := LOCUTIONS_TK_other

	acc := set.New()
	mw := false
	pref := false

	if j.Value.(*Word).Len() == 0 {
		LOG.Trace("checking (" + form + ")")
		if st.accMW.Size() == 0 {
			this.check(form, acc, &mw, &pref, st)
		} else {
			for _, i := range st.accMW.List() {
				LOG.Trace("   acc_mw: [" + i.(string) + "]")
				this.check(i.(string)+"_"+form, acc, &mw, &pref, st)
			}
		}
	} else {
		first := j.Value.(*Word).Front()

		if this.onlySelected {
			first = j.Value.(*Word).selectedBegin(0).Element
			LOG.Trace("Only selected is set.")
		}
		for a := first; a != nil; a = a.Next() {
			bm := false
			bp := false
			lem = "<" + a.Value.(*Analysis).getLemma() + ">"
			tag = a.Value.(*Analysis).getTag()
			if this.Tags != nil {
				tag = this.Tags.GetShortTag(tag)
			}
			LOG.Trace("checking (" + form + "," + lem + "," + tag + ")")
			if st.accMW.Size() == 0 {
				this.check(form, acc, &bm, &bp, st)
				this.check(lem, acc, &bm, &bp, st)
				if this.check(tag, acc, &bm, &bp, st) {
					j.Value.(*Word).unselectAllAnalysis(0)
					a.Value.(*Analysis).markSelected(0)
				}

				mw = mw || bm
				pref = pref || bp
			} else {
				for _, i := range st.accMW.List() {
					LOG.Trace("   acc_mw: [" + i.(string) + "]")
					this.check(i.(string)+"_"+form, acc, &bm, &bp, st)
					this.check(i.(string)+"_"+lem, acc, &bm, &bp, st)
					if this.check(i.(string)+"_"+tag, acc, &bm, &bp, st) {
						j.Value.(*Word).unselectAllAnalysis(0)
						a.Value.(*Analysis).markSelected(0)
					}
					mw = mw || bm
					pref = pref || bp
				}
			}
		}
	}

	LOG.Trace("  fora :" + If(mw, "MW", "noMW").(string) + "," + If(pref, "PREF", "noPref").(string))
	if mw {
		token = LOCUTIONS_TK_mw
	} else if pref {
		token = LOCUTIONS_TK_pref
	}

	st.overLongest++
	st.accMW = acc

	LOG.Trace("Encoded word: [" + form + "," + lem + "," + tag + "] token=" + strconv.Itoa(token))
	return token
}

func (this *Locutions) ResetActions(st *LocutionStatus) {
	st.longestMW.Clear()
	st.accMW.Clear()
	st.components = make([]*Word, 0)
	st.mwAnalysis = st.mwAnalysis.Init()
}

func (this *Locutions) ValidMultiWord(w *Word, st *LocutionStatus) bool {
	var lemma, tag, check, par string
	var nc int
	la := list.New()
	valid := false
	ambiguous := false

	LOG.Trace(" Form MW: (" + w.getLCForm() + ") " + " comp=" + strconv.Itoa(len(st.components)-st.overLongest))
	LOG.Trace(" longest_mw #candidates: (" + strconv.Itoa(st.longestMW.Size()) + ")")

	for _, m := range st.longestMW.List() {
		form := m.(string)
		if this.locut[form] != "" {
			LOG.Trace("Matched locution: (" + form + ")")

			mwData := this.locut[form]
			p := strings.Index(mwData, "|")
			tags := mwData[0:p]
			ldata := Split(tags, "#")
			amb := mwData[p+1:]
			ambiguous = ambiguous || (amb == "A")

			LOG.Trace("   found entry (" + tags + ")")
			for _, k := range ldata {
				LOG.Trace("   process item (" + k + ")")
				items := Split(k, " ")
				lemma = items[0]
				tag = items[1]
				LOG.Trace("   yields pair (" + lemma + "," + tag + ")")

				p = If(string(lemma[0]) == "$", 0, -1).(int)

				for p > -1 {
					lf := lemma[p+1 : p+2]
					pos, _ := strconv.Atoi(lemma[p+2 : p+3])
					var repl string
					LOG.Trace("n_selected=" + strconv.Itoa(st.components[pos-1].getNSelected(0)))
					if lf == "F" {
						repl = st.components[p-1].getLCForm()
					} else if lf == "L" {
						repl = st.components[p-1].getLemma(0)
					} else {
						LOG.Panic("Invalid lemma in locution entry " + form + " " + lemma + " " + tag)
					}

					lemma = strings.Replace(lemma, lemma[p:p+3], repl, p)
					p = strings.Index(lemma, "$")
				}

				if string(tag[0]) != "$" {
					la.PushBack(NewAnalysis(lemma, tag))
					valid = true
				} else {
					p = If(string(tag[0]) == ":", 0, -1).(int)
					if p > -1 {
						LOG.Panic("Invalid tag in locution entry: " + form + " " + lemma + " " + tag)
					}

					check = tag[p+1:]
					nc, _ = strconv.Atoi(tag[1 : p-1])
					LOG.Trace("Getting tag from word $" + strconv.Itoa(nc) + ", constraint:" + check)

					found := false
					for a := st.components[nc-1].Front(); a != nil; a = a.Next() {
						LOG.Trace("   checking analysis: " + a.Value.(*Analysis).getLemma() + " " + a.Value.(*Analysis).getTag())
						par = a.Value.(*Analysis).getTag()
						if strings.Index(par, check) == 0 {
							found = true
							la.PushBack(NewAnalysis(lemma, par))
						}
					}

					if !found {
						LOG.Trace("Validation failed: Tag " + tag + " nof found in word. Locution entry: " + form + " " + lemma + " " + tag)
					}
					valid = found
				}
			}
		}
	}

	st.mwAnalysis = la
	st.mwAmbiguous = ambiguous
	return valid
}

func (this *Locutions) SetMultiwordAnalysis(i *Word, fstate int, st *LocutionStatus) {
	i.setAnalysis(List2Array(st.mwAnalysis)...)
	i.setAmbiguousMw(st.mwAmbiguous)
	LOG.Trace("Analysis set to: (" + i.getLemma(0) + "," + i.getTag(0) + ") " + If(st.mwAmbiguous, "[A]", "[I]").(string))
}

func (this *Locutions) matching(se *Sentence, i *list.Element) bool {
	var j, sMatch, eMatch *list.Element
	var newstate, state, token, fstate int
	found := false

	LOG.Trace("Checking for mw starting at word '" + i.Value.(*Word).getForm() + "'")

	pst := NewLocutionStatus()
	se.setProcessingStatus(pst)

	state = this.initialState
	fstate = 0
	this.ResetActions(pst)

	pst.shiftBegin = 0

	sMatch = i
	eMatch = nil
	for j = i; state != this.stopState && j != nil; j = j.Next() {
		token = this.ComputeToken(state, j, se)
		newstate = this.trans[state][token]

		//this.StateActions(state, newstate, token,j ,pst)

		state = newstate
		if this.final.Has(state) {
			eMatch = j
			fstate = state
			LOG.Trace("New candidate found")
		}
	}

	LOG.Trace("STOP state reached. Check longest match")
	if eMatch != nil {
		LOG.Trace("Match found")
		i = this.BuildMultiword(se, sMatch, eMatch, fstate, &found, pst)
	}
	se.clearProcessingStatus()

	return found
}

func (this *Locutions) analyze(se *Sentence) {
	var i *list.Element
	found := false

	for i = se.Front(); i != nil; i = i.Next() {
		if !i.Value.(*Word).isLocked() {
			if this.matching(se, i) {
				found = true
				for i.Value.(*Word).expired {
					i = i.Next()
				}
			}
		} else {
			LOG.Trace("Word '" + i.Value.(*Word).getForm() + "' is locked. Skipped.")
		}
	}
	if found {
		se.rebuildWordIndex()
	}
}

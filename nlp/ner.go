package nlp

import (
	"container/list"
	"github.com/fatih/set"
	"regexp"
	"strconv"
	"strings"
)

const (
	NER_NE_TAG = 1 + iota
	NER_TITLE_LIMIT
	NER_AC_TITLE_LIMIT
	NER_SPLIT_MW
)

const NER_TYPE = 1

const (
	NP_NER_TYPE = 1 + iota
	NP_FUNCTION
	NP_SPECIAL
	NP_NAMES
	NP_NE_IGNORE
	NP_REX_NOUNADJ
	NP_REX_CLOSED
	NP_REX_DATNUMPUNT
	NP_AFFIXES
)

const (
	NP_RE_NA  = "^(NC|AQ)"
	NP_RE_DNP = "^[FWZ]"
	NP_RE_CLO = "^[DSC]"
)

const (
	NP_ST_IN = 1 + iota
	NP_ST_NP
	NP_ST_FUN
	NP_ST_PREF
	NP_ST_SUF
	NP_ST_STOP
)

const (
	NP_TK_sUnkUpp = 1 + iota
	NP_TK_sNounUpp
	NP_TK_mUpper
	NP_TK_mFun
	NP_TK_mPref
	NP_TK_mSuf
	NP_TK_other
)

type NERStatus struct {
	AutomatStatus
	initialNoun bool
}

func NewNERStatus() *NERStatus {
	return &NERStatus{}
}

type NERModule struct {
	Automat
	TitleLength        int
	AllCapsTitleLength int
	NETag              string
	splitNPs           bool
}

func NewNERModule(npFile string) *NERModule {
	this := NERModule{
		TitleLength:        0,
		AllCapsTitleLength: 0,
		splitNPs:           false,
	}

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("NE_Tag", NER_NE_TAG)
	cfg.AddSection("TitleLimit", NER_TITLE_LIMIT)
	cfg.AddSection("AllCapsTitleLimit", NER_AC_TITLE_LIMIT)
	cfg.AddSection("SplitMultiwords", NER_SPLIT_MW)
	cfg.skipUnknownSections = true

	if !cfg.Open(npFile) {
		CRASH("Error opening file "+npFile, MOD_NER)
	}

	line := ""

	for cfg.GetContentLine(&line) {
		switch cfg.GetSection() {
		case NER_NE_TAG:
			{
				this.NETag = line
				break
			}
		case NER_TITLE_LIMIT:
			{
				this.TitleLength, _ = strconv.Atoi(line)
				break
			}
		case NER_AC_TITLE_LIMIT:
			{
				this.AllCapsTitleLength, _ = strconv.Atoi(line)
				break
			}
		case NER_SPLIT_MW:
			{
				this.splitNPs = (line == "yes")
				break
			}
		default:
			break
		}
	}

	return &this
}

func (this *NERModule) BuildMultiword(se *Sentence, start *list.Element, end *list.Element, fs int, built *bool, st *NERStatus) *list.Element {
	var i *list.Element
	mw := list.New()
	var form string

	LOG.Trace("  Building multiword")
	for i = start; i != end; i = i.Next() {
		mw.PushBack(i.Value.(*Word))
		form += i.Value.(*Word).getForm() + "_"
		LOG.Trace("   added next [" + form + "]")
	}

	mw.PushBack(i.Value.(*Word))
	form += i.Value.(*Word).getForm()
	LOG.Trace("   added next [" + form + "]")

	w := NewMultiword(form, mw)
	end1 := end
	end1 = end1.Next()

	if this.ValidMultiWord(w, st) {
		if this.splitNPs && start != end {
			LOG.Trace("Valid Multiword. Split NP is on: keeping separate words")
			for j := start; j != nil && j != end1; j = j.Next() {
				if strings.Title(j.Value.(*Word).getForm()) == j.Value.(*Word).getForm() {
					LOG.Trace("   Splitting. Set " + this.NETag + " for " + j.Value.(*Word).getForm())
					j.Value.(*Word).setAnalysis(NewAnalysis(j.Value.(*Word).getLCForm(), this.NETag))
					j.Value.(*Word).setFoundInDict(true)
				}
			}

			this.ResetActions(st)
			i = end
			*built = true
		} else {
			LOG.Trace("Multiword found, but rejected. Sentence untouched")
			end = end.Next()
			se.InsertBefore(w, start)
			for i = start; i != end; i = i.Next() {
				i.Value.(*Word).expired = true
			}

			LOG.Trace("New word inserted")
			this.SetMultiwordAnalysis(w, fs, st)
			*built = true
		}
	} else {
		LOG.Trace("Multiword found but rejected. Sentence untouched")
		this.ResetActions(st)
		i = start
		*built = false
	}

	return i
}

func (this *NERModule) ValidMultiWord(w *Word, st *NERStatus) bool {
	mw := w.getWordsMw()
	if this.TitleLength > 0 && mw.Len() >= this.TitleLength {
		lw := false
		for p := mw.Front(); p != nil; p = p.Next() {
			lw = HasLowercase(p.Value.(*Word).getForm())
		}
		return lw
	} else {
		return true
	}
}

func (this *NERModule) SetMultiwordAnalysis(w *Word, fstate int, st *NERStatus) {
	LOG.Trace("   adding NP analysis")
	w.addAnalysis(NewAnalysis(w.getLCForm(), this.NETag))
}

type NER struct {
	who *NP
}

func NewNER(npFile string) *NER {
	this := NER{}

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("Type", NER_TYPE)
	cfg.skipUnknownSections = true

	if !cfg.Open(npFile) {
		LOG.Panic("Error opening file " + npFile)
	}

	nerType := ""
	line := ""

loop:
	for cfg.GetContentLine(&line) {
		switch cfg.GetSection() {
		case NER_TYPE:
			{
				nerType = strings.ToLower(line)
				break loop
			}
		default:
			break
		}
	}
	if nerType == "basic" {
		this.who = NewNP(npFile)
	}
	return &this
}

func (this *NERModule) ResetActions(st *NERStatus)                            {}
func (this *NERModule) ComputeToken(p int, i *list.Element, se *Sentence) int { return 0 }

type NP struct {
	*NERModule
	fun            *set.Set
	punct          *set.Set
	names          *set.Set
	ignoreTags     map[string]int
	ignoreWords    map[string]int
	prefixes       *set.Set
	suffixes       *set.Set
	RENounAdj      *regexp.Regexp
	REClosed       *regexp.Regexp
	REDateNumPunct *regexp.Regexp
}

func NewNP(npFile string) *NP {
	this := NP{
		fun:            set.New(set.ThreadSafe).(*set.Set),
		punct:          set.New(set.ThreadSafe).(*set.Set),
		names:          set.New(set.ThreadSafe).(*set.Set),
		ignoreTags:     make(map[string]int),
		ignoreWords:    make(map[string]int),
		prefixes:       set.New(set.ThreadSafe).(*set.Set),
		suffixes:       set.New(set.ThreadSafe).(*set.Set),
		RENounAdj:      regexp.MustCompile(NP_RE_NA),
		REClosed:       regexp.MustCompile(NP_RE_CLO),
		REDateNumPunct: regexp.MustCompile(NP_RE_DNP),
	}
	this.NERModule = NewNERModule(npFile)
	this.final = set.New(set.ThreadSafe).(*set.Set)

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("Type", NP_NER_TYPE)
	cfg.AddSection("FunctionWords", NP_FUNCTION)
	cfg.AddSection("SpecialPunct", NP_SPECIAL)
	cfg.AddSection("Names", NP_NAMES)
	cfg.AddSection("Ignore", NP_NE_IGNORE)
	cfg.AddSection("RE_NounAdj", NP_REX_NOUNADJ)
	cfg.AddSection("RE_Closed", NP_REX_CLOSED)
	cfg.AddSection("RE_DateNumPunct", NP_REX_DATNUMPUNT)
	cfg.AddSection("Affixes", NP_AFFIXES)
	cfg.skipUnknownSections = true

	if !cfg.Open(npFile) {
		CRASH("Error opening file "+npFile, MOD_NER)
	}

	line := ""

	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
		case NP_NER_TYPE:
			{
				if strings.ToLower(line) != "basic" {
					CRASH("Invalid configuration file for 'basic' NER, "+npFile, MOD_NER)
				}
				break
			}

		case NP_FUNCTION:
			{
				this.fun.Add(line)
				break
			}

		case NP_SPECIAL:
			{
				this.punct.Add(line)
				break
			}

		case NP_NAMES:
			{
				this.names.Add(line)
				break
			}

		case NP_NE_IGNORE:
			{
				key := items[0]
				tpe, _ := strconv.Atoi(items[1])
				if IsCapitalized(key) {
					this.ignoreTags[key] = tpe + 1
				} else {
					this.ignoreWords[key] = tpe + 1
				}
				break
			}

		case NP_REX_NOUNADJ:
			{
				this.RENounAdj = regexp.MustCompile(line)
				break
			}

		case NP_REX_CLOSED:
			{
				this.REClosed = regexp.MustCompile(line)
				break
			}

		case NP_REX_DATNUMPUNT:
			{
				this.REDateNumPunct = regexp.MustCompile(line)
				break
			}

		case NP_AFFIXES:
			{
				word := items[0]
				tpe := items[1]
				if tpe == "SUF" {
					this.suffixes.Add(word)
				} else if tpe == "PRE" {
					this.prefixes.Add(word)
				} else {
					WARNING("Ignored affix with unknown type '"+tpe+"' in file", MOD_NER)
				}
				break
			}
		}
	}

	this.initialState = NP_ST_IN
	this.stopState = NP_ST_STOP
	this.final.Add(NP_ST_NP)
	this.final.Add(NP_ST_SUF)

	var s, t int
	for s = 0; s < AUTOMAT_MAX_STATES; s++ {
		for t = 0; t < AUTOMAT_MAX_TOKENS; t++ {
			this.trans[s][t] = NP_ST_STOP
		}
	}

	this.trans[NP_ST_IN][NP_TK_sUnkUpp] = NP_ST_NP
	this.trans[NP_ST_IN][NP_TK_sNounUpp] = NP_ST_NP
	this.trans[NP_ST_IN][NP_TK_mUpper] = NP_ST_NP
	this.trans[NP_ST_IN][NP_TK_mPref] = NP_ST_PREF

	this.trans[NP_ST_PREF][NP_TK_mPref] = NP_ST_PREF
	this.trans[NP_ST_PREF][NP_TK_mUpper] = NP_ST_NP

	this.trans[NP_ST_NP][NP_TK_mUpper] = NP_ST_NP
	this.trans[NP_ST_NP][NP_TK_mFun] = NP_ST_FUN
	this.trans[NP_ST_NP][NP_TK_mSuf] = NP_ST_SUF

	this.trans[NP_ST_FUN][NP_TK_mUpper] = NP_ST_NP
	this.trans[NP_ST_FUN][NP_TK_mFun] = NP_ST_FUN

	this.trans[NP_ST_SUF][NP_TK_mSuf] = NP_ST_SUF

	LOG.Trace("analyzer succesfully created")

	return &this
}

func (this *NP) ComputeToken(state int, j *list.Element, se *Sentence) int {
	var form, formU string
	var token int
	var sbegin bool

	formU = j.Value.(*Word).getForm()
	form = j.Value.(*Word).getLCForm()

	token = NP_TK_other

	if j == se.Front() {
		sbegin = true
	} else {
		ant := j
		ant = ant.Prev()
		sbegin = false

		for a := ant.Value.(*Word).Front(); a != nil && !sbegin; a = a.Next() {
			sbegin = (this.punct.Has(a.Value.(*Analysis).getTag()))
		}
	}

	ignore := 0

	it := 0
	itKey := ""
	iw := this.ignoreWords[form]
	if iw != 0 {
		ignore = iw + 1
	} else {
		found := false
		for an := j.Value.(*Word).Front(); an != nil && !found; an = an.Next() {
			itKey = an.Value.(*Analysis).getTag()
			it = this.ignoreTags[itKey]
			found = it != 0
		}
		if found {
			ignore = it + 1
		}
	}

	if ignore == 2 {
		LOG.Trace("Ignorable word (" + form + If(it != 0, ","+itKey+")", ")").(string) + ". Ignore = 0")
		if state == NP_ST_NP {
			token = NP_TK_mUpper
		} else {
			nxt := j
			nxt = nxt.Next()
			if nxt != nil && IsCapitalized(nxt.Value.(*Word).getForm()) {
				token = If(sbegin, NP_TK_sNounUpp, NP_TK_mUpper).(int)
			}
		}
	} else if ignore == 3 {
		LOG.Trace("Ignorable word (" + form + If(it != 0, ","+itKey+")", ")").(string) + ". Ignore = 1")
	} else if sbegin {
		LOG.Trace("Non-ignorable word, sbegin (" + form + ")")

		if len(formU) > 1 && AllCaps(formU) {
			token = NP_TK_sNounUpp
		} else if !j.Value.(*Word).isLocked() && IsCapitalized(formU) && !this.fun.Has(form) && !j.Value.(*Word).isMultiword() && !j.Value.(*Word).findTagMatch(this.REDateNumPunct) {
			if j.Value.(*Word).getNAnalysis() == 0 {
				token = NP_TK_sUnkUpp
			} else if !j.Value.(*Word).findTagMatch(this.REClosed) && (j.Value.(*Word).findTagMatch(this.RENounAdj) || this.names.Has(form)) {
				token = NP_TK_sNounUpp
			}
		}
	} else if !j.Value.(*Word).isLocked() {
		LOG.Trace("non-ignorable word, non-locked (" + form + ")")
		if IsCapitalized(formU) && !j.Value.(*Word).findTagMatch(this.REDateNumPunct) {
			token = NP_TK_mUpper
		} else if this.fun.Has(form) {
			token = NP_TK_mFun
		} else if this.prefixes.Has(form) {
			token = NP_TK_mPref
		} else if this.suffixes.Has(form) {
			token = NP_TK_mSuf
		}
	}

	LOG.Trace("Next word is: [" + formU + "] token=" + strconv.Itoa(token))
	LOG.Trace("Leaving state " + strconv.Itoa(state) + " with token " + strconv.Itoa(token))

	return token
}

func (this *NP) ResetActions(st *NERStatus) {
	st.initialNoun = false
}

func (this *NP) StateActions(origin int, state int, token int, j *list.Element, st *NERStatus) {
	if state == NP_ST_NP {
		LOG.Trace("Actions for state NP")
		st.initialNoun = (token == NP_TK_sNounUpp)
	}

	LOG.Trace("State actions completed. initialNoun=" + strconv.FormatBool(st.initialNoun))
}

func (this *NP) SetMultiWordAnalysis(i *list.Element, fstate int, st *NERStatus) {
	if st.initialNoun && i.Value.(*Word).getNWordsMw() == 1 {
		LOG.Trace("copying first word analysis list")
		i.Value.(*Word).copyAnalysis(i.Value.(*Word).getWordsMw().Front().Value.(*Word))
	}

	this.NERModule.SetMultiwordAnalysis(i.Value.(*Word), fstate, st)
}

func (this *NP) matching(se *Sentence, i *list.Element) bool {
	var j, sMatch, eMatch *list.Element
	var newstate, state, token, fstate int
	found := false

	LOG.Trace("Checking for mw starting at word '" + i.Value.(*Word).getForm() + "'")

	pst := NewNERStatus()
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

		this.StateActions(state, newstate, token, j, pst)

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

func (this *NP) analyze(se *Sentence) {
	found := false
	for i := se.Front(); i != nil; i = i.Next() {
		if !i.Value.(*Word).isLocked() {
			if this.matching(se, i) {
				found = true
				i = i.Next()
				if i == nil {
					break
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

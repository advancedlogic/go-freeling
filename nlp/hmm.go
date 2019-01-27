package nlp

import (
	"container/list"
	"fmt"
	"github.com/fatih/set"
	"github.com/petar/GoLLRB/llrb"
	"math"
	"strconv"
	"strings"
)

const UNOBS_INITIAL_STATE = "0.x"
const UNOBS_WORD = "<UNOBSERVED_WORD>"

type Bigram struct {
	First  string
	Second string
}

func (this *Bigram) Key() interface{} {
	return this.First + "#" + this.Second
}

type Element struct {
	state *Bigram
	kbest int
	prob  float64
}

func NewElement(state *Bigram, kbest int, prob float64) *Element {
	return &Element{
		state: state,
		kbest: kbest,
		prob:  prob,
	}
}

func (this *Element) Less(i llrb.Item) bool {
	proba := this.prob
	probb := i.(*Element).prob
	return proba < probb
}

type Trellis struct {
	kbest        int
	trl          []Map
	ZERO_logprob float64
	InitState    *Bigram
	EndState     *Bigram
}

const TRELLIS_ZERO_logprob = -math.MaxFloat64

func NewTrellis(T int, kb int) *Trellis {
	TRACE(5, "New trellis T:"+strconv.Itoa(T)+" - kb:"+strconv.Itoa(kb), MOD_HMM)
	trl := make([]Map, T)
	for i, _ := range trl {
		trl[i] = make(Map)
	}
	return &Trellis{
		kbest:        kb,
		trl:          trl,
		ZERO_logprob: -math.MaxFloat64,
		InitState:    &Bigram{"0", "0"},
		EndState:     &Bigram{"ENDSTATE", "ENDSTATE"},
	}
}

func (this *Trellis) insert(t int, s *Bigram, sa *Bigram, kb int, p float64) {
	i, ok := this.trl[t].Get(s)
	if !ok {
		TRACE(4, "    Inserting. Is a new element, init list.", MOD_HMM)
		m := llrb.New()
		m.InsertNoReplace(NewElement(sa, kb, p))
		this.trl[t].Insert(s, m)
	} else {
		TRACE(4, "    Inserting. Not a new element, add. List size="+strconv.Itoa(i.(*llrb.LLRB).Len())+"/"+strconv.Itoa(this.kbest), MOD_HMM)
		j := i.(*llrb.LLRB).Min()
		if i.(*llrb.LLRB).Len() == this.kbest && p < j.(*Element).prob {
			TRACE(4, "    Not worth inserting", MOD_HMM)
			return
		}

		i.(*llrb.LLRB).InsertNoReplace(NewElement(sa, kb, p))

		if i.(*llrb.LLRB).Len() > this.kbest {
			i.(*llrb.LLRB).Delete(i.(*llrb.LLRB).Min())
			TRACE(4, "    list too long. Last erased", MOD_HMM)
		}
	}
	i, ok = this.trl[t].Get(s)
	if ok && i.(*llrb.LLRB).Len() > 0 {
		TRACE(4, "MAX:"+strconv.FormatFloat(i.(*llrb.LLRB).Max().(*Element).prob, 'f', -1, 64)+" MIN:"+strconv.FormatFloat(i.(*llrb.LLRB).Min().(*Element).prob, 'f', -1, 64), MOD_HMM)
	}
}

func (this *Trellis) delta(t int, s *Bigram, k int) float64 {
	if k > this.kbest-1 {
		CRASH("Requested k-best path index is larger than number of stored paths.", MOD_HMM)
	}

	ti, ok := this.trl[t].Get(s)
	var res float64
	if !ok {
		res = this.ZERO_logprob
	} else {
		n := 0
		j := ti.(*llrb.LLRB).Max()
		ti.(*llrb.LLRB).DescendLessOrEqual(ti.(*llrb.LLRB).Max(), func(i llrb.Item) bool {
			if n < k {
				j = i
				n++
				return true
			} else {
				return false
			}
		})
		res = j.(*Element).prob
	}
	return res
}

func (this *Trellis) phi(t int, s *Bigram, k int) *Pair {
	if k > this.kbest {
		panic("Requested k-best path index is larger than number of stored paths")
	}

	n := 0
	tj, _ := this.trl[t].Get(s)
	j := tj.(*llrb.LLRB).Max()
	tj.(*llrb.LLRB).DescendLessOrEqual(tj.(*llrb.LLRB).Max(), func(i llrb.Item) bool {
		if n < k {
			j = i
			n++
			return true
		} else {
			return false
		}
	})

	return &Pair{j.(*Element).state, j.(*Element).kbest}
}

func (this *Trellis) nbest(t int, s *Bigram) int {
	j, ok := this.trl[t].Get(s)
	if ok {
		return j.(*llrb.LLRB).Len()
	}
	return 0
}

type processor struct{}

func (this processor) analyze(ls *list.List) {
	for s := ls.Front(); s != nil; s = s.Next() {
		this.analyzeSentence(s.Value.(string))
	}
}

func (this processor) analyzeSentence(s string) {

}

const FORCE_NONE = 0
const FORCE_TAGGER = 1
const FORCE_RETOK = 1

type POSTAGGER interface {
	annotate(s *Sentence)
}

type POSTagger struct {
	retok bool
	force int
}

func NewPosTagger(r bool, f int) *POSTagger {
	return &POSTagger{
		retok: r,
		force: f,
	}
}

func (this *HMMTagger) Analyze(s *Sentence) {
	if s.Len() == 0 {
		return
	} else {
		w := s.Front().Value.(*Word)

		if w.Len() > 0 && w.Front().Value.(*Analysis).getProb() < 0 {
			CRASH("No lexical probabilities!  Make sure you used the 'probabilities' module before the tagger.", MOD_HMM)
		}

		this.annotate(s)

		if this.force == FORCE_TAGGER {
			this.forceSelect(s)
		}
	}
}

func (this *POSTagger) forceSelect(se *Sentence) {
	for w := se.Front(); w != nil; w = w.Next() {
		i := w.Value.(*Word).selectedBegin(0).Element.Value.(*Analysis)
		w.Value.(*Word).unselectAllAnalysis(0)
		w.Value.(*Word).selectAnalysis(i, 0)
	}
}

const (
	UNIGRAM = 1 + iota
	BIGRAM
	TRIGRAM
	INITIAL
	WORD
	SMOOTHING
	FORBIDDEN
	TAGSET
)

type HMMTagger struct {
	POSTagger
	Tags      *TagSet
	PTag      map[string]float64
	PBg       Map
	PTrg      map[string]float64
	PInitial  Map
	PWord     map[string]float64
	Forbidden map[string][]string

	probInitial    float64
	probUnobserved float64
	pA_cache       map[string]float64
	kbest          int
	c              [3]float64
}

func NewHMMTagger(hmmFile string, rtk bool, force int, kb int) *HMMTagger {

	var prob, coef float64
	var nom1, aux, ftags string
	path := hmmFile[0:strings.LastIndex(hmmFile, "/")]

	this := HMMTagger{
		PTag:      make(map[string]float64),
		PBg:       make(Map),
		PTrg:      make(map[string]float64),
		PInitial:  make(Map),
		PWord:     make(map[string]float64),
		Forbidden: make(map[string][]string),
		pA_cache:  make(map[string]float64),
		kbest:     kb,
	}

	this.POSTagger.force = force

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("Tag", UNIGRAM)
	cfg.AddSection("Bigram", BIGRAM)
	cfg.AddSection("Trigram", TRIGRAM)
	cfg.AddSection("Initial", INITIAL)
	cfg.AddSection("Word", WORD)
	cfg.AddSection("Smoothing", SMOOTHING)
	cfg.AddSection("Forbidden", FORBIDDEN)
	cfg.AddSection("TagsetFile", TAGSET)

	if !cfg.Open(hmmFile) {
		panic("Error opening file " + hmmFile)
	}

	line := ""
	for cfg.GetContentLine(&line) {
		items := strings.Split(line, " ")
		switch cfg.GetSection() {
		case UNIGRAM:
			{
				nom1 = items[0]
				prob, _ = strconv.ParseFloat(items[1], 64)
				this.PTag[nom1] = prob
				break
			}
		case BIGRAM:
			{
				nom1 = items[0]
				prob, _ = strconv.ParseFloat(items[1], 64)
				bg := strings.Split(nom1, ".")
				this.PBg.Insert(&Bigram{bg[0], bg[1]}, prob)
				break
			}
		case TRIGRAM:
			{
				nom1 = items[0]
				prob, _ = strconv.ParseFloat(items[1], 64)
				this.PTrg[nom1] = prob
				break
			}
		case INITIAL:
			{
				nom1 = items[0]
				prob, _ = strconv.ParseFloat(items[1], 64)
				if nom1 == UNOBS_INITIAL_STATE {
					this.probInitial = prob
				} else {
					bg := strings.Split(nom1, ".")
					this.PInitial.Insert(&Bigram{bg[0], bg[1]}, prob)
				}
				break
			}
		case WORD:
			{
				nom1 = items[0]
				prob, _ = strconv.ParseFloat(items[1], 64)
				if nom1 == UNOBS_WORD {
					this.probUnobserved = prob
				} else {
					this.PWord[nom1] = prob
				}
				break
			}
		case SMOOTHING:
			{
				nom1 = items[0]
				coef, _ = strconv.ParseFloat(items[1], 64)
				if nom1 == "c1" {
					this.c[0] = coef
				} else if nom1 == "c2" {
					this.c[1] = coef
				} else if nom1 == "c3" {
					this.c[2] = coef
				}
				break
			}
		case FORBIDDEN:
			{
				if this.Tags == nil {
					CRASH(fmt.Sprintf("<TagsetFile> section should appear before <Forbidden> in file %s\n", hmmFile), MOD_HMM)
				}
				aux = items[2]
				err := false
				l := make([]string, 3)
				TRACE(4, fmt.Sprintf("reading forbidden (%s)\n", aux), MOD_HMM)
				ltg := strings.Split(aux, ".")
				for i := 0; i < 3; i++ {
					TRACE(4, fmt.Sprintf("    ...processing (%s)\n", ltg[i]), MOD_HMM)
					p := strings.Index(ltg[i], "<")
					if p > -1 {
						l[i] = ltg[i][p:]
						ltg[i] = ltg[i][0:p]
					}

					err = ((l[i] != "" || i != 0) && (ltg[i] == "*" || ltg[i] == "0"))
				}

				if err {
					WARNING(fmt.Sprintf("Wrong format for forbidden trigram %s. Ignored.", aux), MOD_HMM)
				} else {
					stg := make([]string, 0)
					for i := 0; i < 3; i++ {
						if ltg[i] == "*" || ltg[i] == "0" {
							stg = append(stg, ltg[i])
							ltg[i] = ""
						} else {
							stg = append(stg, this.Tags.GetShortTag(ltg[i]))
							if stg[i] == ltg[i] {
								ltg[i] = ""
							}
						}
					}

					tr := strings.Join(stg, ".")
					lt := strings.Join(ltg, ".")
					lm := strings.Join(l, ".")

					s := ""
					if lt != ".." || lm != ".." {
						s = lm + "#" + lt
					}

					TRACE(4, fmt.Sprintf("Inserting forbidden (%s,%s)\n", tr, s), MOD_HMM)
					if this.Forbidden[tr] == nil {
						this.Forbidden[tr] = make([]string, 0)
					}
					this.Forbidden[tr] = append(this.Forbidden[tr], s)
				}
				break
			}
		case TAGSET:
			{
				ftags = items[0]
				TRACE(3, "Loading tagset file "+path+"/"+ftags, MOD_HMM)
				this.Tags = NewTagset(path + "/" + strings.Replace(ftags, "./", "", -1))
				break
			}
		default:
			break
		}
	}
	if this.probInitial == -1.0 || this.probUnobserved == -1.0 {
		CRASH("HMM model missing '"+UNOBS_INITIAL_STATE+" and/or "+UNOBS_WORD+"' entries", MOD_HMM)
	}
	TRACE(3, "Analyzer succesfully created", MOD_HMM)
	return &this
}

func (this *HMMTagger) isForbidden(trig string, w *list.Element) bool {
	if len(this.Forbidden) == 0 {
		return false
	}

	return true
}

func (this *HMMTagger) ProbA_log(state_i *Bigram, state_j *Bigram, w *list.Element) float64 {
	forb := false
	var prob float64 = 0.0

	t3 := state_j.Second
	t2t3 := state_j.First + "." + state_j.Second
	t1t2t3 := state_i.First + "." + state_i.Second + "." + t3

	if this.isForbidden("*."+t2t3, w) || this.isForbidden(t1t2t3, w) {
		prob = 0
		forb = true
	} else {
		var d float64 = 0.0
		d = this.pA_cache[t1t2t3]
		if d != 0 {
			TRACE(5, fmt.Sprintf("      cached pa(%s)= %f\n", t1t2t3, d), MOD_HMM)
			return d
		}

		k := this.PTag[t3]
		if k != 0 {
			prob += this.c[0] * k
		} else {
			k = this.PTag["x"]
			prob += this.c[0] * k
		}

		b, ok := this.PBg.Get(state_j)
		if ok {
			prob += this.c[1] * b.(float64)
		}

		k = this.PTrg[t1t2t3]
		if k != 0 {
			prob += this.c[2] * k
		}
	}

	prob = math.Log(prob)
	if !forb {
		this.pA_cache[t1t2t3] = prob
	}

	return prob
}

func (this *HMMTagger) ProbB_log(state_i *Bigram, obs *Word) float64 {
	var pb_log, plog_word_tag, plog_word, plog_st float64
	var tag2 string

	tag2 = state_i.Second

	k := this.PWord[obs.getLCForm()]
	if k == 0 {
		plog_word = this.probUnobserved
	} else {
		plog_word = k
	}

	TRACE(5, "Probability for word "+obs.getLCForm()+" = "+strconv.FormatFloat(plog_word, 'f', -1, 64), MOD_HMM)
	k = this.PTag[tag2]

	if k == 0 {
		k = this.PTag["x"]
	}

	plog_st = math.Log(k)
	TRACE(5, "Probability for tag "+tag2+"/x = "+strconv.FormatFloat(plog_st, 'f', -1, 64), MOD_HMM)
	var pa float64 = 0
	for a := obs.Front(); a != nil; a = a.Next() {
		if this.Tags.GetShortTag(a.Value.(*Analysis).getTag()) == tag2 {
			pa += a.Value.(*Analysis).getProb()
		}
	}

	plog_word_tag = math.Log(pa)
	TRACE(5, "Probability word/state "+obs.getLCForm()+"/"+tag2+" = "+strconv.FormatFloat(plog_word_tag, 'f', -1, 64), MOD_HMM)
	pb_log = plog_word_tag + plog_word - plog_st

	TRACE(5, fmt.Sprintf("     plog_word_tag=%f", plog_word_tag), MOD_HMM)
	TRACE(5, fmt.Sprintf("     plog_word=%f", plog_word), MOD_HMM)
	TRACE(5, fmt.Sprintf("     plog_st=%f", plog_st), MOD_HMM)
	TRACE(5, fmt.Sprintf("     pb=%f", pb_log), MOD_HMM)

	return pb_log
}

func (this *HMMTagger) ProbPi_log(state_i *Bigram) float64 {
	var ppi_log float64
	k, ok := this.PInitial.Get(state_i)
	if ok {
		ppi_log = k.(float64)
		TRACE(6, "Initial log_probability for state_i ["+state_i.First+","+state_i.Second+"] = "+strconv.FormatFloat(ppi_log, 'f', -1, 64), MOD_HMM)
	} else if state_i.First == "0" {
		ppi_log = this.probInitial
	} else {
		ppi_log = TRELLIS_ZERO_logprob
	}
	return ppi_log
}

func (this *HMMTagger) SequenceProb_log(se Sentence, k int) float64 {
	var p float64 = 0
	var tag, nexttag string
	var st, nextst *Bigram

	w := se.Front()

	tag = this.Tags.GetShortTag(w.Value.(*Word).getTag(k))
	st = &Bigram{"0", tag}
	p = this.ProbPi_log(st)
	p += this.ProbB_log(st, w.Value.(*Word))
	w = w.Next()

	for w != nil {
		nexttag = this.Tags.GetShortTag(w.Value.(*Word).getTag(k))
		nextst = &Bigram{tag, nexttag}
		p += this.ProbA_log(st, nextst, w)
		p += this.ProbB_log(nextst, w.Value.(*Word))
		tag = nexttag
		st = nextst
		w = w.Next()
	}

	return p
}

func (this *HMMTagger) annotate(se *Sentence) {
	var lemm *list.List
	var emms *list.Element
	var emmsant *list.Element

	var w *list.Element
	var ka *list.Element
	var max float64 = 0
	var aux float64 = 0
	var emm float64 = 0
	var pi float64 = 0
	//kd := make(map[string]float64)
	//ks := make(map[string]string)
	var tag string
	var t int

	tr := NewTrellis(se.Len()+1, this.kbest)

	lemm = this.FindStates(se)
	w = se.Front()

	emms = lemm.Front()
	for i, _ := range emms.Value.(*set.Set).List() {
		tk := emms.Value.(*set.Set).List()[i]
		k := tk.(*Bigram)
		pi = this.ProbPi_log(k)
		emm = this.ProbB_log(k, w.Value.(*Word))
		aux = pi + emm
		tr.insert(0, k, tr.InitState, 0, aux)
	}

	t = 1
	emmsant = lemm.Front()
	emms = lemm.Front().Next()

	for w = se.Front().Next(); w != nil; w = w.Next() {
		for _, k := range emms.Value.(*set.Set).List() {
			emm = this.ProbB_log(k.(*Bigram), w.Value.(*Word))
			for _, kant := range emmsant.Value.(*set.Set).List() {
				if kant.(*Bigram).Second == k.(*Bigram).First {
					for kb := 0; kb < tr.nbest(t-1, kant.(*Bigram)); kb++ {
						pant := tr.delta(t-1, kant.(*Bigram), kb)
						ptrans := this.ProbA_log(kant.(*Bigram), k.(*Bigram), w)
						aux = pant + ptrans + emm
						tr.insert(t, k.(*Bigram), kant.(*Bigram), kb, aux)
					}
				}
			}
		}

		t++
		emmsant = emms
		emms = emms.Next()
	}

	max = TRELLIS_ZERO_logprob
	w = se.Back()
	emms = lemm.Back()

	for _, k := range emms.Value.(*set.Set).List() {
		for kb := 0; kb < tr.nbest(se.Len()-1, k.(*Bigram)); kb++ {
			aux = tr.delta(se.Len()-1, k.(*Bigram), kb)
			tr.insert(se.Len(), tr.EndState, k.(*Bigram), kb, aux)
		}
	}

	for w := se.Front(); w != nil; w = w.Next() {
		w.Value.(*Word).unselectAllAnalysis(0)
	}

	for bp := 0; bp < tr.nbest(se.Len(), tr.EndState); bp++ {
		back := tr.phi(se.Len(), tr.EndState, bp)
		st := back.first.(*Bigram)
		kb := back.second.(int)

		tag = st.Second
		w = se.Back()
		for t := se.Len() - 1; t >= 0; t-- {
			bestk := list.New()
			max = 0.0

			for ka = w.Value.(*Word).Front(); ka != nil; ka = ka.Next() {
				if this.Tags.GetShortTag(ka.Value.(*Analysis).getTag()) == tag {
					if ka.Value.(*Analysis).getProb() > max {
						max = ka.Value.(*Analysis).getProb()
						bestk = bestk.Init()
						bestk.PushBack(ka.Value.(*Analysis))
					} else if ka.Value.(*Analysis).getProb() == max {
						bestk.PushBack(ka.Value.(*Analysis))
					}

				}
			}

			for k := bestk.Front(); k != nil; k = k.Next() {
				w.Value.(*Word).selectAnalysis(k.Value.(*Analysis), bp)
			}

			if t > 0 {
				back = tr.phi(t, st, kb)
				st = back.first.(*Bigram)
				kb = back.second.(int)

				tag = st.Second
				w = w.Prev()
			}
		}
	}

	TRACE(3, "sentence analyzed", MOD_HMM)
}

type emission_states struct {
	*set.Set
}

func (this *HMMTagger) FindStates(sent *Sentence) *list.List {
	st := set.New(set.ThreadSafe).(*set.Set)
	ls := list.New()
	w2 := sent.Front()
	TRACE(3, "obtaining the states that may have emmited the initial word: "+w2.Value.(*Word).getForm(), MOD_HMM)
	for a2 := w2.Value.(*Word).selectedBegin(0).Element; a2 != nil; a2 = a2.Next() {
		st.Add(&Bigram{"0", this.Tags.GetShortTag(a2.Value.(*Analysis).getTag())})
	}
	ls.PushBack(st)

	for w1, w2 := w2, w2.Next(); w1 != nil && w2 != nil; w1, w2 = w2, w2.Next() {
		TRACE(3, "obtaining the states that may have emmited the word: "+w2.Value.(*Word).getForm(), MOD_HMM)
		st := set.New(set.ThreadSafe).(*set.Set)
		for a1 := w1.Value.(*Word).selectedBegin(0).Element; a1 != nil; a1 = a1.Next() {
			for a2 := w2.Value.(*Word).selectedBegin(0).Element; a2 != nil; a2 = a2.Next() {
				st.Add(&Bigram{this.Tags.GetShortTag(a1.Value.(*Analysis).getTag()), this.Tags.GetShortTag(a2.Value.(*Analysis).getTag())})
			}
		}

		ls.PushBack(st)
	}

	return ls
}

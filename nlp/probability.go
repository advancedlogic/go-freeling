package nlp

import (
	"container/list"
	set "gopkg.in/fatih/set.v0"
	"strconv"
	"strings"
)

const (
	PROBABILITY_SINGLE_TAG = 1 + iota
	PROBABILITY_CLASS_TAG
	PROBABILITY_FORM_TAG
	PROBABILITY_UNKNOWN
	PROBABILITY_THEETA
	PROBABILITY_SUFFIXES
	PROBABILITY_SUFF_BIASS
	PROBABILITY_LAMBDA_LEX
	PROBABILITY_LAMBDA_CLASS
	PROBABILITY_TAGSET
)

type Probability struct {
	ProbabilityThreshold  float64
	Tags                  *TagSet
	BiassSuffixes         float64
	LidstoneLambdaLexical float64
	LidstoneLambdaClass   float64
	activateGuesser       bool
	singleTags            map[string]float64
	classTags             map[string]map[string]float64
	lexicalTags           map[string]map[string]float64
	unkTags               map[string]float64
	unkSuffS              map[string]map[string]float64
	theeta                float64
	longSuff              int
}

func NewProbability(probFile string, Threashold float64) *Probability {
	this := Probability{
		singleTags:  make(map[string]float64),
		classTags:   make(map[string]map[string]float64),
		lexicalTags: make(map[string]map[string]float64),
		unkTags:     make(map[string]float64),
		unkSuffS:    make(map[string]map[string]float64),
	}

	line := ""
	var key, frq, tag, ftags string
	var probab, sumUnk, sumSing, count float64
	var tmpMap map[string]float64

	this.activateGuesser = true
	this.ProbabilityThreshold = Threashold
	this.BiassSuffixes = 0.3
	this.LidstoneLambdaLexical = 0.1
	this.LidstoneLambdaClass = 1.0

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("SingleTagFreq", PROBABILITY_SINGLE_TAG)
	cfg.AddSection("ClassTagFreq", PROBABILITY_CLASS_TAG)
	cfg.AddSection("FormTagFreq", PROBABILITY_FORM_TAG)
	cfg.AddSection("UnknownTags", PROBABILITY_UNKNOWN)
	cfg.AddSection("Theeta", PROBABILITY_THEETA)
	cfg.AddSection("Suffixes", PROBABILITY_SUFFIXES)
	cfg.AddSection("BiassSuffixes", PROBABILITY_SUFF_BIASS)
	cfg.AddSection("LidstoneLambdaLexical", PROBABILITY_LAMBDA_LEX)
	cfg.AddSection("LidstoneLambdaClass", PROBABILITY_LAMBDA_CLASS)
	cfg.AddSection("TagsetFile", PROBABILITY_TAGSET)

	if !cfg.Open(probFile) {
		CRASH("Error opening file "+probFile, MOD_PROBABILITY)
	}

	sumUnk = 0
	sumSing = 0
	this.longSuff = 0

	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
		case PROBABILITY_SINGLE_TAG:
			{
				key = items[0]
				frq = items[1]
				probab, _ = strconv.ParseFloat(frq, 64)
				this.singleTags[key] = probab
				sumSing += probab
				break
			}
		case PROBABILITY_CLASS_TAG:
			{
				tmpMap = make(map[string]float64)
				key = items[0]
				for i := 1; i < len(items)-1; i = i + 2 {
					tag = items[i]
					frq = items[i+1]
					probab, _ = strconv.ParseFloat(frq, 64)
					tmpMap[tag] = probab
				}
				this.classTags[key] = tmpMap
				break
			}
		case PROBABILITY_FORM_TAG:
			{
				tmpMap = make(map[string]float64)
				key = items[0]
				//clas = items[1]
				for i := 2; i < len(items)-1; i = i + 2 {
					tag = items[i]
					frq = items[i+1]
					probab, _ = strconv.ParseFloat(frq, 64)
					tmpMap[tag] = probab
				}
				this.lexicalTags[key] = tmpMap
				break
			}
		case PROBABILITY_UNKNOWN:
			{
				key = items[0]
				frq = items[1]
				probab, _ = strconv.ParseFloat(frq, 64)
				this.unkTags[key] = probab
				sumUnk += probab
				break
			}
		case PROBABILITY_THEETA:
			{
				frq = items[0]
				this.theeta, _ = strconv.ParseFloat(frq, 64)
				break
			}
		case PROBABILITY_SUFFIXES:
			{
				tmpMap = make(map[string]float64)
				key = items[0]
				frq1 := items[1]
				this.longSuff = If(len(key) > this.longSuff, len(key), this.longSuff).(int)
				count, _ = strconv.ParseFloat(frq1, 64)
				for i := 2; i < len(items)-1; i = i + 2 {
					tag = items[i]
					frq = items[i+1]
					probab, _ = strconv.ParseFloat(frq, 64)
					probab /= count
					tmpMap[tag] = probab
				}
				this.unkSuffS[key] = tmpMap
				break
			}
		case PROBABILITY_SUFF_BIASS:
			{
				frq = items[0]
				this.BiassSuffixes, _ = strconv.ParseFloat(frq, 64)
				break
			}
		case PROBABILITY_LAMBDA_LEX:
			{
				frq = items[0]
				this.LidstoneLambdaLexical, _ = strconv.ParseFloat(frq, 64)
				break
			}
		case PROBABILITY_LAMBDA_CLASS:
			{
				frq = items[0]
				this.LidstoneLambdaClass, _ = strconv.ParseFloat(frq, 64)
				break
			}
		case PROBABILITY_TAGSET:
			{
				ftags = items[0]
				break
			}
		default:
			break
		}
	}

	for k, _ := range this.unkTags {
		this.unkTags[k] /= sumUnk
	}

	for k, _ := range this.singleTags {
		this.singleTags[k] /= sumSing
	}

	path := probFile[0:strings.LastIndex(probFile, "/")]
	this.Tags = NewTagset(path + "/" + strings.Replace(ftags, "./", "", -1))

	TRACE(3, "analyzer succesfully created", MOD_PROBABILITY)

	return &this
}

func (this *Probability) Analyze(se *Sentence) {
	for pos := se.Front(); pos != nil; pos = pos.Next() {
		this.AnnotateWord(pos.Value.(*Word))
	}
}

func (this *Probability) AnnotateWord(w *Word) {

	var sum float64
	na := w.getNAnalysis()

	TRACE(2, "--Assigning probabilities to: "+w.getForm(), MOD_PROBABILITY)

	if na > 0 && (w.foundInDict() || strings.HasPrefix(w.getTag(0), "F") || strings.HasPrefix(w.getTag(0), "Z") || w.hasRetokenizable()) {
		//TRACE(2, "Form with analysis. Found in dict (" + )
		this.smoothing(w)
		sum = 1
	} else if this.activateGuesser {
		TRACE(2, "Form with NO analysis. Guessing " + w.getForm(), MOD_PROBABILITY)
		var mass float64 = 1.0
		for li := w.Front(); li != nil; li = li.Next() {
			li.Value.(*Analysis).setProb(mass / float64(w.getNAnalysis()))
		}

		sum = this.guesser(w, mass)

		for li := w.Front(); li != nil; li = li.Next() {
			li.Value.(*Analysis).setProb(li.Value.(*Analysis).getProb() / sum)
		}

		na = w.getNAnalysis()
	}

	w.sort()
	w.selectAllAnalysis(0)

	for li := w.Front(); li != nil; li = li.Next() {
		rtk := li.Value.(*Analysis).getRetokenizable()
		for k := rtk.Front(); k != nil; k = k.Next() {
			this.AnnotateWord(k.Value.(*Word))
		}
	}

}

func (this *Probability) smoothing(w *Word) {
	na := w.getNAnalysis()
	if na == 1 {
		TRACE(2, "Inambiguous form, set prob to 1", MOD_PROBABILITY)
		w.Front().Value.(*Analysis).setProb(1)
		return
	}

	TRACE(2, "Form with analysis. Smoothing probabilities", MOD_PROBABILITY)

	var tagShorts = make(map[string]float64)
	for li := w.Front(); li != nil; li = li.Next() {
		tagShorts[this.Tags.GetShortTag(li.Value.(*Analysis).getTag())]++
	}

	var tmpMap map[string]float64
	sum := 0.0
	form := w.getLCForm()
	it := this.lexicalTags[form]
	var usingBackoff bool
	if it != nil {
		TRACE(2, "Form "+form+" lexical probabilities found", MOD_PROBABILITY)
		tmpMap = it
		usingBackoff = false
	} else {

		TRACE(2, "Form "+form+" lexical probabilities not found", MOD_PROBABILITY)
		usingBackoff = true
		c := ""
		cNP := ""
		for k, _ := range tagShorts {
			cNP += "-" + k
			if k != "NP" {
				c += "-" + k
			}
		}

		cNP = cNP[1:]
		if c != "" {
			c = c[1:]
		}

		TRACE(3, "Ambiguity class: ["+cNP+"]. Secondary class: ["+c+"]", MOD_PROBABILITY)

		it = this.classTags[cNP]
		if it != nil {
			TRACE(2, "Ambiguity class "+cNP+" probabilities found.", MOD_PROBABILITY)
			tmpMap = it
		} else if c != "" && c != cNP {
			it = this.classTags[c]
			if it != nil {
				TRACE(2, "Secondary ambiguity class '"+c+"' probabilities found", MOD_PROBABILITY)
				tmpMap = it
			}
		}

		if tmpMap == nil {
			tmpMap = this.singleTags
		}
	}

	for k, v := range tagShorts {
		p := tmpMap[k]
		sum += p * v
	}

	lambda := If(usingBackoff, this.LidstoneLambdaClass, this.LidstoneLambdaLexical).(float64)
	var norm float64 = sum + lambda*float64(na)
	for li := w.Front(); li != nil; li = li.Next() {
		p := tmpMap[this.Tags.GetShortTag(li.Value.(*Analysis).getTag())]
		li.Value.(*Analysis).setProb((p + lambda) / norm)
	}

	if usingBackoff {
		TRACE(2, "Using suffixes to smooth probs", MOD_PROBABILITY)
		norm := 0.0
		p := make([]float64, w.Len())
		i := 0
		for li := w.Front(); li != nil; li = li.Next() {
			p[i] = this.computeProbability(li.Value.(*Analysis).getTag(), li.Value.(*Analysis).getProb(), w.getForm())
			norm += p[i]
			i++
		}

		i = 0
		for li := w.Front(); li != nil; li = li.Next() {
			li.Value.(*Analysis).setProb((1-this.BiassSuffixes)*li.Value.(*Analysis).getProb() + this.BiassSuffixes*p[i]/norm)
			i++
		}
		p = nil
	}

}

func (this *Probability) computeProbability(tag string, prob float64, s string) float64 {
	x := prob
	spos := len(s)
	found := true
	var pt float64
	TRACE(4, " suffixes. Tag "+tag+" initial prob="+strconv.FormatFloat(prob, 'f', -1, 64), MOD_PROBABILITY)
	for spos > 0 && found {
		spos--
		is := this.unkSuffS[s[spos:]]
		found = is != nil
		if found {
			pt = is[tag]
			if pt != 0 {
				TRACE(4, "    found prob for suffix -"+s[spos:], MOD_PROBABILITY)
			} else {
				pt = 0
				TRACE(4, "    NO prob found for suffix -"+s[spos:], MOD_PROBABILITY)
			}

			x = (pt + this.theeta*x) / (1 + this.theeta)
		}
	}

	TRACE(4, "      final prob="+strconv.FormatFloat(x, 'f', -1, 64), MOD_PROBABILITY)
	return x
}

func (this *Probability) guesser(w *Word, mass float64) float64 {
	form := w.getLCForm()

	sum := If(w.getNAnalysis() > 0, mass, 0.0).(float64)
	sum2 := 0.0

	TRACE(2, "Initial sum="+strconv.FormatFloat(sum, 'f', 3, 64), MOD_PROBABILITY)

	stags := set.New()
	for li := w.Front(); li != nil; li = li.Next() {
		stags.Add(this.Tags.GetShortTag(li.Value.(*Analysis).getTag()))
	}

	la := list.New()
	for k, v := range this.unkTags {
		TRACE(2, "   guesser checking tag "+k, MOD_PROBABILITY)
		hasit := stags.Has(this.Tags.GetShortTag(k))

		if !hasit {
			p := this.computeProbability(k, v, form)
			a := NewAnalysis(form, k)
			a.setProb(p)

			if p >= this.ProbabilityThreshold {
				sum += p
				w.addAnalysis(a)
				TRACE(2, "   added. sum is:"+strconv.FormatFloat(sum, 'f', 3, 64), MOD_PROBABILITY)
			} else {
				sum2 += p
				la.PushBack(a)
			}
		}
	}

	if w.getNAnalysis() == 0 {
		w.setAnalysis(List2Array(la)...)
		sum = sum2
	}

	return sum
}

func (this *Probability) setActivateGuesser(b bool) {
	this.activateGuesser = b
}

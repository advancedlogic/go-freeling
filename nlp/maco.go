package nlp

type MacoOptions struct {
	Lang                                                                                                                              string
	LocutionsFile, QuantitiesFile, AffixFile, CompoundFile, DictionaryFile, ProbabilityFile, NPdataFile, PunctuationFile, UserMapFile string
	Decimal, Thousand                                                                                                                 string
	ProbabilityThreshold                                                                                                              float64
	InverseDict, RetokContractions                                                                                                    bool
}

func NewMacoOptions(lang string) *MacoOptions {
	return &MacoOptions{
		Lang:                 lang,
		UserMapFile:          "",
		LocutionsFile:        "",
		QuantitiesFile:       "",
		AffixFile:            "",
		ProbabilityFile:      "",
		DictionaryFile:       "",
		NPdataFile:           "",
		PunctuationFile:      "",
		CompoundFile:         "",
		Decimal:              "",
		Thousand:             "",
		ProbabilityThreshold: 0.001,
		InverseDict:          false,
		RetokContractions:    true,
	}
}

func (this *MacoOptions) SetDataFiles(usr, pun, dic, aff, comp, loc, nps, qty, prb string) {
	this.UserMapFile = usr
	this.LocutionsFile = loc
	this.QuantitiesFile = qty
	this.AffixFile = aff
	this.ProbabilityFile = prb
	this.DictionaryFile = dic
	this.NPdataFile = nps
	this.PunctuationFile = pun
	this.CompoundFile = comp
}

func (this *MacoOptions) SetNumericalPoint(dec string, tho string) {
	this.Decimal = dec
	this.Thousand = tho
}

func (this *MacoOptions) SetThreshold(t float64) {
	this.ProbabilityThreshold = t
}

func (this *MacoOptions) SetInverseDict(b bool) {
	this.InverseDict = b
}

func (this *MacoOptions) SetRetokContractions(b bool) {
	this.RetokContractions = b
}

type Maco struct {
	MultiwordsDetection, NumbersDetection, PunctuationDetection, DatesDetection, QuantitiesDetection, DictionarySearch, ProbabilityAssignment, UserMap, NERecognition bool
	loc                                                                                                                                                               *Locutions
	dic                                                                                                                                                               *Dictionary
	prob                                                                                                                                                              *Probability
	punct                                                                                                                                                             *Punts
	npm                                                                                                                                                               *NER
	/*
		numb *Numbers
		dates *Dates
		quant *Quantities

		user *regexp.Regexp
	*/
}

func NewMaco(opts *MacoOptions) *Maco {
	this := Maco{
		MultiwordsDetection:   false,
		NumbersDetection:      false,
		PunctuationDetection:  false,
		DatesDetection:        false,
		QuantitiesDetection:   false,
		DictionarySearch:      false,
		ProbabilityAssignment: false,
		UserMap:               false,
		NERecognition:         false,
	}

	if opts.PunctuationFile != "" {
		this.punct = NewPunts(opts.PunctuationFile)
		this.PunctuationDetection = true
	}

	if opts.DictionaryFile != "" {
		this.dic = NewDictionary(opts.Lang, opts.DictionaryFile, opts.AffixFile, opts.CompoundFile, opts.InverseDict, opts.RetokContractions)
		this.DictionarySearch = true
	}

	if opts.LocutionsFile != "" {
		this.loc = NewLocutions(opts.LocutionsFile)
		this.MultiwordsDetection = true
	}

	if opts.NPdataFile != "" {
		this.npm = NewNER(opts.NPdataFile)
		this.NERecognition = true
	}

	if opts.ProbabilityFile != "" {
		this.prob = NewProbability(opts.ProbabilityFile, opts.ProbabilityThreshold)
		this.ProbabilityAssignment = true
	}

	return &this
}

func (this *Maco) Analyze(s *Sentence) {
	if this.PunctuationDetection && this.punct != nil {
		this.punct.analyze(s)
	}

	if this.DictionarySearch && this.dic != nil {
		this.dic.Analyze(s)
	}

	if this.MultiwordsDetection && this.loc != nil {
		this.loc.analyze(s)
	}

	if this.NERecognition && this.npm != nil {
		this.npm.who.analyze(s)
	}

	if this.ProbabilityAssignment && this.prob != nil {
		this.prob.Analyze(s)
	}
}

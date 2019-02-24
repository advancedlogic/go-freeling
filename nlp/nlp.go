package nlp

import (
	"container/list"
	"os"
	"strings"

	"github.com/fatih/set"
	"github.com/kdar/factorlog"

	"github.com/drankou/go-freeling/models"
	"github.com/drankou/go-freeling/wordnet"
)

var LOG *factorlog.FactorLog

const (
	PANIC   = factorlog.PANIC
	FATAL   = factorlog.FATAL
	ERROR   = factorlog.ERROR
	WARN    = factorlog.WARN
	DEBUG   = factorlog.DEBUG
	INFO    = factorlog.INFO
	VERBOSE = factorlog.TRACE

	TAG_NP = "NP"
)

func init() {
	frmt := `%{Color "red" "ERROR"}%{Color "yellow" "WARN"}%{Color "green" "INFO"}%{Color "cyan" "DEBUG"}%{Color "blue" "TRACE"}[%{Date} %{Time}] [%{SEVERITY}:%{File}:%{Line}] %{Message}%{Color "reset"}`
	LOG = factorlog.New(os.Stdout, factorlog.NewStdFormatter(frmt))
	LOG.SetMinMaxSeverity(factorlog.PANIC, factorlog.TRACE)
}

type NLPOptions struct {
	Severity          factorlog.Severity
	DataPath          string
	Lang              string
	TokenizerFile     string
	SplitterFile      string
	MorfoOptions      *MacoOptions
	TaggerFile        string
	ShallowParserFile string
	SenseFile         string
	UKBFile           string
	DisambiguatorFile string
	Status            func()
}

func NewNLPOptions(dataPath string, lang string, f func()) *NLPOptions {
	return &NLPOptions{
		DataPath: dataPath,
		Lang:     lang,
		Status:   f,
	}
}

type NLPEngine struct {
	options       *NLPOptions
	tokenizer     *Tokenizer
	splitter      *Splitter
	morfo         *Maco
	tagger        *HMMTagger
	grammar       *Grammar
	shallowParser *ChartParser
	sense         *Senses
	dsb           *UKB
	disambiguator *Disambiguator
	filter        *set.Set
	mitie         *MITIE
	WordNet       *wordnet.WN
}

func NewNLPEngine(options *NLPOptions) *NLPEngine {
	this := NLPEngine{
		options: options,
	}

	LOG.SetMinMaxSeverity(factorlog.PANIC, options.Severity)

	if options.TokenizerFile != "" {
		this.tokenizer = NewTokenizer(options.DataPath + "/" + options.Lang + "/" + options.TokenizerFile)
		this.options.Status()
	}

	if options.SplitterFile != "" {
		this.splitter = NewSplitter(options.DataPath + "/" + options.Lang + "/" + options.SplitterFile)
		this.options.Status()
	}

	if options.MorfoOptions != nil {
		this.morfo = NewMaco(options.MorfoOptions)
		this.options.Status()
	}

	if options.SenseFile != "" {
		this.sense = NewSenses(options.DataPath + "/" + options.Lang + "/" + options.SenseFile)
		this.options.Status()
	}

	if options.TaggerFile != "" {
		this.tagger = NewHMMTagger(options.DataPath+"/"+options.Lang+"/"+options.TaggerFile, true, FORCE_TAGGER, 1)
		this.options.Status()
	}

	if options.ShallowParserFile != "" {
		this.grammar = NewGrammar(options.DataPath + "/" + options.Lang + "/" + options.ShallowParserFile)
		this.shallowParser = NewChartParser(this.grammar)
		this.options.Status()
	}

	if options.UKBFile != "" {
		this.dsb = NewUKB(options.DataPath + "/" + options.Lang + "/" + options.UKBFile)
		this.options.Status()
	}

	if options.DisambiguatorFile != "" {
		if strings.HasPrefix(options.DisambiguatorFile, "common") {
			this.disambiguator = NewDisambiguator(options.DataPath + "/" + options.DisambiguatorFile)
		} else {
			this.disambiguator = NewDisambiguator(options.DataPath + "/" + options.Lang + "/" + options.DisambiguatorFile)
		}
		this.options.Status()
	}

	this.mitie = NewMITIE(options.DataPath + "/" + options.Lang + "/mitie/ner_model.dat")
	this.options.Status()
	return &this
}

func (this *NLPEngine) Workflow(document *models.DocumentEntity, output chan *models.DocumentEntity) {
	defer func() {
		if r := recover(); r != nil {
			err, _ := r.(error)
			if err != nil {
				output <- nil //err.Error()
			} else {
				output <- nil
			}
		}
	}()

	document.Init()
	tokens := list.New()
	url := document.Url
	content := document.Content

	if url != "" && content == "" {
		crawler := NewDefaultCrawler()
		article := crawler.Analyze(url)
		document.Title = article.Title
		document.Description = article.MetaDescription
		document.Keywords = article.MetaKeywords
		document.TopImage = article.TopImage
		document.Content = article.CleanedText
	}

	body := StringsAppend(document.Title, document.Description, document.Keywords, document.Content)

	if this.tokenizer != nil {
		this.tokenizer.Tokenize(body, 0, tokens)
	}

	sentences := list.New()

	if this.splitter != nil {
		sid := this.splitter.OpenSession()
		this.splitter.Split(sid, tokens, true, sentences)
		this.splitter.CloseSession(sid)
	}

	for ss := sentences.Front(); ss != nil; ss = ss.Next() {
		s := ss.Value.(*Sentence)
		if this.morfo != nil {
			this.morfo.Analyze(s)
		}
		if this.sense != nil {
			this.sense.Analyze(s)
		}
		if this.tagger != nil {
			this.tagger.Analyze(s)
		}
		if this.shallowParser != nil {
			this.shallowParser.Analyze(s)
		}
	}

	if this.dsb != nil {
		this.dsb.Analyze(sentences)
	}

	entities := make(map[string]int64)

	for ss := sentences.Front(); ss != nil; ss = ss.Next() {
		se := models.NewSentenceEntity()
		body := ""
		s := ss.Value.(*Sentence)
		for ww := s.Front(); ww != nil; ww = ww.Next() {
			w := ww.Value.(*Word)
			a := w.Front().Value.(*Analysis)

			base := w.getForm()
			lemma := a.getLemma()
			pos := a.getTag()
			props := a.getProb()
			annotation := this.WordNet.Annotate(base, pos)

			te := models.NewTokenEntity(base, lemma, pos, props, annotation)
			if pos == TAG_NP {
				entities[base]++
			}
			body += base + " "
			se.AddTokenEntity(te)
		}
		body = strings.Trim(body, " ")
		se.SetBody(body)
		se.SetSentence(s)

		document.AddSentenceEntity(se)
	}

	tempEntities := set.New(set.ThreadSafe).(*set.Set)

	mitieEntities := this.mitie.Process(body)
	for e := mitieEntities.Front(); e != nil; e = e.Next() {
		entity := e.Value.(*models.Entity)
		tempEntities.Add(entity.GetValue())
	}

	for name, frequency := range entities {
		name = strings.Replace(name, "_", " ", -1)
		if !tempEntities.Has(name) {
			document.AddUnknownEntity(name, frequency)
		}
	}

	document.Entities = mitieEntities
	output <- document
}

func (this *NLPEngine) PrintList(document *models.DocumentEntity) {
	ls := document.Sentences()
	for l := ls.Front(); l != nil; l = l.Next() {
		for w := l.Value.(*Sentence).Front(); w != nil; w = w.Next() {
			item := w.Value.(*Word).getForm() + ":"
			for a := w.Value.(*Word).Front(); a != nil; a = a.Next() {
				if a.Value.(*Analysis).isSelected(0) {
					item += a.Value.(*Analysis).getTag()
				}
			}
			println(item)
		}
	}
}

func (this *NLPEngine) PrintTree(document *models.DocumentEntity) {
	ls := document.Sentences()
	for l := ls.Front(); l != nil; l = l.Next() {
		tr := l.Value.(*models.SentenceEntity).GetSentence().(*Sentence).pts[0]
		output := new(Output)
		out := ""

		output.PrintTree(&out, tr.begin(), 0)

		LOG.Trace(out)
		println(out)
	}
}

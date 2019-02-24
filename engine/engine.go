package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/cheggaaa/pb"

	"github.com/drankou/go-freeling/nlp"
	. "github.com/drankou/go-freeling/terminal"
	"github.com/drankou/go-freeling/wordnet"
)

type Engine struct {
	semaphore *sync.Mutex
	NLP       *nlp.NLPEngine
	Ready     bool
}

func NewEngine() *Engine {
	return &Engine{
		semaphore: new(sync.Mutex),
		Ready:     false,
	}
}

var path = "./"
var lang = "en"

func (e *Engine) InitNLP() {
	e.semaphore.Lock()
	defer e.semaphore.Unlock()
	if e.Ready {
		return
	}
	Infoln("Init Natural Language Processing Engine")
	initialized := false
	count := 80
	bar := pb.StartNew(count)
	bar.ShowPercent = true
	bar.ShowCounters = false

	inc := func() {
		for i := 0; i < 10; i++ {
			bar.Increment()
		}
	}

	start := time.Now().UnixNano()
	nlpOptions := nlp.NewNLPOptions(path+"data/", lang, inc)
	nlpOptions.Severity = nlp.ERROR
	nlpOptions.TokenizerFile = "tokenizer.dat"
	nlpOptions.SplitterFile = "splitter.dat"
	nlpOptions.TaggerFile = "tagger.dat"
	nlpOptions.ShallowParserFile = "chunker/grammar-chunk.dat"
	nlpOptions.SenseFile = "senses.dat"
	nlpOptions.UKBFile = "" //"ukb.dat"
	nlpOptions.DisambiguatorFile = "common/knowledge.dat"

	macoOptions := nlp.NewMacoOptions(lang)
	macoOptions.SetDataFiles("", path+"data/common/punct.dat", path+"data/"+lang+"/dicc.src", "", "", path+"data/"+lang+"/locucions-extended.dat", path+"data/"+lang+"/np.dat", "", path+"data/"+lang+"/probabilitats.dat")

	nlpOptions.MorfoOptions = macoOptions

	nlpEngine := nlp.NewNLPEngine(nlpOptions)

	stop := time.Now().UnixNano()
	delta := (stop - start) / (1000 * 1000)
	initialized = true
	bar.FinishPrint(fmt.Sprintf("Data loaded in %dms", delta))

	wn := wordnet.NewWordNet()
	nlpEngine.WordNet = wn

	e.NLP = nlpEngine
	e.Ready = initialized
}

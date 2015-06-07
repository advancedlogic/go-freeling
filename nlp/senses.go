package nlp

import "container/list"

const SENSES_DUP_ANALYSIS = 1

type Senses struct {
	duplicate bool
	semdb     *SemanticDB
}

func NewSenses(wsdFile string) *Senses {
	this := Senses{
		semdb: NewSemanticDB(wsdFile),
	}

	cfg := NewConfigFile(true, "")
	cfg.AddSection("DuplicateAnalysis", SENSES_DUP_ANALYSIS)

	if !cfg.Open(wsdFile) {
		LOG.Panic("Error opening file " + wsdFile)
	}

	line := ""
	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
		case SENSES_DUP_ANALYSIS:
			{
				key := items[0]
				if key == "yes" {
					this.duplicate = true
				}
				break
			}
		default:
			break
		}
	}

	LOG.Trace("Analyzer succesfully created")

	return &this
}

func (this *Senses) Analyze(sentence *Sentence) {
	for w := sentence.Front(); w != nil; w = w.Next() {
		newla := list.New()
		for a := w.Value.(*Word).Front(); a != nil; a = a.Next() {
			lsen := this.semdb.getWordSenses(w.Value.(*Word).getLCForm(), a.Value.(*Analysis).getLemma(), a.Value.(*Analysis).getTag())

			if lsen.Len() == 0 {
				if this.duplicate {
					newla.PushBack(a.Value.(*Analysis))
				}
			} else {
				if this.duplicate {
					ss := list.New()
					s := lsen.Front()
					newpr := a.Value.(*Analysis).getProb() / float64(lsen.Len())
					for ; s != nil; s = s.Next() {
						newan := NewAnalysisFromAnalysis(a.Value.(*Analysis))
						ss = ss.Init()
						ss.PushBack(s.Value.(string))
						var lsi *list.Element
						lsenNoRanks := list.New()
						for lsi = ss.Front(); lsi != nil; lsi = lsi.Next() {
							lsenNoRanks.PushBack(FloatPair{lsi.Value.(string), 0.0})
						}
						newan.setSenses(lsenNoRanks)
						newan.setProb(newpr)
						newla.PushBack(newan)

						LOG.Trace("  Duplicating analysis for sense " + s.Value.(string))
					}
				} else {
					var lsi *list.Element
					lsenNoRanks := list.New()
					for lsi = lsen.Front(); lsi != nil; lsi = lsi.Next() {
						lsenNoRanks.PushBack(FloatPair{lsi.Value.(string), 0.0})
					}
					a.Value.(*Analysis).setSenses(lsenNoRanks)
				}
			}
		}

		if this.duplicate {
			w.Value.(*Word).setAnalysis(newla)
		}

		LOG.Trace("Sentences annotated by the sense module")
	}
}

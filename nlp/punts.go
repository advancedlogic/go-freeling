package nlp

import (
	"container/list"
	"strings"
	"unicode"
)

const PUNTS_OTHER = "<Other>"

type Punts struct {
	tagOthers string
	*Database
}

func NewPunts(puntFile string) *Punts {
	this := Punts{}
	this.Database = NewDatabaseFromFile(puntFile)
	this.tagOthers = this.accessDatabase(PUNTS_OTHER)

	return &this
}

func (this *Punts) analyze(se *Sentence) {
	var form string
	var i *list.Element

	for i = se.Front(); i != nil; i = i.Next() {
		form = i.Value.(*Word).getForm()
		TRACE(3, "Checking form "+form, MOD_PUNTS)
		data := this.accessDatabase(form)
		if data != "" {
			TRACE(3, "   ["+form+"] found in map: known punctuation", MOD_PUNTS)
			lemma := data[0:strings.Index(data, " ")]
			tag := data[strings.Index(data, " ")+1:]
			i.Value.(*Word).setAnalysis(NewAnalysis(lemma, tag))
			i.Value.(*Word).lockAnalysis()
		} else {
			TRACE(3, "   ["+form+"] not found in map: known punctuation", MOD_PUNTS)
			if !(unicode.IsNumber(rune(form[0])) || unicode.IsLetter(rune(form[0]))) {
				TRACE(3, "   ["+form+"] no alphanumeric char found. tag as "+this.tagOthers, MOD_PUNTS)
				i.Value.(*Word).setAnalysis(NewAnalysis(form, this.tagOthers))
			}
		}

	}
}

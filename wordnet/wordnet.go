package wordnet

import (
	. "github.com/advancedlogic/go-freeling/models"
	. "github.com/advancedlogic/go-freeling/terminal"
	. "github.com/fluhus/gostuff/nlp/wordnet"
)

type WN struct {
	wn *WordNet
}

type partOfSpeech struct {
	short string
	long  string
}

func getPOS(p string) (pos *partOfSpeech) {

	pos = new(partOfSpeech)

	switch p {

	case "JJ", "JJR", "JJS":
		pos.short = "a" //adjective
		pos.long = "adjective"
		break

	case "NNS", "NN", "NNP", "NP00000", "NP", "NP00G00", "NP00O00", "NP00V00", "NP00SP0", "NNPS":
		pos.short = "n" //noun
		pos.long = "noun"
		break

	case "RB", "RBR", "RBS", "WRB":
		pos.short = "r" //adverb
		pos.long = "adverb"
		break

	case "MD", "VBG", "VB", "VBN", "VBD", "VBP", "VBZ":
		pos.short = "v" //verb
		pos.long = "verb"
		break

	default:
		return nil
	}
	return pos
}

func NewWordNet() *WN {
	wn, err := Parse("./data/dict")

	instance := new(WN)

	if err != nil {
		Errorln(err.Error())
		Outputln("There was an error during parsing WordNet database")
	} else {
		instance.wn = wn
	}

	return instance

}

func (this *WN) Annotate(word string, pos string) []*Annotation {
	if this.wn == nil {
		return nil
	}

	wnPOS := getPOS(pos)

	if wnPOS == nil {
		return nil
	}

	result := this.wn.Search(word)[wnPOS.short]

	annotation := []*Annotation{}

	for _, synset := range result {
		annotation = append(annotation, &Annotation{wnPOS.long, synset.Word, synset.Gloss})
	}

	return annotation
}

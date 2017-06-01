package wordnet

import (
	. "../models"
	. "../terminal"
	. "github.com/fluhus/gostuff/nlp/wordnet"
)

type WN struct {
	wn *WordNet
}

func getPOS(p string) (pos string) {

	switch p {

	case "JJ", "JJR", "JJS":
		pos = "a" //adjective
		break

	case "NNS", "NN", "NNP", "NP00000", "NP", "NP00G00", "NP00O00", "NP00V00", "NP00SP0", "NNPS":
		pos = "n" //noun
		break

	case "RB", "RBR", "RBS", "WRB":
		pos = "r" //adverb
		break

	case "MD", "VBG", "VB", "VBN", "VBD", "VBP", "VBZ":
		pos = "v" //verb
		break

	default:
		pos = ""
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
	wnPOS := getPOS(pos)

	if pos == "" {
		return nil
	}

	if this.wn == nil {
		return nil
	}

	result := this.wn.Search(word)[wnPOS]

	annotation := []*Annotation{}

	for _, synset := range result {
		annotation = append(annotation, &Annotation{synset.Pos, synset.Word, synset.Gloss})
	}

	return annotation
}

package nlp

import (
	"container/list"
	"strings"
)

type WDW struct {
	who  string
	did  string
	what string
}

func NewWDW(who string, did string, what string) *WDW {
	return &WDW{
		who:  who,
		did:  did,
		what: what,
	}
}

type WhoDidWhat struct {
	WDWs *list.List
}

func NewWhoDidWhat() *WhoDidWhat {
	return &WhoDidWhat{
		WDWs: list.New(),
	}
}

func (this *WhoDidWhat) Analyze(sentence *Sentence) {
	tr := sentence.pts[0].begin()
	state := 0
	who := ""
	did := ""
	what := ""

	for n := tr.pnode.siblingBegin(); n.pnode != tr.pnode.siblingEnd().pnode; n = n.siblingPlusPlus() {
		label := n.pnode.info.(*Node).getLabel()
		if (label == "sn-chunk" || label == "n-chunk") && state == 0 {
			this.extractLeaves(&who, n)
			state++
		} else if label == "vb-chunk" && state == 1 {
			this.extractLeaves(&did, n)
			state++
		} else if (label == "sn-chunk" || label == "n-chunk") && state == 2 {
			this.extractLeaves(&what, n)
			state++
		} else if state == 3 && label != "vb-chunk" {
			wdw := NewWDW(who, did, what)
			this.WDWs.PushBack(wdw)
			who = ""
			did = ""
			what = ""
			state = 0
		} else {
			who = ""
			did = ""
			what = ""
			state = 0
		}
	}
}

func (this *WhoDidWhat) extractLeaves(output *string, n *ParseTreeIterator) {
	if n.pnode.numChildren() == 0 {
		w := n.pnode.info.(*Node).getWord()
		if w == nil {
			return
		}
		form := w.getForm()
		if strings.Index(*output, form) == -1 {
			*output += w.getForm() + " "
		}
	} else {
		for d := n.pnode.siblingBegin(); d.pnode != n.pnode.siblingEnd().pnode; d = d.siblingPlusPlus() {
			this.extractLeaves(output, d)
		}
	}
}

package nlp

type Output struct{}

func (this Output) outputSense(a *Analysis) string {
	res := ""
	ls := a.getSenses()
	if ls.Len() > 0 {
		for l := ls.Front(); l != nil; l = l.Next() {
			res += " " + l.Value.(*Pair).first.(string) + ""
		}
	}

	return res
}

func (this Output) PrintTree(output *string, n *ParseTreeIterator, depth int) {
	*output += CreateStringWithChar(depth*2, " ")
	if n.pnode.numChildren() == 0 {
		if n.pnode.info.(*Node).isHead() {
			*output += "+"
		}
		w := n.pnode.info.(*Node).getWord()
		if w == nil {
			return
		}
		*output += "(" + w.getForm() + " " + w.getLemma(0) + " " + w.getTag(0) + ")\n"
		//TODO: outputSense
	} else {
		if n.pnode.info.(*Node).isHead() {
			*output += "+"
		}
		*output += n.pnode.info.(*Node).getLabel() + "_[\n"
		for d := n.pnode.siblingBegin(); d.pnode != n.pnode.siblingEnd().pnode; d = d.siblingPlusPlus() {
			this.PrintTree(output, d, depth+1)
		}
		*output += CreateStringWithChar(depth*2, " ") + "]\n"
	}
}

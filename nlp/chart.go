package nlp

import (
	"container/list"
	//"fmt"
	"strconv"
	"strings"
)

type ChartParser struct {
	gram *Grammar
}

func NewChartParser(gram *Grammar) *ChartParser {
	return &ChartParser{
		gram: gram,
	}
}

func (this *ChartParser) getStartSymbol() string {
	return this.gram.getStartSymbol()
}

func (this *ChartParser) Analyze(s *Sentence) {
	//println("Chunking sentence", s.GetBody())
	//var w *list.Element
	//println("CHUNKER ")
	for k := 0; k < s.numKBest(); k++ {
		ch := NewChart(this.gram)
		//println("LOOP ")
		ch.loadSentence(s, k)
		ch.parse()
		//panic("")
		//println("Sentence parsed")
		tr := ch.getTree(ch.getSize()-1, 0, "")
		for w, n := s.Front(), tr.begin(); w != nil && n.pnode != tr.end().pnode; n = n.PlusPlus() {
			LOG.Tracef("Completing tree: %s children %d", n.pnode.info.(*Node).getLabel(), n.pnode.numChildren())
			if n.pnode.numChildren() == 0 {
				n.pnode.info.(*Node).setWord(w.Value.(*Word))
				n.pnode.info.(*Node).setLabel(w.Value.(*Word).getTag(k))
				w = w.Next()
			}
		}

		tr.buildNodeIndex(s.sentID)
		s.setParseTree(tr, k)
	}
}

type Edge struct {
	*Rule
	matched  *list.List
	backpath *list.List
}

func NewEdge() *Edge {
	this := Edge{}
	this.Rule = NewRule()
	return &this
}

func NewEdgeFromString(s string, ls *list.List, pgov int) *Edge {
	this := Edge{
		matched:  list.New(),
		backpath: list.New(),
	}
	this.Rule = NewRuleFromString(s, ls, pgov)
	return &this
}

func NewEdgeFromEdge(edge *Edge) *Edge {
	matched := list.New()
	matched.PushBackList(edge.matched)
	backpath := list.New()
	backpath.PushBackList(edge.backpath)
	this := Edge{
		matched:  matched,
		backpath: backpath,
		Rule:     NewRuleFromRule(edge.Rule),
	}
	return &this
}

func (this *Edge) String() string {
	output := ""
	output += "Head:" + this.getHead() + "\n"
	output += "Governor:" + strconv.Itoa(this.getGovernor()) + "\n"
	output += "Right:\n"
	right := this.getRight()
	for r := right.Front(); r != nil; r = r.Next() {
		output += "    - " + r.Value.(string) + "\n"
	}

	output += "Matched:\n"
	matched := this.getMatched()
	for m := matched.Front(); m != nil; m = m.Next() {
		output += "    - " + m.Value.(string) + "\n"
	}

	output += "Backpath:\n"
	backpath := this.getBackpath()
	for b := backpath.Front(); b != nil; b = b.Next() {
		output += "    - (" + strconv.Itoa(b.Value.(Pair).first.(int)) + "," + strconv.Itoa(b.Value.(Pair).second.(int)) + ")\n"
	}

	return output
}

func (this *Edge) getMatched() *list.List {
	return this.matched
}

func (this *Edge) getBackpath() *list.List {
	return this.backpath
}

func (this *Edge) active() bool {
	return this.right.Len() > 0
}

func (this *Edge) shift(a, b int) {
	//println("==== SHIFT ====")
	//println(this.String())
	//println("shift(", a, ",", b, ")")
	this.matched.PushBack(this.right.Front().Value.(string))
	this.right.Remove(this.right.Front())
	this.backpath.PushBack(Pair{a, b})
	//println(this.String())
	//println("==== END ====")
}

type Chart struct {
	table []*list.List
	size  int
	gram  *Grammar
}

func NewChart(gram *Grammar) *Chart {
	return &Chart{
		gram: gram,
	}
}

func (this *Chart) getSize() int { return this.size }

/*
func (this *Chart) getCell(i int, j int) *list.List {
	return this.table[i][j]
}
*/

func (this *Chart) loadSentence(s *Sentence, k int) {
	var j, n int
	var w *list.Element

	//println("Loading sentence")

	n = s.Len()
	this.table = make([]*list.List, (1+n)*n/2)

	//println("Loading sentence 2")

	j = 0
	l := list.New()
	for w = s.Front(); w != nil; w = w.Next() {
		ce := list.New()
		//println("Processing word " + w.Value.(*Word).getForm())
		for a := w.Value.(*Word).selectedBegin(k).Element; a != w.Value.(*Word).selectedBegin(k).Next(); a = a.Next() {
			//println("selected tags")
			e := NewEdgeFromString(a.Value.(*Analysis).getTag(), l, 0)
			ce.PushBack(e)
			//println(" created edge " + a.Value.(*Analysis).getTag() + " in cell (0," + strconv.Itoa(j) + ")")
			this.findAllRules(e, ce, 0, j)

			e1 := NewEdgeFromString(a.Value.(*Analysis).getTag()+"("+w.Value.(*Word).getLCForm()+")", l, 0)
			ce.PushBack(e1)
			//println(" created edge " + a.Value.(*Analysis).getTag() + "(" + w.Value.(*Word).getLCForm() + ") in cell (0," + strconv.Itoa(j) + ")")
			this.findAllRules(e1, ce, 0, j)

			e2 := NewEdgeFromString(a.Value.(*Analysis).getTag()+"<"+a.Value.(*Analysis).getLemma()+">", l, 0)
			ce.PushBack(e2)
			//println(" created edge " + a.Value.(*Analysis).getTag() + "<" + a.Value.(*Analysis).getLemma() + "> in cell (0," + strconv.Itoa(j) + ")")
			this.findAllRules(e2, ce, 0, j)
		}

		this.table[this.index(0, j)] = ce
		j++
	}
	this.size = j

	for k := 0; k < j; k++ {
		//println("Rule[", k, "]=", this.table[this.index(0, k)].Len())
		ce := this.table[this.index(0, k)]
		//println("Found", ce.Len(), "Rules")
		for e := ce.Front(); e != nil; e = e.Next() {
			//println(e.Value.(*Edge).String())
			//println("_______")
		}
		//println("=========")
	}

	//println("Sentence loaded")
}

func (this *Chart) parse() {
	var k, i, a int
	for k = 1; k < this.size; k++ {
		for i = 0; i < this.size-k; i++ {
			ce := list.New()
			for a = 0; a < k; a++ {
				//println("Visiting cell (" + strconv.Itoa(a) + "," + strconv.Itoa(i) + ")")
				for ed := this.table[this.index(a, i)].Front(); ed != nil; ed = ed.Next() {
					e := NewEdgeFromEdge(ed.Value.(*Edge))
					if e.active() {
						//println("   Active edge for " + newed.getHead())
						ls := e.getRight()
						if this.canExtend(ls.Front().Value.(string), k-a-1, i+a+1) {
							//println("     it can be extended with " + ls.Front().Value.(string) + " at " + strconv.Itoa(k-a-1) + " " + strconv.Itoa(i+a+1))
							e.shift(k-a-1, i+a+1)
							ce.PushBack(e)
							if !e.active() {
								this.findAllRules(e, ce, k, i)
							}
						} else {
							//println("     it can NOT be extended with " + ls.Front().Value.(string) + " at " + strconv.Itoa(k-a-1) + " " + strconv.Itoa(i+a+1))
						}
					}
				}
			}
			this.table[this.index(k, i)] = ce
		}
	}

	this.ndump()

	best := NewEdge()
	gotroot := false

	for ed := this.table[this.index(this.size-1, 0)].Front(); ed != nil; ed = ed.Next() {
		if !ed.Value.(*Edge).active() && !this.gram.isNoTop(ed.Value.(*Edge).getHead()) && this.betterEdge(ed.Value.(*Edge), best) {
			gotroot = true
			best = ed.Value.(*Edge)
		}
	}

	if !gotroot {
		//println("adding fictitious root at [" + strconv.Itoa(this.size-1) + ",0]")
		lp := this.cover(this.size-1, 0)
		ls := list.New()
		for p := lp.Front(); p != nil; p = p.Next() {
			best = NewEdge()
			for ed := this.table[this.index(p.Value.(Pair).first.(int), p.Value.(Pair).second.(int))].Front(); ed != nil; ed = ed.Next() {
				if !ed.Value.(*Edge).active() && this.betterEdge(ed.Value.(*Edge), best) {
					best = ed.Value.(*Edge)
				}
			}

			ls.PushBack(best.getHead())
			//println("Inactive edge selected for (" + strconv.Itoa(p.Value.(Pair).first.(int)) + "," + strconv.Itoa(p.Value.(Pair).second.(int)) + ") is " + best.getHead())
		}

		e1 := NewEdgeFromString(this.gram.getStartSymbol(), ls, GRAMMAR_NOGOV)
		//println("created fictitious rule")

		for p := lp.Front(); p != nil; p = p.Next() {
			e1.shift(p.Value.(Pair).first.(int), p.Value.(Pair).second.(int))
		}

		this.table[this.index(this.size-1, 0)].PushBack(e1)
	}

	//this.ndump()

	//this.dump()

}

func (this *Chart) cover(a, b int) *list.List {
	x := 0
	y := 0
	var i, j int
	var f bool
	var ed *list.Element
	var lp, lr *list.List

	if a < 0 || b < 0 || a+b >= this.size {
		return list.New()
	}

	//println("Covering under (" + strconv.Itoa(a) + "," + strconv.Itoa(b) + ")")

	f = false

	best := NewEdge()

	for i = a; !f && i >= 0; i-- {
		for j = b; j < b+(a-i)+1; j++ {
			//println("ED len:" + strconv.Itoa(this.table[this.index(i, j)].Len()))
			for ed = this.table[this.index(i, j)].Front(); ed != nil; ed = ed.Next() {
				LOG.Tracef("ed.active:%b 1st best:%b", ed.Value.(*Edge).active(), this.betterEdge(ed.Value.(*Edge), best))
				if !ed.Value.(*Edge).active() && this.betterEdge(ed.Value.(*Edge), best) {
					x = i
					y = j
					best = ed.Value.(*Edge)
					f = true
				}
			}
		}
	}

	//println("  Highest cell found is (" + strconv.Itoa(x) + "," + strconv.Itoa(y) + ")")
	//println("   Pending (" + strconv.Itoa(y-b-1) + "," + strconv.Itoa(b) + ") (" + strconv.Itoa((a+b)-(x+y+1)) + "," + strconv.Itoa(x+y+1) + ")")

	if !f {
		LOG.Panic("Inconsistent chart or wrongly loaded sentence")
	}

	lp = this.cover(y-b-1, b)
	lr = this.cover((a+b)-(x+y+1), x+y+1)

	lp.PushBack(Pair{x, y})
	for tlr := lr.Front(); tlr != nil; tlr = tlr.Next() {
		lp.PushBack(tlr.Value.(Pair))
	}

	return lp
}

func (this *Chart) betterEdge(e1 *Edge, e2 *Edge) bool {
	h1 := e1.getHead()
	h2 := e2.getHead()
	start := this.gram.getStartSymbol()

	//fmt.Printf("Better Edge -> h1: %s h2:%s start:%s\n", h1, h2, start)

	if h1 == start && h2 != start {
		return true
	}
	if h1 != start && h2 == start {
		return false
	}

	//fmt.Printf("H1 is terminal: %t H2 is terminal: %t\n", this.gram.isTerminal(h1), this.gram.isTerminal(h2))
	//fmt.Printf("H1 specificity: %d H2 specificity: %d\n", this.gram.getSpecificity(h1), this.gram.getSpecificity(h2))
	if this.gram.isTerminal(h1) && this.gram.isTerminal(h2) {
		return this.gram.getSpecificity(h1) < this.gram.getSpecificity(h2)
	}

	//fmt.Printf("H1 priority: %d H2 priority: %d\n", this.gram.getPriority(h1), this.gram.getPriority(h2))
	//LOG.Tracef("e1 matched len: %d e2 matched len: %d", e1.getMatched().Len(), e2.getMatched().Len())

	if !this.gram.isTerminal(h1) && !this.gram.isTerminal(h2) {
		if this.gram.getPriority(h1) < this.gram.getPriority(h2) {
			return true
		}
		if this.gram.getPriority(h1) > this.gram.getPriority(h2) {
			return false
		}

		return e1.getMatched().Len() > e2.getMatched().Len()
	}

	return !this.gram.isTerminal(h1) && this.gram.isTerminal(h2)
}

func (this *Chart) index(i, j int) int {
	return j + i*(this.size+1) - (i+1)*i/2
}

func (this *Chart) canExtend(hd string, i int, j int) bool {
	b := false
	for ed := this.table[this.index(i, j)].Front(); !b && ed != nil; ed = ed.Next() {
		//println("   rule head:" + ed.Value.(*Edge).getHead() + ", active: " + strconv.FormatBool(ed.Value.(*Edge).active()))
		b = (!ed.Value.(*Edge).active() && this.checkMatch(hd, ed.Value.(*Edge).getHead()))
	}

	return b
}

func (this *Chart) checkMatch(searched string, found string) bool {
	var s, t, m string
	LOG.Tracef("check_match: searched %s found %s", searched, found)
	if searched == found {
		return true
	}

	n := strings.Index(searched, "*")
	if n == -1 {
		return false
	}

	if strings.Index(found, searched[0:n]) != 0 {
		return false
	}

	n = MultiIndex(found, "(<")
	if n == -1 {
		s = found
		t = ""
	} else {
		s = found[0:n]
		t = found[n:]
	}

	n = MultiIndex(searched, "(<")
	if n == -1 {
		m = ""
	} else {
		m = searched[n:]
	}

	file := strings.Index(m, "\"") > -1

	if !file {
		return (s+m == found)
	} else {
		return this.gram.inFileMap(t, m)
	}
}

func (this *Chart) findAllRules(e *Edge, ce *list.List, k int, i int) {
	d := list.New()
	if this.gram.isTerminal(e.getHead()) {
		lr := this.gram.getRulesRightWildcard(e.getHead()[0:1])
		for r := lr.Front(); r != nil; r = r.Next() {
			newR := NewRuleFromRule(r.Value.(*Rule))
			if this.checkMatch(newR.getRight().Front().Value.(string), e.getHead()) {
				//println("    --> Match for " + e.getHead() + ". adding WILDCARD rule [" + newR.getHead() + "==>" + newR.getRight().Front().Value.(string) + "...etc")
				ed := NewEdgeFromString(newR.getHead(), newR.getRight(), newR.getGovernor())
				ed.shift(k, i)
				ce.PushBack(ed)
				if !ed.active() {
					d.PushBack(ed.getHead())
				}
			}
		}
	}

	d.PushBack(e.getHead())
	for d.Len() > 0 {
		lr := this.gram.getRulesRight(d.Front().Value.(string))
		for r := lr.Front(); r != nil; r = r.Next() {
			newR := NewRuleFromRule(r.Value.(*Rule))
			//println("    --> adding rule [" + newR.getHead() + "==>" + newR.getRight().Front().Value.(string) + "..etc] with gov=" + strconv.Itoa(newR.getGovernor()))
			ed := NewEdgeFromString(newR.getHead(), newR.getRight(), newR.getGovernor())
			ed.shift(k, i)
			ce.PushBack(ed)
			if !ed.active() {
				d.PushBack(ed.getHead())
			}
		}
		d.Remove(d.Front())
	}

	//println("Found " + strconv.Itoa(ce.Len()) + " rules")
}

func (this *Chart) ndump() {
	for a := 0; a < this.size; a++ {
		for i := 0; i < this.size-a; i++ {
			if this.table[this.index(a, i)].Len() > 0 {
				//println("Cell (" + strconv.Itoa(a) + "," + strconv.Itoa(i) + ")\n")
				for ed := this.table[this.index(a, i)].Front(); ed != nil; ed = ed.Next() {
					//println(ed.Value.(*Edge).String())
				}
			}
		}
	}
}

func (this *Chart) dump() {
	for a := 0; a < this.size; a++ {
		for i := 0; i < this.size-a; i++ {
			if this.table[this.index(a, i)].Len() > 0 {
				out := "Cell (" + strconv.Itoa(a) + "," + strconv.Itoa(i) + ")\n"
				for ed := this.table[this.index(a, i)].Front(); ed != nil; ed = ed.Next() {
					out += "    " + ed.Value.(*Edge).getHead() + " ==>"
					ls := ed.Value.(*Edge).getMatched()
					for s := ls.Front(); s != nil; s = s.Next() {
						out += " " + s.Value.(string)
					}
					out += " ."
					ls = ed.Value.(*Edge).getRight()
					for s := ls.Front(); s != nil; s = s.Next() {
						out += " " + s.Value.(string)
					}
					lp := ed.Value.(*Edge).getBackpath()
					out += "   Backpath:"
					for p := lp.Front(); p != nil; p = p.Next() {
						out += "(" + strconv.Itoa(p.Value.(Pair).first.(int)) + "," + strconv.Itoa(p.Value.(Pair).second.(int)) + ")"
					}
					out += "\n"
				}
				//println(out)
			}
		}
	}
}

func (this *Chart) getTree(x int, y int, lab string) *ParseTree {
	label := lab
	if label == "" {
		best := NewEdge()
		for ed := this.table[this.index(x, y)].Front(); ed != nil; ed = ed.Next() {
			if !ed.Value.(*Edge).active() && !this.gram.isHidden(ed.Value.(*Edge).getHead()) && this.betterEdge(ed.Value.(*Edge), best) {
				label = ed.Value.(*Edge).getHead()
				best = ed.Value.(*Edge)
			}
		}
	}
	node := NewNodeFromLabel(label)
	tr := NewOneNodeParseTree(node)

	//fmt.Printf("  Building tree for (%d,%d):%s\n", x, y, label)
	if label == this.gram.getStartSymbol() || !this.gram.isTerminal(label) {
		best := NewEdge()
		for ed := this.table[this.index(x, y)].Front(); ed != nil; ed = ed.Next() {
			if !ed.Value.(*Edge).active() && label == ed.Value.(*Edge).getHead() && this.betterEdge(ed.Value.(*Edge), best) {
				best = ed.Value.(*Edge)
				//fmt.Printf("  Selected best: %s\n", best.getHead())
			}
		}

		//fmt.Printf("   expanding..\n")

		//println(best.String())

		r := best.getMatched()
		bp := best.getBackpath()
		g := best.getGovernor()

		headset := false
		//fmt.Printf("   Best Governor: " + strconv.Itoa(g) + " Best Matches: " + strconv.Itoa(r.Len()) + " Best Backpath:" + strconv.Itoa(bp.Len()) + "\n")

		for ch, s, p := 0, r.Front(), bp.Front(); s != nil && p != nil; ch, s, p = ch+1, s.Next(), p.Next() {
			//fmt.Printf("   Entering down to child %s\n", s.Value.(string))
			child := this.getTree(p.Value.(Pair).first.(int), p.Value.(Pair).second.(int), s.Value.(string))
			childLabel := child.begin().pnode.info.(*Node).getLabel()
			if this.gram.isHidden(childLabel) || this.gram.isOnlyTop(childLabel) || (this.gram.isFlat(childLabel) && label == childLabel) {
				//fmt.Printf("   -Child is hidden or flat %s  %s\n", label, childLabel)
				for x := child.siblingBegin(); x.pnode != child.siblingEnd().pnode; x = x.siblingPlusPlus() {
					if ch == g {
						headset = true
					} else {
						x.pnode.info.(*Node).setHead(false)
					}
					tr.appendChild(x.pnode)
				}

				//fmt.Printf("    skipped, sons raised. Headset=%s\n", If(headset, "YES", "NO").(string))
			} else {
				//fmt.Printf("    -Child is NOT hidden or flat %s %s\n", label, childLabel)
				if ch == g {
					child.begin().pnode.info.(*Node).setHead(true)
					headset = true
				}
				tr.appendChild(child)
				//fmt.Printf("    added. Headset=%s\n", If(headset, "YES", "NO").(string))
			}
		}

		if !headset && label != this.gram.getStartSymbol() {
			LOG.Warnf("  Unset rule governor for %s at (%d,%d)", label, x, y)
		}

	}

	return tr
}

package nlp

import (
	"container/list"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/set"
)

type ProcessorStatus struct {
}

type Analysis struct {
	lemma         string
	tag           string
	prob          float64
	distance      float64
	senses        *list.List
	retok         *list.List
	selectedKBest *set.Set

	User []string
}

func NewAnalysis(lemma string, tag string) *Analysis {
	return &Analysis{
		lemma:         lemma,
		tag:           tag,
		prob:          -1.0,
		distance:      -1.0,
		senses:        list.New(),
		retok:         list.New(),
		selectedKBest: set.New(set.ThreadSafe).(*set.Set),
	}
}

func NewAnalysisFromAnalysis(a *Analysis) *Analysis {
	selectedKBest := set.New(set.ThreadSafe).(*set.Set)
	selectedKBest.Add(a.selectedKBest.List()...)
	this := Analysis{
		lemma:         a.lemma,
		tag:           a.tag,
		prob:          a.prob,
		distance:      a.distance,
		senses:        list.New(),
		retok:         list.New(),
		selectedKBest: selectedKBest,
	}

	for s := a.senses.Front(); s != nil; s = s.Next() {
		this.senses.PushBack(s.Value.(*Senses))
	}

	for r := a.retok.Front(); r != nil; r = r.Next() {
		this.retok.PushBack(r.Value.(*Word))
	}

	return &this
}

func (this Analysis) String() string {
	return fmt.Sprintf("Lemma:%s [%s] - %f/%d", this.getLemma(), this.getTag(), this.getProb(), this.selectedKBest.Size())
}

func (this *Analysis) Tag() string {
	if this.selectedKBest.Size() > 0 {
		return this.tag
	} else {
		return ""
	}
}

func (this *Analysis) init(lemma string, tag string) {
	this.lemma = lemma
	this.tag = tag
	this.prob = -1.0
	this.distance = -1.0
	this.User = nil
	this.User = make([]string, 0)

	this.senses = this.senses.Init()
	this.retok = this.retok.Init()
	this.selectedKBest.Clear()
}

func (this *Analysis) setLemma(l string)              { this.lemma = l }
func (this *Analysis) setTag(p string)                { this.tag = p }
func (this *Analysis) setProb(p float64)              { this.prob = p }
func (this *Analysis) setDistance(d float64)          { this.distance = d }
func (this *Analysis) setRetokenizable(lw *list.List) { this.retok = lw }
func (this *Analysis) hasProb() bool                  { return this.prob >= 0.0 }
func (this *Analysis) getLemma() string               { return this.lemma }
func (this *Analysis) getProb() float64               { return this.prob }
func (this *Analysis) getDistance() float64           { return this.distance }
func (this *Analysis) isRetokenizable() bool          { return this.retok.Len() > 0 }
func (this *Analysis) getRetokenizable() *list.List   { return this.retok }
func (this *Analysis) getTag() string                 { return this.tag }
func (this *Analysis) getSenses() *list.List          { return this.senses }
func (this *Analysis) getSenseString() string         { return "" } //TODO
func (this *Analysis) setSenses(ls *list.List)        { this.senses = ls }
func (this *Analysis) maxKBest() int {
	kbest := -1
	this.selectedKBest.Each(func(it interface{}) bool {
		item := it.(int)
		if item > kbest {
			kbest = item
		}
		return true
	})
	return kbest
}
func (this *Analysis) isSelected(k int) bool { return this.selectedKBest.Has(k) }
func (this *Analysis) markSelected(k int)    { this.selectedKBest.Add(k) }
func (this *Analysis) unmarkSelected(k int)  { this.selectedKBest.Remove(k) }

type WordConstIterator struct {
	*list.Element
	ibeg  *list.Element
	iend  *list.Element
	tpe   int
	kbest int
}

type Word struct {
	*list.List
	form          string
	lcForm        string
	phForm        string
	multiword     *list.List
	ambiguousMw   bool
	alternatives  *list.List
	start, finish int
	inDict        bool
	locked        bool
	position      int
	SELECTED      int
	UNSELECTED    int
	ALL           int
	user          []string
	expired       bool
}

func NewWord() *Word {
	w := Word{
		multiword:    list.New(),
		alternatives: list.New(),
		SELECTED:     0,
		UNSELECTED:   1,
		ALL:          2,
		expired:      false,
	}
	w.List = list.New()
	return &w
}

func NewWordFromLemma(f string) *Word {
	w := Word{
		form:         f,
		phForm:       "",
		lcForm:       strings.ToLower(f),
		inDict:       true,
		locked:       false,
		ambiguousMw:  false,
		position:     -1,
		expired:      false,
		multiword:    list.New(),
		alternatives: list.New(),
	}
	w.List = list.New()
	return &w
}

func NewMultiword(f string, a *list.List) *Word {
	w := Word{
		form:        f,
		phForm:      "",
		lcForm:      strings.ToLower(f),
		multiword:   a,
		start:       a.Front().Value.(*Word).getSpanStart(),
		finish:      a.Back().Value.(*Word).getSpanFinish(),
		inDict:      true,
		locked:      false,
		ambiguousMw: false,
		position:    -1,
		expired:     false,
	}
	w.List = list.New()
	return &w
}

func (this Word) String() string {
	return this.form
}

func (this *Word) clone(w Word) {
	this.form = w.form
	this.phForm = w.phForm
	this.multiword = w.multiword
	this.start = w.start
	this.inDict = w.inDict
	this.locked = w.locked
	this.user = w.user
	this.alternatives = w.alternatives
	this.ambiguousMw = w.ambiguousMw
	this.position = w.position
}

func (this *Word) copyAnalysis(w *Word) {
	this.List = this.List.Init()

	for i := w.Front(); i != nil; i = i.Next() {
		this.PushBack(i.Value.(*Analysis))
	}
}

func (this *Word) setAnalysis(analysis ...interface{}) {
	this.List = this.List.Init()
	for _, analysis := range analysis {
		this.PushBack(analysis.(*Analysis))
		this.Back().Value.(*Analysis).markSelected(0)
	}
}

func (this *Word) addAnalysis(analysis *Analysis) {
	this.PushBack(analysis)
	this.Back().Value.(*Analysis).markSelected(0)
}

func (this *Word) getNAnalysis() int               { return this.Len() }
func (this *Word) lockAnalysis()                   { this.locked = true }
func (this *Word) isLocked() bool                  { return this.locked }
func (this *Word) getAnalysis() *list.List         { return this.List }
func (this *Word) getAnalysisBegin() *list.Element { return this.Front() }
func (this *Word) getAnalysisEnd() *list.Element   { return this.Back() }

func (this *Word) selectAllAnalysis(k int) {
	for i := this.Front(); i != nil; i = i.Next() {
		i.Value.(*Analysis).markSelected(k)
	}
}

func (this *Word) unselectAllAnalysis(k int) {
	for i := this.Front(); i != nil; i = i.Next() {
		i.Value.(*Analysis).unmarkSelected(k)
	}
}

func (this *Word) selectAnalysis(tag *Analysis, k int) { tag.markSelected(k) }

func (this *Word) getNSelected(k int) int {
	n := 0
	for i := this.Front(); i != nil; i = i.Next() {
		if i.Value.(*Analysis).isSelected(k) {
			n++
		}
	}
	return n
}

func (this *Word) getNUnselected(k int) int      { return this.Len() - this.getNSelected(k) }
func (this *Word) isMultiword() bool             { return this.multiword.Len() > 0 }
func (this *Word) isAmbiguousMw() bool           { return this.ambiguousMw }
func (this *Word) setAmbiguousMw(a bool)         { this.ambiguousMw = a }
func (this *Word) getNWordsMw() int              { return this.multiword.Len() }
func (this *Word) getWordsMw() *list.List        { return this.multiword }
func (this *Word) setSpan(start int, finish int) { this.start = start; this.finish = finish }
func (this *Word) getSpanStart() int             { return this.start }
func (this *Word) getSpanFinish() int            { return this.finish }

func (this *Word) findTagMatch(re *regexp.Regexp) bool {
	found := false
	for an := this.Front(); an != nil && !found; an = an.Next() {
		tag := an.Value.(*Analysis).getTag()
		found = re.MatchString(tag)
	}

	return found
}

func (this *Word) selectedBegin(k int) *WordConstIterator {
	p := this.Front()
	for p != nil && !p.Value.(*Analysis).isSelected(k) {
		p = p.Next()
	}

	wci := WordConstIterator{}
	wci.Element = p
	wci.ibeg = this.Front()
	wci.iend = this.Back()
	wci.tpe = this.SELECTED
	wci.kbest = k

	return &wci

}

func (this *Word) unselectedBegin(k int) *WordConstIterator {
	p := this.Front()
	for p != nil {
		if p != this.Back() && p.Value.(*Analysis).isSelected(k) {
			p = p.Next()
		}
	}

	wci := WordConstIterator{}
	wci.Element = p
	wci.ibeg = this.Front()
	wci.iend = this.Back()
	wci.tpe = this.UNSELECTED
	wci.kbest = k
	return &wci
}

func (this *Word) selectedEnd(k int) *list.Element { return this.Back() }

func (this *Word) unselectedEnd(k int) *list.Element { return this.Back() }

func (this *Word) numKBest() int {
	mx := 0
	for a := this.Front(); a != nil; a = a.Next() {
		y := a.Value.(*Analysis).maxKBest() + 1
		mx = If(y > mx, y, mx).(int)
	}

	return mx
}

func (this *Word) setForm(form string) { this.form = form }
func (this *Word) getForm() string     { return this.form }

func (this *Word) FormTag() string {
	form := this.getForm()
	tag := ""
	var senses *list.List
	for a := this.Front(); a != nil; a = a.Next() {
		if a.Value.(*Analysis).selectedKBest.Size() > 0 {
			tag = a.Value.(*Analysis).getTag()
			senses = a.Value.(*Analysis).getSenses()
			break
		}
	}

	dsb := make([]string, 0)
	for s := senses.Front(); s != nil; s = s.Next() {
		if s.Value.(FloatPair).second > 0 && s.Value.(FloatPair).first != "" {
			dsb = append(dsb, s.Value.(FloatPair).first+":"+strconv.FormatFloat(s.Value.(FloatPair).second, 'f', 3, 64))
		}
	}
	return fmt.Sprintf("%s/%s:[%s]", form, tag, strings.Join(dsb, "|"))
}

func (this *Word) getLCForm() string { return this.lcForm }
func (this *Word) getPHForm() string { return this.phForm }

func (this *Word) getLemma(k int) string {
	if this.getNAnalysis() != 0 {
		return this.selectedBegin(k).Value.(*Analysis).getLemma()
	} else {
		return ""
	}
}

func (this *Word) getTag(k int) string {
	if this.getNAnalysis() != 0 {
		return this.selectedBegin(k).Value.(*Analysis).getTag()
	} else {
		return ""
	}
}

func (this *Word) getSenses(k int) *list.List {
	return this.selectedBegin(k).Value.(*Analysis).getSenses()
}

func (this *Word) setSenses(ls *list.List, k int) {
	this.selectedBegin(k).Value.(*Analysis).setSenses(ls)
}

func (this *Word) getPosition() int      { return this.position }
func (this *Word) setPosition(i int)     { this.position = i }
func (this *Word) foundInDict() bool     { return this.inDict }
func (this *Word) setFoundInDict(b bool) { this.inDict = b }

func (this *Word) hasRetokenizable() bool {
	has := false
	for i := this.Front(); i != nil; i = i.Next() {
		has = i.Value.(*Analysis).isRetokenizable()
	}

	return has
}

func (this *Word) sort() {
	tmpLs := make([]*Analysis, this.Len())
	count := 0
	for i := this.Front(); i != nil; i = i.Next() {
		tmpLs[count] = i.Value.(*Analysis)
		count++
	}

	for i := 0; i < len(tmpLs); i++ {
		for j := i + 1; j < len(tmpLs); j++ {
			if tmpLs[i].getProb() < tmpLs[j].getProb() {
				tmp := tmpLs[j]
				tmpLs[j] = tmpLs[i]
				tmpLs[i] = tmp
			}
		}
	}

	this.List = this.List.Init()
	for i := 0; i < len(tmpLs); i++ {
		this.PushBack(tmpLs[i])
	}
}

type Sentence struct {
	*list.List
	sentID   string
	wpos     []*Word
	pts      map[int]*ParseTree
	status   *list.List
	predArgs map[int]Pair
}

func NewSentence() *Sentence {
	sentence := Sentence{}
	sentence.List = list.New()
	sentence.status = list.New()
	sentence.predArgs = make(map[int]Pair)
	sentence.pts = make(map[int]*ParseTree)
	return &sentence
}

func (this *Sentence) setSentenceID(sid string) { this.sentID = sid }
func (this *Sentence) getSentenceID() string    { return this.sentID }

func (this *Sentence) setParseTree(tr *ParseTree, k int) {
	this.pts[k] = tr
	this.pts[k].rebuildNodeIndex()
}

func (this *Sentence) getParseTree(k int) *ParseTree { return this.pts[k] }

func (this *Sentence) getProcessingStatus() interface{}   { return this.status.Back().Value }
func (this *Sentence) setProcessingStatus(st interface{}) { this.status.PushBack(st) }
func (this *Sentence) clearProcessingStatus() {
	if this.status.Len() == 0 {
		return
	}
	status := this.status.Back()
	this.status.Remove(status)
}

func (this *Sentence) rebuildWordIndex() {
	this.wpos = make([]*Word, this.Len())
	i := 0
	for w := this.Front(); w != nil; w = w.Next() {
		if w.Value.(*Word).expired {
			j := w.Prev()
			LOG.Trace("Word " + w.Value.(*Word).getForm() + " is going to be erased")
			EmptyFunc(this.Remove(w))
			w = j
		} else {
			LOG.Trace("Word " + w.Value.(*Word).getForm() + " has " + strconv.Itoa(w.Value.(*Word).getNAnalysis()) + " analysis")
			this.wpos[i] = w.Value.(*Word)
			w.Value.(*Word).setPosition(i)
			i++
		}
	}

	/*
		TODO
		if this.isParsed

		if this.isDepParsed

	*/
}

func (this *Sentence) numKBest() int {
	if this.Len() == 0 {
		return 0
	} else {
		return this.Front().Value.(*Word).numKBest()
	}
}

type Node struct {
	nodeID string
	head   bool
	chunk  int
	label  string
	w      *Word
	user   []string
}

func NewNode() *Node {
	return &Node{
		label:  "",
		w:      nil,
		head:   false,
		chunk:  0,
		nodeID: "-",
	}
}

func NewNodeFromLabel(label string) *Node {
	return &Node{
		label:  label,
		w:      nil,
		head:   false,
		chunk:  0,
		nodeID: "-",
	}
}

func (this *Node) getNodeID() string     { return this.nodeID }
func (this *Node) setNodeID(id string)   { this.nodeID = id }
func (this *Node) getLabel() string      { return this.label }
func (this *Node) setLabel(label string) { this.label = label }
func (this *Node) getWord() *Word        { return this.w }
func (this *Node) setWord(w *Word)       { this.w = w }
func (this *Node) isHead() bool          { return this.head }
func (this *Node) setHead(head bool)     { this.head = head }
func (this *Node) isChunk() bool         { return this.chunk != 0 }
func (this *Node) setChunk(chunk int)    { this.chunk = chunk }
func (this *Node) getChunkOrd() int      { return this.chunk }

type ParseTree struct {
	isEmpty     bool
	parent      *ParseTree
	first, last *ParseTree
	prev, next  *ParseTree
	info        interface{}
	nodeIndex   map[string]*ParseTree
	wordIndex   []*ParseTree
}

func NewEmptyParseTree() *ParseTree {
	return &ParseTree{
		isEmpty: true,
		parent:  nil,
		first:   nil,
		last:    nil,
		prev:    nil,
		next:    nil,
	}
}

func NewOneNodeParseTree(info interface{}) *ParseTree {
	return &ParseTree{
		isEmpty: false,
		parent:  nil,
		first:   nil,
		last:    nil,
		prev:    nil,
		next:    nil,
		info:    info,
	}
}

func NewParseTreeFromParseTree(t *ParseTree) *ParseTree {
	this := NewEmptyParseTree()
	this.Clone(t)
	return this
}

func (this *ParseTree) Clear() {
	p := this.first
	for p != nil {
		q := p.next
		p = nil
		p = q
	}

	this.isEmpty = true
	this.parent = nil
	this.first = nil
	this.last = nil
	this.prev = nil
	this.next = nil
}

func (this *ParseTree) numChildren() int {
	n := 0
	for s := this.first; s != nil; s = s.next {
		n++
	}
	return n
}

func (this *ParseTree) getParent() *ParseTree {
	return this.parent
}

func (this *ParseTree) nthChild(n int) *ParseTree {
	i := this.first
	for n > 0 && i != nil {
		i = i.next
		n--
	}

	return i
}

func (this *ParseTree) GetInfo() interface{} {
	return this.info
}

func (this *ParseTree) Empty() bool {
	return this.isEmpty
}

func (this *ParseTree) isRoot() bool {
	return this.parent == nil
}

func (this *ParseTree) hasAncestor(p *ParseTree) bool {
	t := this
	for t != nil && t != p {
		t = t.parent
	}
	return t != nil
}

func (this *ParseTree) appendChild(child *ParseTree) {
	x := NewEmptyParseTree()
	x.Clone(child)

	x.next = nil
	x.prev = nil
	x.parent = this

	if this.first != nil {
		x.prev = this.last
		this.last.next = x
		this.last = x
	} else {
		this.first = x
		this.last = x
	}
}

func (this *ParseTree) hangChild(child *ParseTree, last bool) {
	if child.prev != nil {
		child.prev.next = child.next
	}
	if child.next != nil {
		child.next.prev = child.prev
	}

	if child.parent != nil {
		if child.prev == nil {
			child.parent.first = child.next
		}
		if child.next == nil {
			child.parent.last = child.prev
		}
	}

	child.prev = nil
	child.next = nil
	child.parent = this

	if this.first == nil {
		this.first = child
		this.last = child
	} else {
		if last {
			child.prev = this.last
			this.last.next = child
		} else {
			child.next = this.first
			this.first.prev = child
			this.first = child
		}
	}
}

func (this *ParseTree) Clone(t *ParseTree) {
	this.isEmpty = t.isEmpty
	this.info = t.info
	this.parent = nil
	this.first = nil
	this.last = nil
	this.prev = nil
	this.next = nil

	for p := t.first; p != nil; p = p.next {
		c := NewEmptyParseTree()
		c.Clone(p)
		c.next = nil
		c.prev = nil
		c.parent = this

		if this.first != nil {
			c.prev = this.last
			this.last.next = c
			this.last = c
		} else {
			this.first = c
			this.last = c
		}
	}

}

func (this *ParseTree) siblingBegin() *ParseTreeIterator {
	return NewParseTreeIteratorFromParseTree(this.first)
}

func (this *ParseTree) siblingEnd() *ParseTreeIterator {
	return NewEmptyParseTreeIterator()
}

func (this *ParseTree) begin() *ParseTreeIterator {
	if this.isEmpty {
		return NewEmptyParseTreeIterator()
	} else {
		return NewParseTreeIteratorFromParseTree(this)
	}
}

func (this *ParseTree) end() *ParseTreeIterator {
	return NewEmptyParseTreeIterator()
}

type ParseTreeIterator struct {
	pnode *ParseTree
}

func NewEmptyParseTreeIterator() *ParseTreeIterator {
	return &ParseTreeIterator{
		pnode: nil,
	}
}

func NewParseTreeIteratorFromParseTree(t *ParseTree) *ParseTreeIterator {
	return &ParseTreeIterator{
		pnode: t,
	}
}

func NewParseTreeIteratorFromParseTreeIterator(o *ParseTreeIterator) *ParseTreeIterator {
	return &ParseTreeIterator{
		pnode: o.pnode,
	}
}

func (this *ParseTreeIterator) siblingPlusPlus() *ParseTreeIterator {
	this.pnode = this.pnode.next
	return this
}

func (this *ParseTreeIterator) siblingMinusMinus() *ParseTreeIterator {
	this.pnode = this.pnode.prev
	return this
}

func (this *ParseTreeIterator) PlusPlus() *ParseTreeIterator {
	if this.pnode.first != nil {
		this.pnode = this.pnode.first
	} else {
		for this.pnode != nil && this.pnode.next == nil {
			this.pnode = this.pnode.parent
		}
		if this.pnode != nil {
			this.pnode = this.pnode.next
		}
	}
	return this
}

func (this *ParseTreeIterator) MinusMinus() *ParseTreeIterator {
	if this.pnode.prev != nil {
		this.pnode = this.pnode.prev
		for this.pnode.last != nil {
			this.pnode = this.pnode.last
		}
	} else {
		this.pnode = this.pnode.parent
	}
	return this
}

func (this *ParseTree) buildNodeIndex(sid string) {
	this.nodeIndex = make(map[string]*ParseTree)
	for i, k := 0, this.begin(); i > -1 && k.pnode != this.end().pnode; i, k = i+1, k.PlusPlus() {
		id := sid + "." + strconv.Itoa(i)
		k.pnode.info.(*Node).setNodeID(id)
		this.nodeIndex[id] = k.pnode
	}
}

func (this *ParseTree) rebuildNodeIndex() {
	this.nodeIndex = make(map[string]*ParseTree)
	this.wordIndex = make([]*ParseTree, 0)
	for i, k := 0, this.begin(); i > -1 && k.pnode != this.end().pnode; i, k = i+1, k.PlusPlus() {
		id := k.pnode.info.(*Node).getNodeID()
		if id != "-" {
			this.nodeIndex[id] = k.pnode
		}
		if k.pnode.numChildren() == 0 {
			this.wordIndex = append(this.wordIndex, k.pnode)
		}
	}
}

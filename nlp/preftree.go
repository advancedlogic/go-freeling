package nlp

type List struct {
	begin, end *ListRecBase
}

func NewList() *List {
	return &List{
		begin: nil,
		end:   nil,
	}
}

/*
func (this *List) push(c rune, wordEnd bool) interface {} {
	n := this.find(c)
	if n == nil {
		if !wordEnd {
			n = NewListRecBase(c)
		} else {
			n = NewListRecBase(c)
		}

		if this.end != nil {
			this.end.next = n
		} else {
			this.begin = n
		}
		this.end = n
	} else if wordEnd {
		tmp := this.begin
		var prev *ListRecBase = nil
		for tmp != n {
			prev = tmp
			tmp = tmp.next
		}

		if n == this.end {
			this.end = nil
		}

		ntmp := n.next
		l := n.nextList
		n = NewListRecData(c)
		n.next = ntmp
		n.nextList = l

		if prev != nil {
			prev.next = n
		} else {
			this.begin = n
		}

		if this.end == nil {
			this.end = n
		}
	}

	if !wordEnd {
		if n.nextList == nil {
			n.nextList = NewList()
		}
		return n.nextList
	} else {
		return n
	}
}
*/
func (this *List) find(c rune) *ListRecBase {
	if this.begin == nil {
		return nil
	}

	tmp := this.begin
	for tmp != nil && tmp.symb != c {
		if tmp.next != nil {
			if tmp.next.symb == c {
				return tmp.next
			}
			tmp = tmp.next.next
		} else {
			tmp = nil
			break
		}
	}

	return If(tmp != nil, tmp, nil).(*ListRecBase)
}

type ListRecBase struct {
	symb     rune
	next     *ListRecBase
	nextList *List
}

func NewListRecBase(s rune) *ListRecBase {
	return &ListRecBase{
		symb:     s,
		next:     nil,
		nextList: nil,
	}
}

/*
func (this *List) find(c rune) *ListRecBase {
	if this.begin == nil {
		return nil
	}
	tmp := this.begin
	for tmp.symb != c {
		if tmp.next != nil {
			if tmp.next.symb == c {
				return tmp.next
			}
			tmp = tmp.next.next
		} else {
			tmp = nil
			break
		}
	}

	return If(tmp != nil, tmp, nil).(*ListRecBase)
}
*/

type ListRec struct {
	*ListRecBase
}

func NewListRec(s rune) *ListRec {
	return &ListRec{
		ListRecBase: NewListRecBase(s),
	}
}

type ListRecEnd struct {
	*ListRecBase
}

func NewListRecEnd(s rune) *ListRec {
	return &ListRec{
		ListRecBase: NewListRecBase(s),
	}
}

type ListRecData struct {
	*ListRecBase
	value []rune
}

func NewListRecData(s rune) *ListRec {
	return &ListRec{
		ListRecBase: NewListRecBase(s),
	}
}

func (this *ListRecData) setValue(value []rune) {
	this.value = value
}

func (this *ListRecData) getValue() []rune {
	return this.value
}

func (this *ListRecData) delete() {
	this.value = nil
}

type PrefTree struct {
	root      *List
	DELIM_LEN int
}

func NewPrefTree() *PrefTree {
	return &PrefTree{
		root:      NewList(),
		DELIM_LEN: len(" "),
	}
}

/*
func (this *PrefTree) addWord(word string, lr *ListRecBase, sIn string) {
	if word == "" {
		return
	}

	var l *List = nil
	if lr == nil {
		l = this.root
		for w := 0; w < len(word); w++ {
			l = l.push(rune(word[w]),w < len(word)).(*List)
		}
	} else {
		//l = lr.(*List)
	}

	var ldata *ListRecData = l.(*ListRecData)
	data := ldata.getValue()
	var newStr []rune
	s := 0
	if data != nil {
		s = len(data) + this.DELIM_LEN + len(sIn) + 1
		newStr = make([]rune, 0)
		newStr = append(newStr, data...)
		newStr = append(newStr, " "[0])
	} else {
		s = len(sIn) + 1
		newStr = make([]rune,1)
		newStr[0] = 0
	}

	newStr = append(newStr, sIn)
	ldata.setValue(newStr)
}
*/

/*
func (this *PrefTree) findWord(word string) *ListRecBase {
	n := this.root
	var lr *ListRecBase = n.find(word[0])
	for word != "" && n != nil && lr != nil {
		n = lr.next
	}
}
*/

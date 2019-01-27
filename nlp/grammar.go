package nlp

import (
	"container/list"
	"github.com/fatih/set"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

type Rule struct {
	head  string
	right *list.List
	gov   int
}

func NewRule() *Rule {
	return &Rule{
		right: list.New(),
		gov:   0,
	}
}

func NewRuleFromString(s string, ls *list.List, p int) *Rule {
	r := list.New()
	for i := ls.Front(); i != nil; i = i.Next() {
		r.PushBack(i.Value.(string))
	}

	return &Rule{
		head:  s,
		right: r,
		gov:   p,
	}
}

func NewRuleFromRule(r *Rule) *Rule {
	this := Rule{
		head: r.head,
		gov:  r.gov,
	}

	this.right = list.New()
	for i := r.right.Front(); i != nil; i = i.Next() {
		this.right.PushBack(i.Value.(string))
	}

	return &this
}

func (this *Rule) setGovernor(p int)    { this.gov = p }
func (this *Rule) getGovernor() int     { return this.gov }
func (this *Rule) getHead() string      { return this.head }
func (this *Rule) getRight() *list.List { return this.right }

const (
	GRAMMAR_CATEGORY = 1 + iota
	GRAMMAR_FORM
	GRAMMAR_LEMMA
	GRAMMAR_COMMENT
	GRAMMAR_HEAD
	GRAMMAR_ARROW
	GRAMMAR_BAR
	GRAMMAR_COMMA
	GRAMMAR_DOT
	GRAMMAR_FLAT
	GRAMMAR_HIDDEN
	GRAMMAR_NOTOP
	GRAMMAR_ONLYTOP
	GRAMMAR_PRIOR
	GRAMMAR_START
	GRAMMAR_FILENAME
)

const GRAMMAR_NOGOV = 99999
const GRAMMAR_DEFGOV = 0

type RulesMap map[string]*list.List

type Grammar struct {
	RulesMap
	nonterminal *set.Set
	wild        RulesMap
	filemap     RulesMap
	prior       map[string]int
	hidden      *set.Set
	flat        *set.Set
	notop       *set.Set
	onlytop     *set.Set
	NOGOV       int
	DEFGOV      int
	start       string
}

func NewGrammar(fname string) *Grammar {
	this := Grammar{
		RulesMap:    make(RulesMap),
		nonterminal: set.New(set.ThreadSafe).(*set.Set),
		wild:        make(RulesMap),
		filemap:     make(RulesMap),
		prior:       make(map[string]int),
		hidden:      set.New(set.ThreadSafe).(*set.Set),
		flat:        set.New(set.ThreadSafe).(*set.Set),
		notop:       set.New(set.ThreadSafe).(*set.Set),
		onlytop:     set.New(set.ThreadSafe).(*set.Set),
	}

	MAX := 32
	var tok, stat, newstat, i, j int
	what := 0
	var trans [32][32]int
	first := false
	wildcard := false
	var head, err, categ, name string
	ls := list.New()
	var priorVal int

	for i = 0; i < MAX; i++ {
		for j = 0; j < MAX; j++ {
			trans[i][j] = 0
		}
	}

	trans[1][GRAMMAR_COMMENT] = 1
	trans[1][GRAMMAR_CATEGORY] = 2
	trans[1][GRAMMAR_PRIOR] = 6
	trans[1][GRAMMAR_START] = 8
	trans[1][GRAMMAR_HIDDEN] = 6
	trans[1][GRAMMAR_FLAT] = 6
	trans[1][GRAMMAR_NOTOP] = 6
	trans[1][GRAMMAR_ONLYTOP] = 6

	trans[2][GRAMMAR_ARROW] = 3

	trans[3][GRAMMAR_CATEGORY] = 4
	trans[3][GRAMMAR_HEAD] = 10

	trans[4][GRAMMAR_COMMA] = 3
	trans[4][GRAMMAR_BAR] = 3
	trans[4][GRAMMAR_DOT] = 1
	trans[4][GRAMMAR_LEMMA] = 5
	trans[4][GRAMMAR_FORM] = 5
	trans[4][GRAMMAR_FILENAME] = 5

	trans[5][GRAMMAR_COMMA] = 3
	trans[5][GRAMMAR_BAR] = 3
	trans[5][GRAMMAR_DOT] = 1

	trans[6][GRAMMAR_CATEGORY] = 7

	trans[7][GRAMMAR_CATEGORY] = 7
	trans[7][GRAMMAR_DOT] = 1

	trans[8][GRAMMAR_CATEGORY] = 9

	trans[9][GRAMMAR_DOT] = 1

	trans[10][GRAMMAR_CATEGORY] = 4

	rules := make([]*Pair, 0)

	rules = append(rules, &Pair{regexp.MustCompile("[ \\t\\n\\r]+"), 0})
	rules = append(rules, &Pair{regexp.MustCompile("%.*"), GRAMMAR_COMMENT})
	rules = append(rules, &Pair{regexp.MustCompile("==>"), GRAMMAR_ARROW})
	rules = append(rules, &Pair{regexp.MustCompile("\\([[:alpha:]_'·\\-]+\\)"), GRAMMAR_FORM})
	rules = append(rules, &Pair{regexp.MustCompile("<[[:lower:]_'·\\-]+>"), GRAMMAR_LEMMA})
	rules = append(rules, &Pair{regexp.MustCompile("\\(\\\"([A-Za-z]:)?[[:alnum:]_\\-\\./\\\\]+\\\"\\)"), GRAMMAR_FILENAME})
	rules = append(rules, &Pair{regexp.MustCompile("<\\\"([A-Za-z]:)?[[:alnum:]_\\-\\./\\\\]+\\\">"), GRAMMAR_FILENAME})
	rules = append(rules, &Pair{regexp.MustCompile("[A-Za-z][\\-A-Za-z0-9]*[*]?"), GRAMMAR_CATEGORY})
	rules = append(rules, &Pair{regexp.MustCompile("@PRIOR"), GRAMMAR_PRIOR})
	rules = append(rules, &Pair{regexp.MustCompile("@START"), GRAMMAR_START})
	rules = append(rules, &Pair{regexp.MustCompile("@HIDDEN"), GRAMMAR_HIDDEN})
	rules = append(rules, &Pair{regexp.MustCompile("@FLAT"), GRAMMAR_FLAT})
	rules = append(rules, &Pair{regexp.MustCompile("@NOTOP"), GRAMMAR_NOTOP})
	rules = append(rules, &Pair{regexp.MustCompile("@ONLYTOP"), GRAMMAR_ONLYTOP})
	rules = append(rules, &Pair{regexp.MustCompile("\\|"), GRAMMAR_BAR})
	rules = append(rules, &Pair{regexp.MustCompile("\\."), GRAMMAR_DOT})
	rules = append(rules, &Pair{regexp.MustCompile(","), GRAMMAR_COMMA})
	rules = append(rules, &Pair{regexp.MustCompile("\\+"), GRAMMAR_HEAD})

	fl := NewLexer(rules)

	filestr, e := ioutil.ReadFile(fname)
	if e != nil {
		CRASH("Error opening file "+fname, MOD_GRAMMAR)
	}
	gov := 0
	havegov := false
	stat = 1
	priorVal = 1
	stream := string(filestr)
	err = ""
	for {
		tok = fl.getToken(stream)
		if tok == -1 {
			break
		}
		newstat = trans[stat][tok]
		switch newstat {
		case 0:
			{
				if tok == GRAMMAR_COMMENT {
					err = "Unexpected comment. Missing dot ending previous rule/directive ?"
				}
				if err == "" {
					err = "Unexpected '" + fl.getText() + "' found."
				}

				LOG.Warnln("File "+fname+", line "+strconv.Itoa(fl.lineno())+":"+err, MOD_GRAMMAR)

				for tok > -1 && tok != GRAMMAR_DOT {
					tok = fl.getToken(stream)
				}
				newstat = 1
				break
			}
		case 1:
			{
				if tok == GRAMMAR_DOT && (stat == 4 || stat == 5) {
					ls.PushBack(categ)
					if !havegov {
						gov = GRAMMAR_DEFGOV
						if ls.Len() != 1 {
							err = "Non-unary rule with no governor. First component taken as governor."
							LOG.Warnln("File "+fname+", line "+strconv.Itoa(fl.lineno())+":"+err, MOD_GRAMMAR)
						}
					}
					this.newRule(head, ls, wildcard, gov)
					gov = GRAMMAR_NOGOV
					havegov = false
				}
				break
			}
		case 2:
			{
				head = fl.getText()
				this.nonterminal.Add(head)
				break
			}
		case 3:
			{
				if tok == GRAMMAR_ARROW {
					ls = list.New()
					first = true
					wildcard = false
				} else if tok == GRAMMAR_COMMA {
					ls.PushBack(categ)
				} else if tok == GRAMMAR_BAR {
					ls.PushBack(categ)
					if !havegov {
						gov = GRAMMAR_DEFGOV
						if ls.Len() != -1 {
							err = "Non-unary rule with no governor. First component taken as governor."
							LOG.Warnln("File "+fname+", line "+strconv.Itoa(fl.lineno())+":"+err, MOD_GRAMMAR)
						}
					}

					this.newRule(head, ls, wildcard, gov)

					gov = GRAMMAR_NOGOV
					havegov = false
					ls = list.New()
					break
				}
			}
		case 4:
			{
				categ = fl.getText()
				if first && strings.Index(categ, "*") > -1 {
					wildcard = true
				}
				first = false
				break
			}
		case 5:
			{
				name = fl.getText()
				categ = categ + name

				if tok == GRAMMAR_FILENAME {
					var sname string

					sname = name[2 : len(name)-2]
					sname = fname[0:strings.LastIndex(fname, "/")+1] + "/" + sname

					fs, e := ioutil.ReadFile(sname)
					if e != nil {
						LOG.Stackln("Error opening file " + sname)
					}

					var op, clo string
					if string(name[0]) == "<" {
						op = "<"
						clo = ">"
					} else if string(name[0]) == "(" {
						op = "("
						clo = ")"
					}

					lines := Split(string(fs), "\n")
					for _, line := range lines {
						lfm, ok := this.filemap[op+line+clo]
						if !ok {
							this.filemap[op+line+clo] = list.New()
							lfm = this.filemap[op+line+clo]
						}
						exists := false
						for l := lfm.Front(); l != nil && !exists; l = l.Next() {
							if l.Value.(string) == name {
								exists = true
								break
							}
						}
						if !exists {
							lfm.PushBack(name)
						}
					}
				}

				break
			}
		case 6:
			{
				what = tok
				break
			}
		case 7:
			{
				categ = fl.getText()
				if this.nonterminal.Has(categ) {
					switch what {
					case GRAMMAR_PRIOR:
						{
							_, ok := this.prior[categ]
							if !ok {
								this.prior[categ] = priorVal
								priorVal++
							}
							break
						}
					case GRAMMAR_HIDDEN:
						{
							this.hidden.Add(categ)
							break
						}
					case GRAMMAR_FLAT:
						{
							this.flat.Add(categ)
							break
						}
					case GRAMMAR_NOTOP:
						{
							this.notop.Add(categ)
							break
						}
					case GRAMMAR_ONLYTOP:
						{
							this.onlytop.Add(categ)
							break
						}
					default:
						break
					}
				} else {
					err = "Terminal symbol '" + fl.getText() + "' not allowed in directive."
					newstat = 0
				}
				break
			}
		case 8:
			{
				if this.start != "" {
					err = "@START specified more than once."
					newstat = 0
				}
				break
			}
		case 9:
			{
				this.start = fl.getText()
				if !this.nonterminal.Has(this.start) {
					this.nonterminal.Add(this.start)
				}
				break
			}
		case 10:
			{
				gov = ls.Len()
				havegov = true
				break
			}
		default:
			break
		}

		stat = newstat
	}

	if this.start == "" {
		err = "@START symbol not specified."
		LOG.Warnln("File " + fname + ", line " + strconv.Itoa(fl.lineno()) + ":" + err)
	}
	if this.hidden.Has(this.start) {
		err = "@START symbol cannot be @HIDDEN."
		LOG.Warnln("File " + fname + ", line " + strconv.Itoa(fl.lineno()) + ":" + err)
	}
	if this.notop.Has(this.start) {
		err = "@START symbol cannot be @NOTOP."
		LOG.Warnln("File " + fname + ", line " + strconv.Itoa(fl.lineno()) + ":" + err)
	}

	for _, x := range this.onlytop.List() {
		if this.hidden.Has(x.(string)) {
			err = "@HIDDEN directive for '" + (x.(string)) + "' overrides @ONLYTOP."
			LOG.Warnln("File " + fname + ", line " + strconv.Itoa(fl.lineno()) + ":" + err)
		}
	}

	/*
		for k,v := range this.filemap {
			//println("FILEMAP ===== ",k," =====")
			for i := v.Front(); i != nil; i = i.Next() {
				//println(i.Value.(string))
			}
		}


		for k,v := range this.RulesMap {
			//println("===== " + k + " =====")
			for i := v.Front(); i != nil; i = i.Next() {
				//println(i.Value.(*Rule).getHead(), i.Value.(*Rule).getRight().Front().Value.(string))

			}
		}

		for k,v := range this.wild {
			//println("===== " + k + " =====")
			for i := v.Front(); i != nil; i = i.Next() {
				//println(i.Value.(*Rule).getHead(), i.Value.(*Rule).getRight().Front().Value.(string))

			}
		}
	*/
	TRACE(3, "Grammar loaded", MOD_GRAMMAR)
	return &this
}

func (this *Grammar) newRule(h string, ls *list.List, w bool, ngov int) {
	r := NewRuleFromString(h, ls, ngov)
	lr, exists := this.RulesMap[ls.Front().Value.(string)]
	if !exists {
		this.RulesMap[ls.Front().Value.(string)] = list.New()
		lr = this.RulesMap[ls.Front().Value.(string)]
	}
	lr.PushBack(r)

	if w {
		lw, exists := this.wild[ls.Front().Value.(string)[0:1]]
		if !exists {
			this.wild[ls.Front().Value.(string)[0:1]] = list.New()
			lw = this.wild[ls.Front().Value.(string)[0:1]]
		}

		lw.PushBack(r)
	}
}

func (this *Grammar) getSpecificity(s string) int {
	if strings.Index(s, "(") > -1 && strings.Index(s, ")") > -1 {
		return 0
	}

	if strings.Index(s, "<") > -1 && strings.Index(s, ">") > -1 {
		return 1
	}

	return 2
}

func (this *Grammar) getPriority(sym string) int {
	x, ok := this.prior[sym]
	if ok {
		return x
	} else {
		return 9999
	}
}

func (this *Grammar) isTerminal(sym string) bool {
	return !this.nonterminal.Has(sym)
}

func (this *Grammar) isHidden(sym string) bool {
	return this.hidden.Has(sym)
}

func (this *Grammar) isFlat(sym string) bool {
	return this.flat.Has(sym)
}

func (this *Grammar) isNoTop(sym string) bool {
	return this.notop.Has(sym)
}

func (this *Grammar) isOnlyTop(sym string) bool {
	return this.onlytop.Has(sym)
}

func (this *Grammar) getStartSymbol() string { return this.start }

func (this *Grammar) getRulesRight(cat string) *list.List {
	ls := this.RulesMap[cat]
	LOG.Tracef("GetRulesRight  for cat:%s", cat)
	if ls != nil {
		for l := ls.Front(); l != nil; l = l.Next() {
			LOG.Tracef("Head:%s right[%d] -> gov:%d", l.Value.(*Rule).getHead(), l.Value.(*Rule).getRight().Len(), l.Value.(*Rule).getGovernor())
		}

	}
	return If(ls != nil, ls, list.New()).(*list.List)
}

func (this *Grammar) getRulesRightWildcard(c string) *list.List {
	ls := this.wild[c]
	LOG.Tracef("GetRulesRightWildcard  for cat:%s", c)
	if ls != nil {
		for l := ls.Front(); l != nil; l = l.Next() {
			LOG.Tracef("Head:%s right[%d] -> gov:%d", l.Value.(*Rule).getHead(), l.Value.(*Rule).getRight().Len(), l.Value.(*Rule).getGovernor())
		}

	}
	return If(ls != nil, ls, list.New()).(*list.List)
}

func (this *Grammar) inFileMap(key string, val string) bool {
	i := this.filemap[key]
	if i == nil {
		return false
	}

	b := false
	for j := i.Front(); j != nil && !b; j = j.Next() {
		b = (j.Value.(string) == val)
	}
	return b
}

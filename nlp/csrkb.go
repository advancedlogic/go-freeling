package nlp

import (
	"container/list"
	"io/ioutil"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const VERTEX_NOT_FOUND = -1
const RE_WNP = "^[NARV]"

const (
	UKB_RELATION_FILE = 1 + iota
	UKB_REX_WNPOS
	UKB_PR_PARAMS
)

type CSRKB struct {
	maxIterations int
	threshold     float64
	damping       float64
	vertexIndex   map[string]int
	outCoef       []float64
	firstEdge     []int
	numEdges      []int
	edges         []int
	numVertices   int
}

type IntPair struct {
	first  int
	second int
}

type IntPairsArray []IntPair

func (a IntPairsArray) Less(i, j int) bool {
	return a[i].first < a[j].first || (a[i].first == a[j].first && a[i].second < a[j].second)
}

func (a IntPairsArray) Len() int { return len(a) }

func (a IntPairsArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func List2IntPairsArray(ls *list.List) []IntPair {
	out := make([]IntPair, ls.Len())
	for i, l := 0, ls.Front(); i < ls.Len() && l != nil; i, l = i+1, l.Next() {
		out[i] = l.Value.(IntPair)
	}
	return out
}

func IntPairsArray2List(a IntPairsArray) *list.List {
	out := list.New()
	for _, i := range a {
		out.PushBack(i)
	}
	return out
}

func NewCSRKB(kbFile string, nit int, thr float64, damp float64) *CSRKB {
	this := CSRKB{
		vertexIndex:   make(map[string]int),
		maxIterations: nit,
		threshold:     thr,
		damping:       damp,
	}

	var syn1, syn2 string
	var pos1, pos2 int
	rels := list.New()
	this.numVertices = 0

	fileString, err := ioutil.ReadFile(kbFile)
	if err != nil {
		LOG.Panic("Error loading file " + kbFile)
	}
	lines := strings.Split(string(fileString), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		items := Split(line, " ")

		syn1 = items[0]
		syn2 = items[1]

		pos1 = this.addVertex(syn1)
		if syn2 != "-" {
			pos2 = this.addVertex(syn2)

			rels.PushBack(IntPair{pos1, pos2})
			rels.PushBack(IntPair{pos2, pos1})
		}

	}

	this.fillCSRTables(this.numVertices, rels)
	return &this
}

func (this *CSRKB) fillCSRTables(nv int, rels *list.List) {
	tmpA := List2IntPairsArray(rels)
	sort.Sort(IntPairsArray(tmpA))
	rels = IntPairsArray2List(tmpA)

	this.edges = make([]int, rels.Len())
	this.firstEdge = make([]int, nv)
	this.numEdges = make([]int, nv)
	this.outCoef = make([]float64, nv)

	n := 0
	r := 0

	p := rels.Front()
	for p != nil && n < nv {
		this.firstEdge[n] = r
		for p != nil && p.Value.(IntPair).first == n {
			this.edges[r] = p.Value.(IntPair).second
			r++
			p = p.Next()
		}
		this.numEdges[n] = r - this.firstEdge[n]
		this.outCoef[n] = 1 / float64(this.numEdges[n])
		n++
	}
}

func (this *CSRKB) addVertex(s string) int {
	this.vertexIndex[s] = this.numVertices
	this.numVertices++
	return this.vertexIndex[s]
}

func (this *CSRKB) size() int { return this.numVertices }

func (this *CSRKB) getVertex(s string) int {
	out := this.vertexIndex[s]
	if out > 0 {
		return out
	} else {
		return VERTEX_NOT_FOUND
	}
}

func (this *CSRKB) pageRank(pv []float64) {
	var ranks [2][]float64
	CURRENT := 0
	NEXT := 1
	initVal := 1.0 / float64(this.numVertices)

	ranks[CURRENT] = ArrayFloatInit(this.numVertices, initVal)
	ranks[NEXT] = ArrayFloatInit(this.numVertices, 0.0)

	nit := 0
	change := this.threshold
	for nit < this.maxIterations && change >= this.threshold {
		change = 0

		for v := 0; v < this.numVertices; v++ {
			rank := 0.0
			for e := this.firstEdge[v]; e < this.firstEdge[v]+this.numEdges[v]; e++ {
				u := this.edges[e]
				rank += ranks[CURRENT][u] * this.outCoef[u]
			}

			ranks[NEXT][v] = rank*this.damping + pv[v]*(1-this.damping)
			change += math.Abs(ranks[NEXT][v] - ranks[CURRENT][v])
			//println(ranks[NEXT][v], change, rank, this.damping, pv[v])
		}

		tmp := NEXT
		NEXT = CURRENT
		CURRENT = tmp
		nit++
	}

	ArrayFloatSwap(pv, ranks[CURRENT])
}

type UKB struct {
	wn       *CSRKB
	RE_wnpos *regexp.Regexp
}

func NewUKB(wsdFile string) *UKB {
	this := UKB{
		RE_wnpos: regexp.MustCompile(RE_WNP),
	}

	path := wsdFile[0:strings.LastIndex(wsdFile, "/")]
	var relFile string

	var thr float64 = 0.000001
	var nit int = 30
	var damp float64 = 0.85

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("RelationFile", UKB_RELATION_FILE)
	cfg.AddSection("RE_Wordnet_PoS", UKB_REX_WNPOS)
	cfg.AddSection("PageRankParameters", UKB_PR_PARAMS)

	if !cfg.Open(wsdFile) {
		LOG.Panic("Error loading file " + wsdFile)
	}

	line := ""

	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
		case UKB_RELATION_FILE:
			{
				fname := items[0]
				if strings.HasPrefix(fname, "../") {
					wsdFile = strings.Replace(wsdFile, "./", "", -1)
					path = wsdFile[0:strings.Index(wsdFile, "/")]
					relFile = path + "/" + strings.Replace(fname, "../", "", -1)
				} else {
					relFile = path + "/" + strings.Replace(fname, "./", "", -1)
				}
				break
			}
		case UKB_REX_WNPOS:
			{
				this.RE_wnpos = regexp.MustCompile(line)
				break
			}
		case UKB_PR_PARAMS:
			{
				key := items[0]
				if key == "Threshold" {
					thr, _ = strconv.ParseFloat(items[1], 64)
				} else if key == "MaxIterations" {
					nit, _ = strconv.Atoi(items[1])
				} else if key == "Damping" {
					damp, _ = strconv.ParseFloat(items[1], 64)
				} else {
					LOG.Warn("Error: Unkown parameter " + key + " in PageRankParameters section in file " + wsdFile)
				}
				break
			}
		default:
			break
		}
	}

	if relFile == "" {
		LOG.Panic("No relation file provided in UKB configuration file " + wsdFile)
	}

	this.wn = NewCSRKB(relFile, nit, thr, damp)

	return &this
}

func (this *UKB) initSynsetVector(ls *list.List, pv []float64) {
	nw := 0
	uniq := make(map[string]*Word)
	for s := ls.Front(); s != nil; s = s.Next() {
		for w := s.Value.(*Sentence).Front(); w != nil; w = w.Next() {
			if this.RE_wnpos.MatchString(w.Value.(*Word).getTag(0)) {
				key := w.Value.(*Word).getLCForm() + "#" + strings.ToLower(w.Value.(*Word).getTag(0))[0:1]
				if uniq[key] == nil {
					nw++
					uniq[key] = w.Value.(*Word)
				}
			}
		}
	}

	for _, u := range uniq {
		lsen := u.getSenses(0)
		nsyn := lsen.Len()
		for s := lsen.Front(); s != nil; s = s.Next() {
			syn := this.wn.getVertex(s.Value.(FloatPair).first)
			if syn == VERTEX_NOT_FOUND {
				LOG.Warn("Unknown synset " + s.Value.(FloatPair).first + " ignored. Please check consistency between sense dictionary and KB")
			} else {
				pv[syn] += (1.0 / float64(nw)) * (1.0 / float64(nsyn))
			}
		}
	}
}

type FloatPair struct {
	first  string
	second float64
}

type FloatPairsArray []FloatPair

func (a FloatPairsArray) Less(i, j int) bool {
	return a[i].first < a[j].first || (a[i].first == a[j].first && a[i].second < a[j].second)
}

func (a FloatPairsArray) Len() int { return len(a) }

func (a FloatPairsArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func List2FloatPairsArray(ls *list.List) []FloatPair {
	out := make([]FloatPair, ls.Len())
	for i, l := 0, ls.Front(); i < ls.Len() && l != nil; i, l = i+1, l.Next() {
		out[i] = l.Value.(FloatPair)
	}
	return out
}

func FloatPairsArray2List(a FloatPairsArray) *list.List {
	out := list.New()
	for _, i := range a {
		out.PushBack(i)
	}
	return out
}

func (this *UKB) extractRanksToSentences(ls *list.List, pv []float64) {
	for s := ls.Front(); s != nil; s = s.Next() {
		for w := s.Value.(*Sentence).Front(); w != nil; w = w.Next() {
			lsen := w.Value.(*Word).getSenses(0)
			for p := lsen.Front(); p != nil; p = p.Next() {
				syn := this.wn.getVertex(p.Value.(FloatPair).first)
				if syn != VERTEX_NOT_FOUND {
					lsen.InsertAfter(FloatPair{p.Value.(FloatPair).first, pv[syn]}, p)
					lsen.Remove(p)
				}
			}

			a := List2FloatPairsArray(lsen)
			sort.Sort(FloatPairsArray(a))
			lsen = FloatPairsArray2List(a)
			w.Value.(*Word).setSenses(lsen, 0)
		}
	}
}

func (this *UKB) Analyze(ls *list.List) {
	pv := make([]float64, this.wn.size())
	this.initSynsetVector(ls, pv)
	this.wn.pageRank(pv)
	this.extractRanksToSentences(ls, pv)
}

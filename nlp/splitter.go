package nlp

import (
	"container/list"
	set "gopkg.in/fatih/set.v0"
	"strconv"
	"strings"
)

const SAME = 100
const VERY_LONG = 1000

const (
	SPLITTER_GENERAL = 1 + iota
	SPLITTER_MARKERS
	SPLITTER_SENT_END
	SPLITTER_SENT_START
)

type Splitter struct {
	SPLIT_AllowBetweenMarkers bool
	SPLIT_MaxWords            int64
	starters                  *set.Set
	enders                    map[string]bool
	markers                   map[string]int
}

func NewSplitter(splitterFile string) *Splitter {
	this := Splitter{
		starters: set.New(),
		enders:   make(map[string]bool),
		markers:  make(map[string]int),
	}

	cfg := NewConfigFile(false, "##")
	cfg.AddSection("General", SPLITTER_GENERAL)
	cfg.AddSection("Markers", SPLITTER_MARKERS)
	cfg.AddSection("SentenceEnd", SPLITTER_SENT_END)
	cfg.AddSection("SentenceStart", SPLITTER_SENT_START)

	if !cfg.Open(splitterFile) {
		CRASH("Error opening file "+splitterFile, MOD_SPLITTER)
	}

	this.SPLIT_AllowBetweenMarkers = true
	this.SPLIT_MaxWords = 0

	nmk := 1
	line := ""

	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
		case SPLITTER_GENERAL:
			{
				name := items[0]
				if name == "AllowBetweenMarkers" {
					this.SPLIT_AllowBetweenMarkers, _ = strconv.ParseBool(items[1])
				} else if name == "MaxWords" {
					this.SPLIT_MaxWords, _ = strconv.ParseInt(items[1], 10, 64)
				} else {
					LOG.Panic("Unexpected splitter option " + name)
				}
				break
			}
		case SPLITTER_MARKERS:
			{
				open := items[0]
				close := items[1]
				if open != close {
					this.markers[open] = nmk
					this.markers[close] = -nmk
				} else {
					this.markers[open] = SAME + nmk
					this.markers[close] = SAME + nmk
				}
				nmk++
				break
			}
		case SPLITTER_SENT_END:
			{
				name := items[0]
				value, _ := strconv.ParseBool(items[1])
				this.enders[name] = !value
				break
			}
		case SPLITTER_SENT_START:
			{
				this.starters.Add(line)
				break
			}
		default:
			break
		}
	}

	LOG.Trace("Analyzer succesfully created")
	return &this
}

type SplitterStatus struct {
	BetweenMark  bool
	NoSplitCount int
	MarkType     *list.List
	MarkForm     *list.List
	buffer       *Sentence
	nsentence    int
}

func (this *Splitter) OpenSession() *SplitterStatus {
	LOG.Trace("Opening new session")
	return &SplitterStatus{
		BetweenMark:  false,
		NoSplitCount: 0,
		MarkType:     list.New(),
		MarkForm:     list.New(),
		buffer:       NewSentence(),
		nsentence:    0,
	}
}

func (this *Splitter) CloseSession(ses *SplitterStatus) {
	LOG.Trace("Closing session")
	ses.MarkType = ses.MarkType.Init()
	ses.MarkForm = ses.MarkForm.Init()
	ses = nil
}

func (this *Splitter) Split(st *SplitterStatus, v *list.List, flush bool, ls *list.List) {
	ls = ls.Init()
	LOG.Trace("Looking for a sentence marker. Max no split is " + strconv.Itoa(int(this.SPLIT_MaxWords)))
	for w := v.Front(); w != nil; w = w.Next() {
		m := this.markers[w.Value.(*Word).getForm()]
		checkSplit := true

		if st.BetweenMark && !this.SPLIT_AllowBetweenMarkers && m != 0 && m == If(m > SAME, 1, -1).(int)*st.MarkType.Front().Value.(int) {
			LOG.Trace("End no split period. marker " + w.Value.(*Word).getForm() + " code: " + strconv.Itoa(m))
			st.MarkType.Remove(st.MarkType.Front())
			st.MarkForm.Remove(st.MarkForm.Front())
			if st.MarkForm.Len() == 0 {
				st.BetweenMark = false
				st.NoSplitCount = 0
			} else {
				st.NoSplitCount++
			}

			st.buffer.PushBack(w.Value.(*Word))
			checkSplit = false
		} else if m > 0 && !this.SPLIT_AllowBetweenMarkers {
			st.MarkForm.PushFront(w.Value.(*Word).getForm())
			st.MarkType.PushFront(m)
			LOG.Trace("Start no split periood, marker " + w.Value.(*Word).getForm() + " code:" + strconv.Itoa(m))
			st.BetweenMark = true
			st.NoSplitCount++
			st.buffer.PushBack(w.Value.(*Word))
			checkSplit = false
		} else if st.BetweenMark {
			LOG.Trace("no-split flag continues set. word " + w.Value.(*Word).getForm() + " expecting code " + strconv.Itoa(st.MarkType.Front().Value.(int)) + " (closing " + st.MarkForm.Front().Value.(string))
			st.NoSplitCount++
			if this.SPLIT_MaxWords == 0 || st.NoSplitCount <= int(this.SPLIT_MaxWords) {
				checkSplit = false
				st.buffer.PushBack(w.Value.(*Word))
			}

			if st.NoSplitCount == VERY_LONG {
				LOG.Warn("Sentence is very long")
			}
		}

		if checkSplit {
			e := this.enders[w.Value.(*Word).getForm()]
			if e {
				if e || this.endOfSentence(w, v) {
					LOG.Trace("Sentence marker [" + w.Value.(*Word).getForm() + "] found")
					st.buffer.PushBack(w.Value.(*Word))
					st.nsentence++
					st.buffer.sentID = strconv.Itoa(st.nsentence)
					ls.PushBack(st.buffer)
					LOG.Trace("Sentence lenght " + strconv.Itoa(st.buffer.Len()))
					nsentence := st.nsentence
					this.CloseSession(st)
					st = this.OpenSession()
					st.nsentence = nsentence
				} else {
					LOG.Trace(w.Value.(*Word).getForm() + " is not a sentence marker here")
					st.buffer.PushBack(w.Value.(*Word))
				}
			} else {
				LOG.Trace(w.Value.(*Word).getForm() + " is not a sentence marker here")
				st.buffer.PushBack(w.Value.(*Word))
			}
		}
	}

	if flush && st.buffer.Len() > 0 {
		LOG.Trace("Flushing the remaining words into a sentence")
		st.nsentence++
		st.buffer.sentID = strconv.Itoa(st.nsentence)
		ls.PushBack(st.buffer)
		nsentence := st.nsentence
		this.CloseSession(st)
		st = this.OpenSession()
		st.nsentence = nsentence
	}
}

func (this *Splitter) endOfSentence(w *list.Element, v *list.List) bool {
	if w == v.Back() {
		return true
	} else {
		r := w
		r = r.Next()
		f := r.Value.(*Word).getForm()

		return strings.Title(f) == f || this.starters.Has(f)
	}
}

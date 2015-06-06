package models

import (
	"container/list"
	"encoding/json"
	"time"

	uuid "github.com/nu7hatch/gouuid"
)

type DocumentEntity struct {
	id          string
	timestamp   int64
	Status      string
	Url         string `param:"url"`
	Title       string `param:"title"`
	Description string `param:"description"`
	Keywords    string `param:"keywords"`
	Content     string `param:"content"`
	TopImage    string
	Language    string   `param:"lang"`
	Flags       []string `param:flags`
	sentences   *list.List
	entities    map[string]int64
}

func NewDocumentEntity() *DocumentEntity {
	return &DocumentEntity{}
}

func (this *DocumentEntity) Init() {
	u4, _ := uuid.NewV4()
	this.id = u4.String()
	this.timestamp = time.Now().UnixNano()
	this.sentences = list.New()
	this.entities = make(map[string]int64)
	this.Status = ""
}

func (this *DocumentEntity) SexpString() string {
	js := this.ToJSON()
	sjs, _ := json.MarshalIndent(js, "", "\t")
	return string(sjs)
}

func (this *DocumentEntity) ToJSON() interface{} {
	js := make(map[string]interface{})
	if this.id != "" {
		js["id"] = this.id
	}

	if this.timestamp > 0 {
		js["timestamp"] = this.timestamp
	}

	if this.Url != "" {
		js["url"] = this.Url
	}

	if this.Title != "" {
		js["title"] = this.Title
	}

	if this.Description != "" {
		js["description"] = this.Description
	}

	if this.Keywords != "" {
		js["keywords"] = this.Keywords
	}

	if this.Content != "" {
		js["content"] = this.Content
	}

	if this.TopImage != "" {
		js["image"] = this.TopImage
	}

	if this.Status != "" {
		js["status"] = this.Status
	}

	if this.sentences.Len() > 0 {
		sentences := make([]interface{}, 0)
		for s := this.sentences.Front(); s != nil; s = s.Next() {
			sentence := s.Value.(*SentenceEntity)
			sentences = append(sentences, sentence.ToJSON())
		}
		js["sentences"] = sentences
	}

	if len(this.entities) > 0 {
		entities := make([]interface{}, 0)
		for name, frequency := range this.entities {
			entity := NewUnknownEntity(name, frequency)
			entities = append(entities, entity.ToJSON())
		}
		js["entities"] = entities
	}

	return js
}

func (this *DocumentEntity) Sentences() *list.List                { return this.sentences }
func (this *DocumentEntity) SetSentences(lss *list.List)          { this.sentences = lss }
func (this *DocumentEntity) AddSentenceEntity(se *SentenceEntity) { this.sentences.PushBack(se) }
func (this *DocumentEntity) AddUnknownEntity(name string, frequency int64) {
	this.entities[name] = frequency
}
func (this *DocumentEntity) String() string {
	return this.Url
}

const (
	CLASS_PERSON = 0 << iota
	CLASS_ORGANIZATION
	CLASS_PLACE
	CLASS_STOPWORD

	ROLE_SUBJECT = 0 << iota
	ROLE_ACTION
	ROLE_OBJECT
)

type TokenEntity struct {
	base   string
	lemma  string
	pos    string
	prob   float64
	class  int
	role   int
	weight float64
	sense  int
}

func NewTokenEntity(base string, lemma string, pos string, prob float64) *TokenEntity {
	return &TokenEntity{
		base:  base,
		lemma: lemma,
		pos:   pos,
		prob:  prob,
	}
}

func (this *TokenEntity) ToJSON() interface{} {
	js := make(map[string]interface{})
	js["base"] = this.base
	js["lemma"] = this.lemma
	js["pos"] = this.pos
	js["prob"] = this.prob
	return js
}

type SentenceEntity struct {
	body     string
	tokens   *list.List
	weight   float64
	sentence interface{}
	wdws     *list.List
}

func NewSentenceEntity() *SentenceEntity {
	return &SentenceEntity{
		tokens: list.New(),
		wdws:   list.New(),
	}
}

func (this *SentenceEntity) ToJSON() interface{} {
	js := make(map[string]interface{})
	if this.body != "" {
		js["body"] = this.body
	}
	if this.tokens.Len() > 0 {
		tokens := make([]interface{}, 0)
		for t := this.tokens.Front(); t != nil; t = t.Next() {
			te := t.Value.(*TokenEntity)
			tokens = append(tokens, te.ToJSON())
		}
		js["tokens"] = tokens
	}

	if this.wdws.Len() > 0 {
		wdws := make([]interface{}, 0)
		for w := this.wdws.Front(); w != nil; w = w.Next() {
			wdw := w.Value.(*WDWEntity)
			wdws = append(wdws, wdw.ToJSON())
		}
		js["wdws"] = wdws
	}
	return js
}

func (this *SentenceEntity) AddTokenEntity(te *TokenEntity) {
	this.tokens.PushBack(te)
}

func (this *SentenceEntity) AddWDWEntity(we *WDWEntity) {
	this.wdws.PushBack(we)
}

func (this *SentenceEntity) SetBody(body string)              { this.body = body }
func (this *SentenceEntity) SetSentence(sentence interface{}) { this.sentence = sentence }

func (this *SentenceEntity) GetSentence() interface{} { return this.sentence }

type WDWEntity struct {
	who  string
	did  string
	what string
}

func NewWDWEntity(who string, did string, what string) *WDWEntity {
	return &WDWEntity{
		who:  who,
		did:  did,
		what: what,
	}
}

func (this *WDWEntity) ToJSON() interface{} {
	js := make(map[string]string)
	if this.who != "" {
		js["who"] = this.who
	}
	if this.did != "" {
		js["did"] = this.did
	}
	if this.what != "" {
		js["what"] = this.what
	}
	return js
}

type UnknownEntity struct {
	name      string
	frequency int64
}

func NewUnknownEntity(name string, frequency int64) *UnknownEntity {
	return &UnknownEntity{
		name:      name,
		frequency: frequency,
	}
}

func (this *UnknownEntity) ToJSON() interface{} {
	js := make(map[string]interface{})
	js["name"] = this.name
	js["frequency"] = this.frequency
	return js
}

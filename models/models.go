package models

import (
	"container/list"
	"encoding/json"
	"fmt"
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
	Unknown     map[string]int64
	Entities    *list.List
}

func NewDocumentEntity() *DocumentEntity {
	return &DocumentEntity{}
}

func (this *DocumentEntity) Init() {
	u4, _ := uuid.NewV4()
	this.id = u4.String()
	this.timestamp = time.Now().UnixNano()
	this.sentences = list.New()
	this.Unknown = make(map[string]int64)
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

	if len(this.Unknown) > 0 {
		unknown := make([]interface{}, 0)
		for name, frequency := range this.Unknown {
			entity := NewUnknownEntity(name, frequency)
			unknown = append(unknown, entity.ToJSON())
		}
		js["unknown"] = unknown
	}

	if this.Entities.Len() > 0 {
		entities := make([]interface{}, 0)
		for e := this.Entities.Front(); e != nil; e = e.Next() {
			entity := e.Value.(*Entity)
			entities = append(entities, entity.ToJSON())
		}
		js["entities"] = entities
	}

	return js
}

func (this *DocumentEntity) Sentences() *list.List {
	return this.sentences
}
func (this *DocumentEntity) SetSentences(lss *list.List) {
	this.sentences = lss
}
func (this *DocumentEntity) AddSentenceEntity(se *SentenceEntity) {
	this.sentences.PushBack(se)
}
func (this *DocumentEntity) AddUnknownEntity(name string, frequency int64) {
	this.Unknown[name] = frequency
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
	base       string
	lemma      string
	pos        string
	prob       float64
	class      int
	role       int
	weight     float64
	sense      int
	annotation []*Annotation
}

type Annotation struct {
	Pos   string   `json:"pos"`
	Word  []string `json:"words"`
	Gloss string   `json:"glossary"`
}

func NewTokenEntity(base string, lemma string, pos string, prob float64, annotation []*Annotation) *TokenEntity {
	return &TokenEntity{
		base:       base,
		lemma:      lemma,
		pos:        pos,
		prob:       prob,
		annotation: annotation,
	}
}

func (this *TokenEntity) ToJSON() interface{} {

	js := make(map[string]interface{})
	js["base"] = this.base
	js["lemma"] = this.lemma
	js["pos"] = this.pos
	js["prob"] = this.prob
	js["annotation"] = this.annotation
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
	return js
}

func (this *SentenceEntity) AddTokenEntity(te *TokenEntity) {
	this.tokens.PushBack(te)
}

func (this *SentenceEntity) SetBody(body string) {
	this.body = body
}
func (this *SentenceEntity) SetSentence(sentence interface{}) {
	this.sentence = sentence
}

func (this *SentenceEntity) GetSentence() interface{} {
	return this.sentence
}

type Entity struct {
	model string
	score float64
	value string
}

func NewEntity(model string, score float64, value string) *Entity {
	return &Entity{
		model: model,
		score: score,
		value: value,
	}
}

func (this *Entity) String() string {
	return fmt.Sprintf("%s:%0.3f:%s", this.model, this.score, this.value)
}

func (this *Entity) ToJSON() interface{} {
	js := make(map[string]interface{})
	js["name"] = this.value
	js["type"] = this.model
	js["score"] = this.score
	return js
}

func (this *Entity) GetValue() string {
	return this.value
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

package lib

import (
	. "github.com/drankou/go-freeling/engine"
	"github.com/drankou/go-freeling/models"
)

type Analyzer struct {
	context *Context
}

func NewAnalyzer() *Analyzer {
	context := NewContext("conf/gofreeling.toml")
	context.InitNLP()
	instance := new(Analyzer)
	instance.context = context

	return instance
}

func (this *Analyzer) Int64(key string, def int64) int64 {
	return this.context.Int64(key, def)
}

func (this *Analyzer) AnalyzeText(document *models.DocumentEntity) *models.DocumentEntity {
	ch := make(chan *models.DocumentEntity)
	defer close(ch)

	go this.context.Engine.NLP.Workflow(document, ch)
	output := <-ch

	return output
}

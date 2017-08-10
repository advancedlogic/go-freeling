package nlp

import (
	goose "github.com/advancedlogic/GoOse"
)

type Crawler struct {
	name    string
	timeout int64
	loop    bool
	npages  int
	nlevels int
}

func NewDefaultCrawler() *Crawler {
	return &Crawler{
		name:    "default",
		timeout: 3000,
		loop:    false,
		npages:  0,
		nlevels: 0,
	}
}

func (this *Crawler) Analyze(url string) *goose.Article {
	g := goose.New()
	article, err := g.ExtractFromURL(url)
	if err != nil {
		// TODO Probably want to handle this error...
		panic(err)
	}
	return article
}

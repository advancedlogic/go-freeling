package nlp

import "container/list"

type Pattern struct {
	patr string
	head int
	tag  string
}

type Compound struct {
	unknownOnly bool
	patterns    *list.List
}

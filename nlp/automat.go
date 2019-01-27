package nlp

import (
	"github.com/fatih/set"
)

const AUTOMAT_MAX_STATES = 100
const AUTOMAT_MAX_TOKENS = 50

type AutomatStatus struct {
	shiftBegin int
}

type Automat struct {
	initialState int
	stopState    int
	trans        [AUTOMAT_MAX_STATES][AUTOMAT_MAX_TOKENS]int
	final        *set.Set
}

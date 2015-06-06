package nlp

import (
	"fmt"
	"log"
)

type traces struct {
	TraceLevel  int
	TraceModule int64
}

func (this traces) errorCrash(msg string, modcode int64) {
	panic(fmt.Sprintf("%s\n", msg))
}

func (this traces) warning(msg string, modcode int64) {
	log.Printf("[WARNING]: %s\n", msg)
}

func (this traces) trace(lv int, msg string, modcode int64) {
	if this.TraceLevel >= lv && this.TraceModule == modcode {
		log.Printf("[TRACE]: %s\n", msg)
	}
}

func CRASH(msg string, modname int64) {
	t := traces{}
	t.errorCrash(msg, modname)
}

func WARNING(msg string, modname int64) {
	t := traces{}
	t.warning(msg, modname)
}

func TRACE(lv int, msg string, modname int64) {
	t := traces{0, MOD_CHART}
	t.trace(lv, msg, modname)
}

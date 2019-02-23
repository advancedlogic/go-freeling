package nlp

import (
	"container/list"
	"regexp"
	"strings"
	"unicode"
)

func Split(s string, sep string) []string {
	items := strings.Split(s, sep)
	newItems := make([]string, 0)
	for _, item := range items {
		item = strings.Trim(item, " ")
		item = strings.Trim(item, "\t")
		if item != "" && item != " " && item != "\t" {
			newItems = append(newItems, item)
		}
	}
	tmp := strings.Join(newItems, "~")
	return strings.Split(tmp, "~")
}

func If(cond bool, trueRes interface{}, falseRes interface{}) interface{} {
	if cond {
		return trueRes
	} else {
		return falseRes
	}
}

func WhiteSpace(c uint8) bool {
	if c == 0x20 || c == 0xc || c == 0x0a || c == 0x0d || c == 0x09 || c == 0x0b || c == 0xc2 {
		return true
	}

	return false
}

func RegExHasSuffix(re *regexp.Regexp, s string) []string {
	if s == "" {
		return make([]string, 0)
	}
	re.Longest()
	outs := re.FindAllStringSubmatch(s, -1)
	newOuts := make([]string, 0)
	if len(outs) > 0 {
		inOuts := outs[0]
		ref := inOuts[0]
		for _, out := range inOuts {
			if out != "" && strings.HasPrefix(s, ref) && strings.HasPrefix(ref, out) {
				newOuts = append(newOuts, strings.TrimSpace(out))
			}
		}
	}
	return newOuts
}

const (
	UPPER_NONE = iota
	UPPER_ALL
	UPPER_1ST
)

func Capitalization(s string) int {
	caps := UPPER_NONE
	if strings.ToUpper(s) == s {
		caps = UPPER_ALL
	} else if strings.Title(s) == s {
		caps = UPPER_1ST
	}

	return caps
}

func Capitalize(form string, caps int, init bool) string {
	cl := form
	if caps == UPPER_ALL {
		cl = strings.ToUpper(cl)
	} else if caps == UPPER_1ST && init {
		cl = strings.Title(cl)
	}

	return cl
}

func List2Array(l *list.List) []interface{} {
	output := make([]interface{}, 0)
	for i := l.Front(); i != nil; i = i.Next() {
		output = append(output, i.Value)
	}
	return output
}

func StrList2StrArray(l *list.List) []string {
	output := make([]string, 0)
	for i := l.Front(); i != nil; i = i.Next() {
		output = append(output, i.Value.(string))
	}
	return output
}

func Array2List(a []interface{}) *list.List {
	output := list.New()
	for _, i := range a {
		output.PushBack(i)
	}

	return output
}

func StrArray2StrList(a []string) *list.List {
	output := list.New()
	for _, i := range a {
		output.PushBack(i)
	}

	return output
}

func EmptyFunc(i interface{}) {}

func Substr(s string, b int, l int) string {
	ln := len(s)
	if b > ln {
		return ""
	}
	if l == -1 || l >= ln {
		return s[b:]
	} else {
		return s[b : b+l]
	}
}

func IsCapitalized(s string) bool {
	s = strings.Split(s, " ")[0]
	if strings.Title(s) == s {
		return true
	}
	return false
}

func AllCaps(s string) bool {
	items := strings.Split(s, " ")
	if len(items) == 1 {
		return false
	}
	for _, item := range items {
		if strings.Title(item) != item {
			return false
		}
	}
	return true
}

func HasLowercase(s string) bool {
	for _, c := range s {
		if unicode.IsLetter(rune(c)) {
			return true
		}
	}

	return false
}

func MultiIndex(s string, i string) int {
	for _, c := range i {
		tmpPos := strings.Index(s, string(c))
		if tmpPos > -1 {
			return tmpPos
		}
	}
	return -1
}

func ListSwap(ls1 *list.List, ls2 *list.List) {
	tmpLs := list.New()
	for l1 := ls1.Front(); l1 != nil; l1 = l1.Next() {
		tmpLs.PushBack(l1)
	}
	ls1 = ls1.Init()
	for l2 := ls2.Front(); l2 != nil; l2 = l2.Next() {
		ls1.PushBack(l2)
	}
	ls2 = ls2.Init()
	for tl := tmpLs.Front(); tl != nil; tl = tl.Next() {
		ls2.PushBack(tl)
	}
}

func ArrayFloatSwap(a1 []float64, a2 []float64) {
	copy(a2, a1)
}

func ArrayFloatInit(l int, def float64) []float64 {
	a := make([]float64, l)
	for i, _ := range a {
		a[i] = def
	}

	return a
}

func StringsAppend(str ...string) string {
	l := 0
	for _, s := range str {
		l += len(s)
	}
	buffer := make([]byte, l)
	l = 0
	for _, s := range str {
		copy(buffer[l:l+len(s)], []byte(s))
		l += len(s)
	}
	return string(buffer)
}

func CreateStringWithChar(n int, c string) string {
	output := make([]byte, n)
	for i := 0; i < n; i++ {
		output[i] = c[0]
	}
	return string(output)
}

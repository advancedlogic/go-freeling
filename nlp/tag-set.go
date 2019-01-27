package nlp

import (
	"container/list"
	"fmt"
	"github.com/fatih/set"
	"strconv"
	"strings"
)

type TagSet struct {
	PAIR_SEP                         string
	MSD_SEP                          string
	feat, val, name, nameInv, valInv map[string]string
	direct                           map[string]*Pair
	directInv                        map[*set.Set]string
	shtagSize                        map[string]*list.List
}

const (
	DIRECT_TRANSLATIONS = 1 + iota
	DECOMPOSITION_RULES
)

func NewTagset(ftagset string) *TagSet {
	this := &TagSet{
		PAIR_SEP:  "=",
		MSD_SEP:   "|",
		feat:      make(map[string]string),
		val:       make(map[string]string),
		name:      make(map[string]string),
		nameInv:   make(map[string]string),
		direct:    make(map[string]*Pair),
		directInv: make(map[*set.Set]string),
		valInv:    make(map[string]string),
		shtagSize: make(map[string]*list.List),
	}
	cfg := NewConfigFile(false, "##")
	cfg.AddSection("DirectTranslations", DIRECT_TRANSLATIONS)
	cfg.AddSection("DecompositionRules", DECOMPOSITION_RULES)

	if !cfg.Open(ftagset) {
		CRASH("Error opening file "+ftagset, MOD_TAG_SET)
	}

	line := ""
	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.section {
		case DIRECT_TRANSLATIONS:
			{
				tag := items[0]
				shtag := items[1]
				msd := ""
				if len(items) > 2 {
					msd = items[2]
				}
				this.direct[tag] = &Pair{shtag, msd}
				s := set.New(set.ThreadSafe).(*set.Set)
				s.Add(msd, this.MSD_SEP)
				this.directInv[s] = tag
				break
			}
		case DECOMPOSITION_RULES:
			{
				cat := items[0]
				shsz := items[1]
				if len(items) > 2 {
					pos := items[2]
					this.name[cat] = pos
					this.nameInv[pos] = cat
				}
				this.shtagSize[cat] = list.New()
				tokens := strings.Split(shsz, ",")
				for _, sitem := range tokens {
					item, _ := strconv.Atoi(sitem)
					this.shtagSize[cat].PushBack(item)
				}
				//TRACE(3, fmt.Sprintf("Read short tag size for %s (%s) %s\n", cat, pos, shsz), MOD_TAG_SET)
				i := 1
				if len(items) > 4 {
					msd := items[4]
					key := cat + "#" + strconv.Itoa(i)
					k := strings.Split(msd, "/")
					this.feat[key] = k[0]
					this.feat[cat+"#"+k[0]] = strconv.Itoa(i)
					v := strings.Split(k[1], ";")
					for j := 0; j < len(v); j++ {
						t := strings.Split(v[j], ":")
						this.val[key+"#"+strings.ToUpper(t[0])] = t[1]
						this.valInv[key+"#"+t[1]] = strings.ToUpper(t[0])
					}

					i++
				}
				break
			}
		default:
			break
		}
	}

	TRACE(1, "Module created successfully", MOD_HMM)

	return this
}

func (this TagSet) GetShortTag(tag string) string {
	TRACE(3, fmt.Sprintf("get short tag for %s\n", tag), MOD_TAG_SET)
	p := this.direct[tag]
	if p != nil {
		TRACE(5, fmt.Sprintf("   Found direct entry %s\n", p.first.(string)), MOD_TAG_SET)
		return p.first.(string)
	} else {
		s := this.shtagSize[tag[0:1]]
		if s != nil {
			if s.Len() == 1 {
				ln := s.Front()
				TRACE(5, fmt.Sprintf("   cutting 1st position sz=%s\n", strconv.Itoa(ln.Value.(int))), MOD_TAG_SET)
				if ln.Value.(int) == 0 {
					return tag
				} else {
					return tag[0:ln.Value.(int)]
				}
			} else {
				shtg := ""
				for k := s.Front(); k != nil; k = k.Next() {
					if k.Value.(int) < len(tag) {
						shtg += string(tag[k.Value.(int)])
					} else {
						WARNING(fmt.Sprintf("Tag %s too short for requested digits. Unchanged", tag), MOD_TAG_SET)
						return tag
					}
				}
				TRACE(5, "   Extracting digits "+shtg+" from "+tag, MOD_HMM)
				return shtg
			}
		}
	}
	WARNING("No rule to get short version of tag '"+tag+"'.", MOD_TAG_SET)
	return tag
}

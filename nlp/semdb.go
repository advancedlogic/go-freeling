package nlp

import (
	"container/list"
	set "gopkg.in/fatih/set.v0"
	"io/ioutil"
	"strings"
)

const (
	SEMDB_WN_POS_MAP = 1 + iota
	SEMDB_DATA_FILES
)

type SenseInfo struct {
	sense   string
	parents *list.List
	semFile string
	words   *list.List
	tonto   *list.List
	sumo    string
	cyc     string
}

func NewSenseInfo(syn string, data string) *SenseInfo {
	this := SenseInfo{
		sense: syn,
	}

	if data != "" {
		fields := StrArray2StrList(Split(data, " "))
		f := fields.Front()
		if string(f.Value.(string)[0]) != "-" {
			this.parents = StrArray2StrList(Split(f.Value.(string), ":"))
		}
		f = f.Next()
		this.semFile = f.Value.(string)
		f = f.Next()
		if string(f.Value.(string)[0]) != "-" {
			this.tonto = StrArray2StrList(Split(f.Value.(string), ":"))
		}
		f = f.Next()
		if string(f.Value.(string)[0]) != "-" {
			this.sumo = f.Value.(string)
		}
		f = f.Next()
		if string(f.Value.(string)[0]) != "-" {
			this.cyc = f.Value.(string)
		}
	}

	return &this
}

func (this *SenseInfo) getParentsString() string {
	out := make([]string, this.parents.Len())
	for i, p := 0, this.parents.Front(); i < this.parents.Len() && p != nil; i, p = i+1, p.Next() {
		out[i] = p.Value.(string)
	}

	return strings.Join(out, ":")
}

type PosMapRule struct {
	pos   string
	wnpos string
	lemma string
}

type SemanticDB struct {
	posMap   *list.List
	formDict *Database
	senseDB  *Database
	wndb     *Database
}

func NewSemanticDB(wsdFile string) *SemanticDB {
	this := SemanticDB{
		posMap: list.New(),
	}
	var formFile, dictFile, wnFile string

	path := wsdFile[0:strings.LastIndex(wsdFile, "/")]

	posset := set.New()
	cfg := NewConfigFile(true, "")
	cfg.AddSection("WNposMap", SEMDB_WN_POS_MAP)
	cfg.AddSection("DataFiles", SEMDB_DATA_FILES)

	if !cfg.Open(wsdFile) {
		LOG.Panic("Error opening configuration file " + wsdFile)
	}

	line := ""

	for cfg.GetContentLine(&line) {
		items := Split(line, " ")
		switch cfg.GetSection() {
		case SEMDB_WN_POS_MAP:
			{
				r := PosMapRule{}
				r.pos = items[0]
				r.wnpos = items[1]
				r.lemma = items[2]
				this.posMap.PushBack(r)
				if r.lemma != "L" && r.lemma != "F" {
					posset.Add(r.lemma)
				}
				break
			}
		case SEMDB_DATA_FILES:
			{
				key := items[0]
				fname := items[1]
				if key == "formDictFile" {
					formFile = path + "/" + strings.Replace(fname, "./", "", -1)
				} else if key == "senseDictFile" {
					dictFile = path + "/" + strings.Replace(fname, "./", "", -1)
				} else if key == "wnFile" {
					wnFile = path + "/" + strings.Replace(fname, "./", "", -1)
				}
				break
			}
		default:
			break
		}
	}

	if formFile == "" || posset.Size() == 0 {
		this.formDict = nil
	} else {
		fileString, err := ioutil.ReadFile(formFile)
		if err != nil {
			LOG.Panic("Error loading file " + formFile)
		}
		lines := strings.Split(string(fileString), "\n")
		this.formDict = NewDatabase(DB_MAP)
		for _, line := range lines {
			items := Split(line, " ")
			form := items[0]
			for i := 1; i < len(items); i = i + 2 {
				lemma := items[i]
				tag := items[i+1]
				if posset.Has(tag) {
					this.formDict.addDatabase(lemma+" "+tag, form)
				}
			}
		}
	}

	if dictFile == "" {
		this.senseDB = nil
	} else {
		fileString, err := ioutil.ReadFile(dictFile)
		if err != nil {
			LOG.Panic("Error loading file " + dictFile)
		}
		lines := strings.Split(string(fileString), "\n")
		this.senseDB = NewDatabase(DB_MAP)
		for _, line := range lines {
			items := Split(line, " ")
			sens := items[0]
			tag := sens[strings.Index(sens, "-")+1:]
			for i := 1; i < len(items); i++ {
				wd := items[i]
				this.senseDB.addDatabase("S:"+sens, wd)
				this.senseDB.addDatabase("W:"+wd+":"+tag, sens)
			}
		}
	}

	if wnFile == "" {
		this.wndb = nil
	} else {
		this.wndb = NewDatabaseFromFile(wnFile)
	}

	return &this
}

func (this *SemanticDB) getWordSenses(form string, lemma string, pos string) *list.List {
	searchList := list.New()
	this.getWNKeys(form, lemma, pos, searchList)

	lsen := list.New()
	for p := searchList.Front(); p != nil; p = p.Next() {
		LOG.Trace("..searching " + p.Value.(Pair).first.(string) + " " + p.Value.(Pair).second.(string))
		s := StrArray2StrList(Split(this.senseDB.accessDatabase("W:"+p.Value.(Pair).first.(string)+":"+p.Value.(Pair).second.(string)), " "))
		for ss := s.Front(); ss != nil; ss = ss.Next() {
			lsen.PushBack(ss.Value.(string))
		}

		if lsen.Len() > 0 {
			LOG.Trace("..senses found " + strings.Join(StrList2StrArray(lsen), " "))
		}
	}

	return lsen
}

func (this *SemanticDB) getSenseWords(sens string) *list.List {
	return StrArray2StrList(Split(this.senseDB.accessDatabase("S:"+sens), " "))
}

func (this *SemanticDB) getSenseInfo(syn string) *SenseInfo {
	sinf := NewSenseInfo(syn, this.wndb.accessDatabase(syn))
	sinf.words = this.getSenseWords(syn)
	return sinf
}

func (this *SemanticDB) getWNKeys(form string, lemma string, tag string, searchList *list.List) {
	searchList = searchList.Init()
	for p := this.posMap.Front(); p != nil; p = p.Next() {
		LOG.Trace("Check tag " + tag + " with posmap " + p.Value.(PosMapRule).pos + " " + p.Value.(PosMapRule).wnpos + " " + p.Value.(PosMapRule).lemma)
		if strings.Index(tag, p.Value.(PosMapRule).pos) == 0 {
			LOG.Trace("   matched")
			var lm string
			if p.Value.(PosMapRule).lemma == "L" {
				lm = lemma
			} else if p.Value.(PosMapRule).lemma == "F" {
				lm = form
			} else {
				LOG.Trace("FOund word matching special map: " + lemma + " " + p.Value.(PosMapRule).lemma)
				lm = this.formDict.accessDatabase(lemma + " " + p.Value.(PosMapRule).lemma)
			}

			fms := StrArray2StrList(Split(lm, " "))
			for ifm := fms.Front(); ifm != nil; ifm = ifm.Next() {
				LOG.Trace("Adding word '" + form + "' to be searched with pos=" + p.Value.(PosMapRule).pos + " and lemma=" + ifm.Value.(string))
				searchList.PushBack(Pair{ifm.Value.(string), p.Value.(PosMapRule).wnpos})
			}
		}
	}
}

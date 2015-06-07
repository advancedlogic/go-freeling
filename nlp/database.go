package nlp

import (
	cache "github.com/pmylund/go-cache"
	"io/ioutil"
	"strings"
)

const (
	DB_MAP = iota
	DB_PREFTREE
)

type Database struct {
	DBType  int
	dbmap   map[string]string
	dbptree *cache.Cache
}

func NewDatabase(t int) *Database {
	this := Database{
		DBType: t,
	}

	if t == DB_MAP {
		this.dbmap = make(map[string]string)
	} else if t == DB_PREFTREE {
		this.dbptree = cache.New(0, 0)
	}
	return &this
}

func NewDatabaseFromFile(dbFile string) *Database {
	this := Database{
		DBType: DB_MAP,
		dbmap:  make(map[string]string),
	}

	if dbFile != "" {
		filestr, err := ioutil.ReadFile(dbFile)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(filestr), "\n")
		if lines[0] == "DB_PREFTREE" {
			this.DBType = DB_PREFTREE
		}

		for i := 1; i < len(lines); i++ {
			line := lines[i]
			if line != "" {
				pos := strings.Index(line, " ")
				key := line[0:pos]
				data := line[pos+1:]
				this.addDatabase(key, data)
			}
		}
	}

	return &this
}

func (this *Database) addDatabase(key string, data string) {
	if this.DBType == DB_MAP {
		p := this.dbmap[key]
		if p != "" {
			this.dbmap[key] = p + " " + data
		} else {
			this.dbmap[key] = data
		}
	} else {
		_, found := this.dbptree.Get(key)
		if found {

		}
	}
}

func (this *Database) accessDatabase(key string) string {
	switch this.DBType {
	case DB_MAP:
		{
			p := this.dbmap[key]
			if p != "" {
				return p
			}
			break
		}
	case DB_PREFTREE:
		{
			//TODO
			break
		}
	default:
		break
	}

	return ""
}

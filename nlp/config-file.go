package nlp

import (
	"io/ioutil"
	"strings"
)

type ConfigFile struct {
	lines                       []string
	sectionsOpen, sectionsClose map[string]int
	section                     int
	SECTION_NONE                int
	SECTION_UNKNOWN             int
	filename                    string
	sectionStart                bool
	commentPrefix               string
	lineNum                     int
	skipUnknownSections         bool
	unkName                     string
}

func NewConfigFile(skip bool, comment string) ConfigFile {
	if comment == "" {
		comment = "##"
	}
	return ConfigFile{
		SECTION_NONE:        -1,
		SECTION_UNKNOWN:     -2,
		skipUnknownSections: skip,
		commentPrefix:       comment,
		sectionsOpen:        make(map[string]int),
		sectionsClose:       make(map[string]int),
	}
}

func (this *ConfigFile) IsOpenSection(s string) bool {
	return len(s) > 2 && s[0] == '<' && s[1] != '/' && s[len(s)-1] == '>'
}

func (this *ConfigFile) IsCloseSection(s string) bool {
	return len(s) > 3 && s[0] == '<' && s[1] == '/' && s[len(s)-1] == '>'
}

func (this *ConfigFile) IsComment(s string) bool {
	return s == "" || strings.Contains(s, this.commentPrefix)
}

func (this *ConfigFile) AddSection(key string, section int) {
	this.sectionsOpen["<"+key+">"] = section
	this.sectionsClose["</"+key+">"] = section
}

func (this *ConfigFile) PrintSections() {
	for k, v := range this.sectionsOpen {
		println(k, v)
	}
}

func (this *ConfigFile) Open(filename string) bool {
	this.filename = filename
	this.section = this.SECTION_NONE
	fileString, err := ioutil.ReadFile(filename)
	if err != nil {
		return false
	}
	lines := strings.Split(string(fileString), "\n")
	this.lines = make([]string, len(lines))
	copy(this.lines, lines)
	this.lineNum = -1
	return true
}

func (this *ConfigFile) GetSection() int {
	return this.section
}

func (this *ConfigFile) GetLineNum() int {
	return this.lineNum
}

func (this *ConfigFile) AtSectionStart() bool {
	return this.sectionStart
}

func (this *ConfigFile) GetContentLine(line *string) bool {
	this.lineNum++
	this.sectionStart = false
	for k := this.lineNum; k < len(this.lines); k++ {
		*line = string(this.lines[k])
		if this.section == this.SECTION_NONE {
			if this.IsOpenSection(*line) {
				section := this.sectionsOpen[*line]
				if section == 0 {
					if !this.skipUnknownSections {
						LOG.Panic("Opening of unknown section " + *line + " in file " + this.filename)
					} else {
						this.section = this.SECTION_UNKNOWN
						this.unkName = (*line)[1 : len(*line)-1]
						LOG.Tracef("Entering unknown section %s in file %s", *line, this.filename)
					}
				} else {
					this.section = section
					this.sectionStart = true
					LOG.Tracef("Entering section %s in file %s", *line, this.filename)
				}
			} else if this.IsCloseSection(*line) {
				LOG.Error("Unexpected closing of section " + *line + " in file " + this.filename)
			} else if !this.IsComment(*line) {
				LOG.Warnf("Ignoring unexpected non-comment line outside sections %s in file %s\n", *line, this.filename)
			} else {
				LOG.Tracef("Skipping comment %s", *line)
			}
		} else if this.section != this.SECTION_NONE {
			if this.IsCloseSection(*line) {
				s := this.sectionsClose[*line]
				if s == 0 {
					if !this.skipUnknownSections {
						LOG.Panic("Closing of unknown section " + *line + " in file " + this.filename)
					} else if this.section == this.SECTION_UNKNOWN {
						if this.unkName != (*line)[2:len(*line)-1] {
							LOG.Panic("Unexpected closing of unknown section " + *line + " in file " + this.filename)
						} else {
							LOG.Tracef("Exiting unknown section %s in file %s", *line, this.filename)
							this.section = this.SECTION_NONE
						}
					} else {
						LOG.Panic("Unexpected section closing " + *line + " in file " + this.filename)
					}
				} else if s != this.section {
					LOG.Panic("Unexpected closing in section " + *line + " in file " + this.filename)
				} else {
					LOG.Tracef("Exiting section %s in file %s", *line, this.filename)
					this.section = this.SECTION_NONE
				}
			} else if this.IsOpenSection(*line) {
				LOG.Panic("Unexpected nested opening of section " + *line + " in file " + this.filename)
			} else if this.section != this.SECTION_UNKNOWN && !this.IsComment(*line) {
				return true
			}
		}
		this.lineNum++
	}

	return this.lineNum < len(this.lines)
}

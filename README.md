# go-freeling

**Natural Language Processing** in GO

This is a partial port of Freeling 3.1 (http://nlp.lsi.upc.edu/freeling/).

License is GPL to respect the License model of Freeling.

This is the list of features already implemented:

* Text tokenization
* Sentence splitting
* Morphological analysis
* Suffix treatment, retokenization of clitic pronouns
* Flexible multiword recognition
* Contraction splitting
* Probabilistic prediction of unknown word categories
* Named entity detection
* PoS tagging
* Chart-based shallow parsing
* Named entity classification (With an external library MITIE - https://github.com/mit-nlp/MITIE)
* Rule-based dependency parsing

-

**How to use it**:

<pre>
go build gofreeling.go

./gofreeling
</pre>

(http server listens on default port 9999 - port can be changed in conf/gofreeling.toml file)

To process a page:

HTTP GET: *http://localhost:9999/analyzer?url=COPY HERE AN URL*

or **Use as API endpoint:**
<pre>
HTTP POST:

http://localhost:9999/analyzer-api

{
    content: 'Text you want to analyze'
}
</pre>

*Response is a self-explaining json*

**Usage as package:**
(*example*)
<pre>
package main

import (
	. "./lib"
	. "./models"
	"fmt"
	"encoding/json"
)

func main() {
	document := new(DocumentEntity)
	analyzer := NewAnalyzer()
	document.Content = "Hello World"
	output := analyzer.AnalyzeText(document)
	
	js := output.ToJSON()
	b, err := json.Marshal(js)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

</pre>

-
TODO:
* clean code
* add comments
* add tests
* ~~implement WordNet-based sense annotation and disambiguation~~

-
**Linguistic Data** to run the server can be download here (English only):

https://www.dropbox.com/s/fwwvfxp2s7dydet/data.zip


**WordNet Database** to add annotation (place it inside `./data` folder)

http://wordnetcode.princeton.edu/3.0/WNdb-3.0.tar.gz
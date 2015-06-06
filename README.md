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

-

**How to use it**:

go build gofreeling.go

./gofreeling

(http server listens on default port 9999 - port can be changed in conf/gofreeling.toml file)

To process a page:

HTTP GET: *http://localhost:9999/analyzer?url=COPY HERE AN URL*

Response is a self-explaining json

-
TODO:
* add NER to json
* clean code
* add comments
* add tests
* implement WordNet-based sense annotation and disambiguation
* implement Rule-based dependency parsing

-
**Linguistic Data** to run the server can be download here (English only):

https://www.dropbox.com/s/fwwvfxp2s7dydet/data.zip
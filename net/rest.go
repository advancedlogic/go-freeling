package net

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	. "../lib"
	"../models"
	. "../terminal"
)

type reqBody struct {
	Content string
}

type HttpServer struct {
	router   *mux.Router
	analyzer *Analyzer
}

func NewHttpServer(analyzer *Analyzer) *HttpServer {
	instance := new(HttpServer)
	instance.router = mux.NewRouter()
	instance.analyzer = analyzer

	return instance
}

func (this *HttpServer) Listen() {
	this.router.HandleFunc("/analyzer", this.URLHandler)
	this.router.HandleFunc("/analyzer-api", this.APIHandler)
	this.router.HandleFunc("/ping", this.PingHandler)

	port := this.analyzer.Int64("http.port", 9999)
	Infof("Http Server listening on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), this.router)
}

func (this *HttpServer) APIHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var body reqBody
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	document := new(models.DocumentEntity)
	document.Content = body.Content

	this.DocumentHandler(document, w)
}

func (this *HttpServer) URLHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	url := params.Get("url")
	document := new(models.DocumentEntity)
	document.Url = url

	this.DocumentHandler(document, w)
}

func (this *HttpServer) DocumentHandler(document *models.DocumentEntity, w http.ResponseWriter) {
	output := this.analyzer.AnalyzeText(document)

	js := output.ToJSON()
	b, err := json.Marshal(js)

	if err != nil {
		w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
	} else {
		w.Write([]byte(fmt.Sprintf("%s\n", string(b))))
	}
}

func (this *HttpServer) PingHandler(w http.ResponseWriter, r *http.Request) {
	Infoln("pong")
	w.Write([]byte("pong"))
}

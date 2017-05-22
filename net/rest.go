package net

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	. "../engine"
	"../models"
	. "../terminal"
)

type HttpServer struct {
	router  *mux.Router
	context *Context
}

func NewHttpServer(context *Context) *HttpServer {
	instance := new(HttpServer)
	instance.router = mux.NewRouter()
	instance.context = context

	return instance
}

func (this *HttpServer) Listen() {
	this.router.HandleFunc("/analyzer", this.URLHandler)
	this.router.HandleFunc("/ping", this.PingHandler)

	port := this.context.Int64("http.port", 9999)
	Infof("Http Server listening on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), this.router)
}

func (this *HttpServer) URLHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	url := params.Get("url")
	document := new(models.DocumentEntity)
	document.Url = url

	ch := make(chan *models.DocumentEntity)
	defer close(ch)

	go this.context.Engine.NLP.Workflow(document, ch)
	output := <-ch

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

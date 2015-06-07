package main

import (
	"github.com/advancedlogic/go-freeling/engine"
	"github.com/advancedlogic/go-freeling/net"
	"github.com/advancedlogic/go-freeling/terminal"
)

var logo = `
   ____             _____              _ _             
  / ___| ___       |  ___| __ ___  ___| (_)_ __   __ _ 
 | |  _ / _ \ _____| |_ | '__/ _ \/ _ \ | | '_ \ / _  |
 | |_| | (_) |_____|  _|| | |  __/  __/ | | | | | (_| |
  \____|\___/      |_|  |_|  \___|\___|_|_|_| |_|\__, |
                                                 |___/ 
			AdvancedLogic 2015 - v.0.1
`

func init() {
	Infof("Go - Freeling - Natural Language Processing for Golang\n")
	Infof("This is a partial port of Freeling 3.1\n")
}

func main() {
	context := NewContext("conf/gofreeling.toml")
	context.InitNLP()

	println(logo)

	httpServer := NewHttpServer(context)
	httpServer.Listen()
}

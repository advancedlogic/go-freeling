package terminal

import (
	"github.com/fatih/color"
	"strings"
)

var white = color.New(color.FgWhite)

func Default(messages ...string) {
	message := strings.Join(messages, " ")
	white.Print(message)
}

func Defaultln(messages ...string) {
	Default(messages...)
	print("\n")
}

func Defaultf(message string, params ...interface{}) {
	white.Printf(message, params...)
}

var green = color.New(color.FgGreen)

func Info(messages ...string) {
	message := strings.Join(messages, " ")
	green.Print(message)
}

func Infoln(messages ...string) {
	Info(messages...)
	print("\n")
}

func Infof(message string, params ...interface{}) {
	green.Printf(message, params...)
}

var yellow = color.New(color.FgYellow)

func Output(messages ...string) {
	message := strings.Join(messages, " ")
	yellow.Print(message)
}

func Outputln(messages ...string) {
	Output(messages...)
	print("\n")
}

func Outputf(message string, params ...interface{}) {
	yellow.Printf(message, params...)
}

var red = color.New(color.FgRed)

func Error(messages ...string) {
	message := strings.Join(messages, " ")
	red.Print(message)
}

func Errorln(messages ...string) {
	Error(messages...)
	print("\n")
}

func Errorf(message string, params ...interface{}) {
	red.Printf(message, params...)
}

func ErrorI(messages ...interface{}) {
	red.Print(messages...)
}

func ErrorIln(messages ...interface{}) {
	red.Println(messages...)
}

package main

import (
	"fmt"
	"html/template"
	"os"
	"github.com/bukind/webtests/filefinder"
)

var (
	ff = filefinder.New(os.ExpandEnv("${GOPATH}/src/github.com/bukind/webtests/03websock"), "03websock", ".")
	helloTmpl = template.Must(template.ParseFiles(ff.Must("templates/hello.html")...))
)

type Page struct {
	Title string
}

func main() {
	page := Page{
		Title: "Hello, world",
	}
	err := helloTmpl.Execute(os.Stdout, page)
	fmt.Println("err=",err)
}

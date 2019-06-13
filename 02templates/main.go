package main

import (
	"fmt"
	"html/template"
	"os"
)

var (
	baseHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
<title>{{block "title" .}}{{end}}</title>{{block "style" .}}{{end}}
</head>
<body>{{template "content"}}
</body>{{block "js" .}}{{end}}
</html>
`
	helloTmpl = template.Must(template.New("hello").Parse(baseHTML+`{{define "title"}}{{.Title}}{{end}}{{define "content"}}
<p>Hello, world</p>{{end}}`))
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

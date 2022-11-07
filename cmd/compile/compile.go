package main

import (
	htmlTemplate "html/template"
	"log"
	"os"
)

func main() {

	htmlTpl := htmlTemplate.Must(htmlTemplate.ParseGlob("*.tpl.html"))
	for _, tpl := range htmlTpl.Templates() {
		log.Printf("template: %s", tpl.Name())
		tpl.Execute(os.Stdout, nil)
	}
}

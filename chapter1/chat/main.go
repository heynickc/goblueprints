package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/heynickc/goblueprints/chapter1/trace"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	// adding the r as a param
	// tells the template to render using data
	// that can be extracted from the http.Request
	// which includes the server address
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse() // parse the flags
	r := newRoom()
	r.tracer = trace.New(os.Stdout)
	// root
	http.Handle("/", &templateHandler{filename: "chat.html"})
	// this is where go r.run() blocks the ListenAndServe method
	http.Handle("/room", r)
	// get the room going
	go r.run()
	// start the web server
	log.Println("Starting the web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

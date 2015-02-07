package main

import (
	"fmt"
	_ "html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	txttemplate "text/template"

	"github.com/russross/blackfriday"
	fsnotify "gopkg.in/fsnotify.v1"
)

var SUFFIXES = [3]string{".md", ".mkd", ".markdown"}

var toc []string
var toc_mutex sync.Mutex
var rootTmpl *txttemplate.Template
var path string

func init() {
	var err error
	rootTmpl, err = txttemplate.New("root").Parse(rootTemplate)
	if err != nil {
		log.Fatal(err)
	}
}

func HasMarkdownSuffix(s string) bool {
	for _, suffix := range SUFFIXES {
		if strings.HasSuffix(strings.ToLower(s), suffix) {
			return true
		}
	}
	return false
}

func AddWatch(w *fsnotify.Watcher) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			w.Add(path)
		} else {
			if HasMarkdownSuffix(path) {
				toc_mutex.Lock()
				toc = append(toc, path)
				toc_mutex.Unlock()
				log.Println("Found", path)
				w.Add(path)
			}
		}
		return nil
	}
}

func WatcherEventLoop(w *fsnotify.Watcher, done chan bool) {
	for {
		select {
		case event := <-w.Events:
			log.Println("Event:", event)
			// TODO(barakmich): On directory creation, stat path if directory, and watch it.
			if HasMarkdownSuffix(event.Name) {
			}

		case err := <-w.Errors:
			log.Println("Error:", err)
			done <- true
		}
	}
}

func RootFunc(w http.ResponseWriter, r *http.Request) {
	toc_mutex.Lock()
	localToc := toc[:]
	toc_mutex.Unlock()
	for i, s := range localToc {
		s = strings.TrimPrefix(s, path)
		localToc[i] = "* " + s
	}
	tocmkd := strings.Join(localToc, "\n")
	bytes := blackfriday.MarkdownCommon([]byte(tocmkd))
	rootTmpl.Execute(w, string(bytes))
}

func main() {
	path = os.Getenv("PWD")
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go WatcherEventLoop(watcher, done)

	log.Println("Watching directory", path)
	err = filepath.Walk(path, AddWatch(watcher))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(toc)

	fmt.Println(blackfriday.HTML_TOC)
	http.HandleFunc("/", RootFunc)
	http.ListenAndServe(":8080", nil)
}

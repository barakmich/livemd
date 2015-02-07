package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/russross/blackfriday"
	"golang.org/x/net/websocket"
	fsnotify "gopkg.in/fsnotify.v1"
)

var SUFFIXES = [3]string{".md", ".mkd", ".markdown"}

var toc []string
var tocMutex sync.Mutex
var rootTmpl *template.Template
var pageTmpl *template.Template
var path string

type state int

const (
	None state = iota
	Open
	Close
)

type Listener struct {
	File   string
	Socket *websocket.Conn
	State  state
}

type Update struct {
	File string
}

type BrowserMsg struct {
	Markdown string
}

func init() {
	var err error
	rootTmpl, err = template.New("root").Parse(rootTemplate)
	if err != nil {
		log.Fatal(err)
	}
	pageTmpl, err = template.New("page").Parse(pageTemplate)
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
				tocMutex.Lock()
				toc = append(toc, path)
				tocMutex.Unlock()
				log.Println("Found", path)
				w.Add(path)
			}
		}
		return nil
	}
}

func WatcherEventLoop(w *fsnotify.Watcher, updates chan Update, done chan bool) {
	for {
		select {
		case event := <-w.Events:
			//			log.Println("Event:", event)
			// TODO(barakmich): On directory creation, stat path if directory, and watch it.
			if HasMarkdownSuffix(event.Name) {
				subfile := strings.TrimPrefix(event.Name, path)
				if event.Op == fsnotify.Write {
					updates <- Update{subfile}
				}
			}
		case err := <-w.Errors:
			log.Println("Error:", err)
			done <- true
		}
	}
}

func writeFileForListener(l Listener) {
	var data []byte
	file, err := os.Open(filepath.Join(path, l.File))
	if err != nil {
		data = []byte("Error: " + err.Error())
	}
	filebytes, err := ioutil.ReadAll(file)
	if err != nil {
		data = []byte("Error: " + err.Error())
	}
	data = blackfriday.MarkdownCommon(filebytes)
	var msg BrowserMsg
	msg.Markdown = string(data)
	err = websocket.JSON.Send(l.Socket, msg)
	if err != nil {
		log.Println("Error sending message:", err)
	}
}

func UpdateListeners(updates chan Update, listeners chan Listener) {
	currentListeners := make([]Listener, 0)
	for {
		select {
		case listener := <-listeners:
			if listener.State == Open {
				log.Println("New listener on", listener.File)
				currentListeners = append(currentListeners, listener)
				writeFileForListener(listener)
			}
			if listener.State == Close {
				for i, l := range currentListeners {
					if l.Socket == listener.Socket {
						log.Println("Deregistering Listener")
						currentListeners = append(currentListeners[:i], currentListeners[i+1:]...)
					}
				}
			}
		case update := <-updates:
			log.Println("Update on", update.File)
			for _, l := range currentListeners {
				if update.File == l.File {
					writeFileForListener(l)
				}
			}
		}
	}
}

func RootFunc(w http.ResponseWriter, r *http.Request) {
	tocMutex.Lock()
	localToc := make([]string, len(toc))
	copy(localToc, toc)
	tocMutex.Unlock()
	log.Println(localToc)
	for i, s := range localToc {
		chop := strings.TrimPrefix(s, path)
		localToc[i] = "* [" + chop + "](/md" + chop + ")"
	}
	tocMkd := strings.Join(localToc, "\n")
	bytes := blackfriday.MarkdownCommon([]byte(tocMkd))
	rootTmpl.Execute(w, string(bytes))
}

func CSSFunc(css string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(css))
	}
}

func PageFunc(w http.ResponseWriter, r *http.Request) {
	subpath := strings.TrimPrefix(r.RequestURI, "/md")
	log.Println("New watcher on ", subpath)
	pageTmpl.Execute(w, subpath)
}

func HandleListener(listeners chan Listener) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		subpath := strings.TrimPrefix(ws.Request().RequestURI, "/ws")
		listeners <- Listener{subpath, ws, Open}
		var closeMessage string
		err := websocket.Message.Receive(ws, &closeMessage)
		if err != nil && err.Error() != "EOF" {
			log.Println("Error before close:", err)
		}
		listeners <- Listener{subpath, ws, Close}
	}
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
	updates := make(chan Update)
	go WatcherEventLoop(watcher, updates, done)

	log.Println("Watching directory", path)
	err = filepath.Walk(path, AddWatch(watcher))
	if err != nil {
		log.Fatal(err)
	}

	listeners := make(chan Listener)
	go UpdateListeners(updates, listeners)

	http.HandleFunc("/", RootFunc)
	http.HandleFunc("/md/", PageFunc)
	http.HandleFunc("/github.css", CSSFunc(githubCss))
	http.Handle("/ws/", websocket.Handler(HandleListener(listeners)))
	http.ListenAndServe(":8080", nil)
}

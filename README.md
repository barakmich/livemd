# livemd

## Overview

I wanted a simple tool that watched file updates as I worked on design docs in [a real editor](http://vim.org).
I also didn't want to `npm install` anything. So I wrote a server in Go, with live Markdown updates over websockets. 

## Libraries Used

* [github.com/russross/blackfriday](https://github.com/russross/blackfriday)
*	[golang.org/x/net/websocket](https://golang.org/x/net/websocket)
*	[gopkg.in/fsnotify.v1](https://github.com/go-fsnotify/fsnotify)

## Usage
```
go get github.com/barakmich/livemd
cd $PROJECT_DIR
livemd
```

And visit [http://localhost:8080](http://localhost:8080/) in your browser. When you save Markdown files in that directory, if you're looking at the file, it will automatically reupdate the content.

## License

BSD

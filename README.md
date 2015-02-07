# livemd

## Overview

I wanted a simple tool that watched file updates as I worked on design docs. 
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

## License

BSD

package main

var pageTemplate = `
<html>
<head>
<title>LiveMarkdown {{.}}</title>
<link rel="stylesheet" type="text/css" href="/github.css">
<script type="text/javascript" src="https://code.jquery.com/jquery-2.1.3.min.js"></script>
<script type="text/javascript">
var ws = new WebSocket("ws://" + location.host + "/ws" + "{{.}}")
ws.onopen = function() {
	$("#content").html("wooot")
}

ws.onmessage = function(evt) {
	var data = JSON.parse(evt.data)
	$("#content").html(data.Markdown)
}

ws.onclose = function() {
}

</script>
<style>
body {
	padding: 10px 20px;
}
</style>
</head>
<body>
<div id="content" class="markdown-body">
</div>
</body>
</html>
`

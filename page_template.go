package main

var pageTemplate = `
<html>
<head>
<title>LiveMarkdown</title>
<script type="text/javascript" src="https://code.jquery.com/jquery-2.1.3.min.js"></script>
<script type="text/javascript">
var ws = new WebSocket("ws://" + location.host + "/ws" + "{{.}}")
ws.onopen = function() {
	$("body").Text("wooot")
}

ws.onmessage = function(evt) {
	$("body").Text(evt)
}

ws.onclose = function() {
}

</script>
</head>
<body>
</body>
</html>
`

package main

var rootTemplate = `
<html>
<head>
<title>LiveMarkdown</title>
<link rel="stylesheet" type="text/css" href="/github.css">
<style>
body {
	padding: 10px 20px;
}
</style>
</head>
<body>
<h1> LiveMD: Listing of files </h1>
{{.}}
</body>
</html>
`

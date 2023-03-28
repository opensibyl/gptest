package gptest

const subPageTemplate = `
<html>
<body>
<pre class="prettyprint">
<code>
%s
</code>
</pre>
<script src="https://cdn.jsdelivr.net/gh/google/code-prettify@master/loader/run_prettify.js"></script>
</body>
</html>
`

const indexTemplate = `
	<!DOCTYPE html>
	<html>
	<head>
		<title>GPT EST Result</title>
	</head>
	<body>
		<h1>GPT EST Result</h1>
		<ul>
		{{range $filePath, $funcs := .}}
			{{range $func := $funcs}}
			<li><a href="{{ $filePath }}_{{ $func.Name }}.html">{{ $filePath }}_{{ $func.Name }}</a></li>
			{{end}}
		{{end}}
		</ul>
	</body>
	</html>
`

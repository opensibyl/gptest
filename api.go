package gptest

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/opensibyl/sibyl2"
	"github.com/opensibyl/sibyl2/pkg/server"
	"github.com/opensibyl/sibyl2/pkg/server/object"
	"github.com/opensibyl/squ/extractor"
	"github.com/opensibyl/squ/indexer"
	object2 "github.com/opensibyl/squ/object"
)

func Run(token string, path string, ctx context.Context) error {
	config := object2.DefaultConfig()
	config.IndexerType = object2.IndexerGolang
	config.RunnerType = object2.RunnerGolang

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}

	log.Println("using local sibyl, starting ...")
	go func() {
		config := object.DefaultExecuteConfig()
		// for performance
		// disable stdout
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = io.Discard
		//
		config.BindingConfigPart.DbType = object.DriverTypeInMemory
		config.EnableLog = false
		config.Port = 9875
		err := server.Execute(config, ctx)
		if err != nil {
			panic(err)
		}
	}()

	// 1. index
	curIndexer, err := indexer.GetIndexer(config.IndexerType, &config)
	err = curIndexer.UploadSrc(ctx)
	PanicIfErr(err)
	log.Println("indexer ready")

	// 2. extract
	// line level diff
	calcContext, cancel := context.WithCancel(ctx)
	defer cancel()
	curExtractor, err := extractor.NewDiffExtractor(&config)
	PanicIfErr(err)
	diffMap, err := curExtractor.ExtractDiffMethods(calcContext)
	PanicIfErr(err)
	log.Printf("diff calc ready: %v\n", len(diffMap))

	// 3. prepare data
	client := GetClient(ClientGpt35)
	client.SetToken(token)
	err = client.Prepare()
	PanicIfErr(err)

	// 4. collect
	cache := make(map[*sibyl2.FunctionWithPath]string)
	for eachFile, eachFuncs := range diffMap {
		for _, eachFunc := range eachFuncs {
			log.Printf("gen case for %v in %v\n", eachFunc.GetName(), eachFile)

			// collect itself
			funcDefs := curIndexer.GetVertexesWithSignature(eachFunc.GetSignature())
			if len(funcDefs) == 0 {
				continue
			}
			vertex, err := curIndexer.GetSibylCache().CallGraph.Graph.Vertex(funcDefs[0])
			PanicIfErr(err)

			askStr := fmt.Sprintf(`
Generate one case for this method:

%s

return me a code snippet only, without markdown wrapper, without any note.
\n
`, vertex.Unit.Content)

			// collect relative info
			referencedCalls := curIndexer.GetSibylCache().FindReverseCalls(vertex)
			if len(referencedCalls) != 0 {
				askStr += fmt.Sprintf(`
It will called by:
%s
`, referencedCalls[0].Unit.Content)
			}

			resp, err := client.Ask(askStr)
			PanicIfErr(err)

			// generate markdownTemplate
			cache[vertex] = resp
		}
	}

	// parse cache and generate html for each file
	fileCache := make(map[string][]*sibyl2.FunctionWithPath)
	for funcWithPath, htmlTemplate := range cache {
		filePath := funcWithPath.Path
		if _, ok := fileCache[filePath]; !ok {
			fileCache[filePath] = []*sibyl2.FunctionWithPath{funcWithPath}
		} else {
			fileCache[filePath] = append(fileCache[filePath], funcWithPath)
		}

		// write to file
		err = os.WriteFile(fmt.Sprintf("%s/%s_%s.html", path, funcWithPath.Path, funcWithPath.Name), []byte(fmt.Sprintf(`
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
`, htmlTemplate)), 0644)
		PanicIfErr(err)
	}

	// generate index.html
	indexTemplate := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>GPT Test Result</title>
	</head>
	<body>
		<h1>GPT Test Result</h1>
		<ul>
		{{range $filePath, $funcs := .}}
			{{range $func := $funcs}}
			<li><a href="{{ $filePath }}_{{ $func.Name }}.html">{{ $func.Name }}</a></li>
			{{end}}
		{{end}}
		</ul>
	</body>
	</html>
	`
	tmpl, err := template.New("index").Parse(indexTemplate)
	PanicIfErr(err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, fileCache)
	PanicIfErr(err)

	err = os.WriteFile(fmt.Sprintf("%s/index.html", path), buf.Bytes(), 0644)
	PanicIfErr(err)

	return nil
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

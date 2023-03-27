package gptest

import (
	"context"
	"fmt"
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

func Run(token string, ctx context.Context) error {
	config := object2.DefaultConfig()
	config.IndexerType = object2.IndexerGolang
	config.RunnerType = object2.RunnerGolang

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

return me a code snippet only, with markdown wrapper, without any note.
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

	// 5. render
	// embed it in html
	final := ""
	for funcInfo, content := range cache {
		final += fmt.Sprintf(`
## %s in %s

%s

---

`, funcInfo.Name, funcInfo.Path, fmt.Sprintf(".remark-code.hljs[```\n\t\t%s\n```]", content))
	}

	htmlTemplate := fmt.Sprintf(`
<!DOCTYPE html>
<html>
  <head>
    <title>GPTEST REPORT</title>
    <meta charset="utf-8">
    <style>
      @import url(https://fonts.googleapis.com/css?family=Yanone+Kaffeesatz);
      @import url(https://fonts.googleapis.com/css?family=Droid+Serif:400,700,400italic);
      @import url(https://fonts.googleapis.com/css?family=Ubuntu+Mono:400,700,400italic);

      body { font-family: 'Droid Serif'; }
      h1, h2, h3 {
        font-family: 'Yanone Kaffeesatz';
        font-weight: normal;
      }
      .remark-code, .remark-inline-code { font-family: 'Ubuntu Mono'; }
    </style>
  </head>
  <body>
    <textarea id="source">

%s

</textarea>
    <script src="https://remarkjs.com/downloads/remark-latest.min.js">
    </script>
    <script>
      var slideshow = remark.create();
    </script>
  </body>
</html>
`, final)

	// write to file
	err = os.WriteFile("gpt_test_result.html", []byte(htmlTemplate), 0644)
	PanicIfErr(err)

	return nil
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

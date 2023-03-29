package gptest

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/opensibyl/sibyl2"
	"github.com/opensibyl/sibyl2/pkg/server"
	"github.com/opensibyl/sibyl2/pkg/server/object"
)

func Run(config SharedConfig, ctx context.Context) error {
	repoInfo, err := GetRepoInfoFromDir(config.SrcDir)
	config.RepoInfo = repoInfo

	if _, err := os.Stat(config.OutputDir); os.IsNotExist(err) {
		err := os.MkdirAll(config.OutputDir, os.ModePerm)
		PanicIfErr(err)
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

	apiClient, err := config.NewSibylClient()
	PanicIfErr(err)

	// 1. index
	curIndexer := &BaseIndexer{
		config:        &config,
		apiClient:     apiClient,
		caseSet:       make(map[string]interface{}),
		vertexMapping: make(map[string]*map[string]interface{}),
	}
	err = curIndexer.UploadSrc(ctx)
	PanicIfErr(err)
	log.Println("indexer ready")

	// 2. extract
	// line level diff
	calcContext, cancel := context.WithCancel(ctx)
	defer cancel()
	curExtractor := &GitExtractor{
		config:    &config,
		apiClient: apiClient,
	}

	PanicIfErr(err)
	diffMap, err := curExtractor.ExtractDiffMethods(calcContext)
	PanicIfErr(err)
	log.Printf("diff calc ready: %v\n", len(diffMap))

	// 3. prepare data
	client := GetClient(ClientGpt35)
	client.SetToken(config.Token)
	err = client.Prepare(config.PromptFile)
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
Generate 1 case for this method:

%s

`, vertex.Unit.Content)

			// collect relative info
			referencedCalls := curIndexer.GetSibylCache().FindReverseCalls(vertex)
			if len(referencedCalls) != 0 {
				askStr += fmt.Sprintf(`
It will called by:
%s
`, referencedCalls[0].Unit.Content)
			}

			// params and returns
			for _, each := range vertex.Parameters {
				clazzWithPaths, _, err := apiClient.RegexQueryApi.ApiV1RegexClazzGet(ctx).
					Repo(config.RepoInfo.RepoId).
					Rev(config.RepoInfo.RevHash).
					Field("name").
					Regex(fmt.Sprintf("^%s$", each.Type)).Execute()
				if err != nil {
					// give up
					log.Printf("failed to get class: %v\n", each.Type)
					continue
				}

				// add related types to gpt
				for _, eachClazz := range clazzWithPaths {
					desc, err := eachClazz.MarshalJSON()
					if err != nil {
						// give up
						log.Printf("failed to marshal: %v\n", eachClazz.Name)
						continue
					}
					askStr += fmt.Sprintf(`
related params and returns types:
%s
`, string(desc))
				}
			}

			askStr += `
return me a code snippet only, without markdown wrapper, without any note.
`
			log.Println(askStr)
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
		p := filepath.Join(config.OutputDir, funcWithPath.Path)
		dirPath := filepath.Dir(p)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err := os.MkdirAll(dirPath, os.ModePerm)
			PanicIfErr(err)
		}

		err = os.WriteFile(fmt.Sprintf("%s/%s_%s.html", dirPath, filepath.Base(funcWithPath.Path), funcWithPath.Name), []byte(fmt.Sprintf(subPageTemplate, funcWithPath.Path, htmlTemplate)), 0644)
		PanicIfErr(err)
	}

	// generate index.html
	tmpl, err := template.New("index").Parse(indexTemplate)
	PanicIfErr(err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, fileCache)
	PanicIfErr(err)

	err = os.WriteFile(fmt.Sprintf("%s/index.html", config.OutputDir), buf.Bytes(), 0644)
	PanicIfErr(err)

	return nil
}

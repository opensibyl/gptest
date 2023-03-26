package gptest

import (
	"context"
	"io"
	"log"

	"github.com/gin-gonic/gin"
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
	for eachFile, eachFuncs := range diffMap {
		for _, eachFunc := range eachFuncs {
			log.Printf("gen case for %v in %v\n", eachFunc.GetName(), eachFile)

			// collect related objects
			funcDefs := curIndexer.GetVertexesWithSignature(eachFunc.GetSignature())
			if len(funcDefs) == 0 {
				continue
			}
			vertex, err := curIndexer.GetSibylCache().CallGraph.Graph.Vertex(funcDefs[0])
			PanicIfErr(err)

			resp, err := client.Ask(vertex.Unit.Content)
			PanicIfErr(err)
			log.Printf("resp: %v\n", resp)
		}
	}

	return nil
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

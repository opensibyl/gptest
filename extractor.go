package gptest

import (
	"context"
	"log"
	"os/exec"
	"strconv"
	"strings"

	openapi "github.com/opensibyl/sibyl-go-client"
	"github.com/opensibyl/sibyl2/pkg/core"
	"github.com/opensibyl/sibyl2/pkg/ext"
)

type GitExtractor struct {
	config    *SharedConfig
	apiClient *openapi.APIClient
}

func (g *GitExtractor) ExtractDiffMap(_ context.Context) (DiffMap, error) {
	gitDiffCmd := exec.Command("git", "diff", g.config.Before, g.config.After)
	gitDiffCmd.Dir = g.config.SrcDir
	patchRaw, err := gitDiffCmd.CombinedOutput()
	if err != nil {
		core.Log.Errorf("git cmd error: %s", patchRaw)
		panic(err)
	}

	return ext.Unified2Affected(patchRaw)
}

func (g *GitExtractor) ExtractDiffMethods(ctx context.Context) (map[string][]openapi.ObjectFunctionWithSignature, error) {
	diffMap, err := g.ExtractDiffMap(ctx)
	if err != nil {
		return nil, err
	}

	// method level diff, and influence
	influencedMethods := make(map[string][]openapi.ObjectFunctionWithSignature)
	for eachFile, eachLineList := range diffMap {
		eachLineStrList := make([]string, 0, len(eachLineList))
		for _, eachLine := range eachLineList {
			eachLineStrList = append(eachLineStrList, strconv.Itoa(eachLine))
		}
		functionWithSignatures, _, err := g.apiClient.BasicQueryApi.
			ApiV1FuncGet(ctx).
			Repo(g.config.RepoInfo.RepoId).
			Rev(g.config.RepoInfo.RevHash).
			File(eachFile).
			Lines(strings.Join(eachLineStrList, ",")).
			Execute()
		if err != nil {
			return nil, err
		}
		log.Printf("%s %v => functions %d", eachFile, eachLineList, len(functionWithSignatures))
		influencedMethods[eachFile] = append(influencedMethods[eachFile], functionWithSignatures...)
	}
	return influencedMethods, nil
}

type DiffMap = map[string][]int

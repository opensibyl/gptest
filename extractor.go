package gptest

import (
	"context"
	"log"
	"os/exec"
	"regexp"
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

func (g *GitExtractor) ExtractDiffMapWithRegex(_ context.Context) (DiffMap, error) {
	gitCmd := exec.Command("git", "ls-files")
	gitCmd.Dir = g.config.SrcDir
	fileListRaw, err := gitCmd.CombinedOutput()
	if err != nil {
		log.Printf("git cmd error: %s", fileListRaw)
		panic(err)
	}
	fileList := strings.Split(string(fileListRaw), "\n")
	regex, err := regexp.Compile(g.config.FileInclude)
	if err != nil {
		return nil, err
	}

	diffMap := make(DiffMap)
	for _, each := range fileList {
		if regex.MatchString(each) {
			diffMap[each] = nil
		}
	}
	return diffMap, nil
}

func (g *GitExtractor) ExtractDiffMethods(ctx context.Context) (map[string][]openapi.ObjectFunctionWithSignature, error) {
	var diffMap DiffMap
	var err error
	if g.config.FileInclude == "" {
		diffMap, err = g.ExtractDiffMap(ctx)
	} else {
		diffMap, err = g.ExtractDiffMapWithRegex(ctx)
	}

	if err != nil {
		return nil, err
	}

	// method level diff, and influence
	influencedMethods := make(map[string][]openapi.ObjectFunctionWithSignature)
	for eachFile, eachLineList := range diffMap {
		var functionWithSignatures []openapi.ObjectFunctionWithSignature
		if eachLineList == nil {
			functionWithSignatures, _, err = g.apiClient.BasicQueryApi.
				ApiV1FuncGet(ctx).
				Repo(g.config.RepoInfo.RepoId).
				Rev(g.config.RepoInfo.RevHash).
				File(eachFile).
				Execute()
		} else {
			eachLineStrList := make([]string, 0, len(eachLineList))
			for _, eachLine := range eachLineList {
				eachLineStrList = append(eachLineStrList, strconv.Itoa(eachLine))
			}
			functionWithSignatures, _, err = g.apiClient.BasicQueryApi.
				ApiV1FuncGet(ctx).
				Repo(g.config.RepoInfo.RepoId).
				Rev(g.config.RepoInfo.RevHash).
				File(eachFile).
				Lines(strings.Join(eachLineStrList, ",")).
				Execute()
		}
		if err != nil {
			return nil, err
		}
		influencedMethods[eachFile] = append(influencedMethods[eachFile], functionWithSignatures...)

	}
	return influencedMethods, nil
}

type DiffMap = map[string][]int

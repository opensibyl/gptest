package gptest

import (
	"net/url"

	"github.com/opensibyl/sibyl-go-client"
)

type SharedConfig struct {
	Token       string    `json:"token"`
	SrcDir      string    `json:"srcDir"`
	OutputDir   string    `json:"outputDir"`
	RepoInfo    *RepoInfo `json:"repoInfo"`
	SibylUrl    string    `json:"sibylUrl"`
	Before      string    `json:"before"`
	After       string    `json:"after"`
	FileInclude string    `json:"fileInclude"`
	PromptFile  string    `json:"promptFile"`
}

func DefaultConfig() SharedConfig {
	return SharedConfig{
		"",
		".",
		".",
		nil,
		"http://127.0.0.1:9875",
		"HEAD~1",
		"HEAD",
		"",
		"",
	}
}

func (conf *SharedConfig) NewSibylClient() (*openapi.APIClient, error) {
	parsed, err := conf.parseSibylUrl()
	if err != nil {
		return nil, err
	}

	configuration := openapi.NewConfiguration()
	configuration.Scheme = parsed.Scheme
	configuration.Host = parsed.Host
	return openapi.NewAPIClient(configuration), nil
}

func (conf *SharedConfig) LocalSibyl() bool {
	parsed, err := conf.parseSibylUrl()
	if err != nil {
		return false
	}
	hostName := parsed.Hostname()
	if hostName == "127.0.0.1" || hostName == "localhost" {
		return true
	}
	return false
}

func (conf *SharedConfig) GetSibylPort() string {
	parsed, err := conf.parseSibylUrl()
	if err != nil {
		return ""
	}
	return parsed.Port()
}

func (conf *SharedConfig) parseSibylUrl() (*url.URL, error) {
	return url.Parse(conf.SibylUrl)
}

type RepoInfo struct {
	RepoId  string `json:"repoId"`
	RevHash string `json:"revHash"`
}

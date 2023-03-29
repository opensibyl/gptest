# GPTEST

unittest generator based on gpt tech

## NOTICE

Currently, this is a toy project. Go just for fun.

## Showcase

- Download a single file from [release page](https://github.com/opensibyl/gptest/releases)
- Run `export OPENAI_TOKEN=abcdefg`
- Run `./gptest`
- Then a html report will open automatically:

![20230329214721275](https://user-images.githubusercontent.com/13421694/228559732-332e086c-da8f-4a76-8812-1a66b61356b8.gif)

Just review, copy, paste and edit.

## Usage

```text
$ ./gptest --help
Usage of ./gptest:
  -after string
    	after (default "HEAD")
  -before string
    	before (default "HEAD~1")
  -include string
    	file include regex
  -output string
    	output (default ".")
  -promptFile string
    	promptFile file
  -src string
    	src (default ".")
  -token string
    	openai token
```

### Diff Mode

By default, it will generate case from the diff between `HEAD~1` and `HEAD`.

You can change it with `--before` and `--after`.

### Regex Mode

Sometimes you just want to add some cases for your existed code.

Running `./gptest --include=".*SOME/PKG.*"` will generate cases for all the files which matching your regex.

## Tech behind

The program then performs the following steps:

1. Indexes the repository
2. Extracts line-level diffs using a GitExtractor object
3. Convert line diff to function diff, and essential context of the diff
4. Ask GPT for a code snippet for each DIFF function
5. The generated test cases are then stored in a cache and used to create HTML files for each function

## Some `Why`

### Why not directly adding them to my codebase?

Because AI is only an assistant, only you know what kind of use case you want.
From another perspective, because of the lack (or not enough) of context, it is difficult for AI to generate use cases
that can be directly executed.

Here is an example:

```golang
func TestGitExtractor_ExtractDiffMethods(t *testing.T) {
// prepare config and context
config := SharedConfig{
SrcDir:     "testdata",
Token:      "",
PromptFile: "",
OutputDir:  "test_output",
Git: GitConfig{
RepoUrl:  "",
RevHash:  "",
FileMask: "",
},
FileInclude: "",
}
ctx := context.Background()

// prepare GitExtractor
extractor := GitExtractor{
config:    &config,
apiClient: nil,
}

// execute method
_, err := extractor.ExtractDiffMethods(ctx)

// assertion
if err != nil {
t.Errorf("Error executing ExtractDiffMethods(): %v", err)
}
}
```

It's a good start. But you still have to spend some time to make it work.

Maybe I am wrong. Feel free to leave your comment in issue board.

### Why don't you copy your code and ask AI directly?

Function is easy to copy. But context of function did not.
Telling AI enough details about your function is really hard.

## Roadmap

- Prompt / Ask improvement
- Report looks not convenient enough for real world usage

## Contribution

Issues and PRs are always welcome. I am really looking forward to use it in production. :)

## License

[Apache 2.0](LICENSE)

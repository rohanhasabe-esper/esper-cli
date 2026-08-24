package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/esper-io/esper-cli/internal/cmd"
	"github.com/esper-io/esper-cli/internal/cmd/generated"
)

var methods = map[string]bool{"get": true, "post": true, "put": true, "patch": true, "delete": true}

type spec struct {
	Info struct {
		Generation string `json:"x-esper-generation"`
	} `json:"info"`
	Paths map[string]map[string]operation `json:"paths"`
}
type operation struct {
	Parameters  []parameter `json:"parameters"`
	Destructive *bool       `json:"x-esper-destructive"`
	Pagination  string      `json:"x-esper-pagination"`
	Verb        string      `json:"x-esper-verb"`
	Noun        string      `json:"x-esper-noun"`
	ScopeParent string      `json:"x-esper-scope-parent"`
}
type parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
}

func main() {
	specDir := flag.String("spec-dir", "spec/openapi", "directory containing canonical OpenAPI files")
	flag.Parse()
	issues := check(*specDir)
	if len(issues) == 0 {
		fmt.Println("contract check passed")
		return
	}
	sort.Strings(issues)
	for _, issue := range issues {
		fmt.Fprintln(os.Stderr, "error:", issue)
	}
	os.Exit(1)
}

func check(specDir string) []string {
	byOperation := map[string]generated.Operation{}
	for _, operation := range generated.Operations() {
		byOperation[key(operation.Generation, operation.Method, operation.Path)] = operation
	}
	entries, err := os.ReadDir(specDir)
	if err != nil {
		return []string{fmt.Sprintf("read spec directory: %v", err)}
	}
	root := cmd.NewRootCommand()
	issues := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(specDir, entry.Name()))
		if err != nil {
			return []string{fmt.Sprintf("read %s: %v", entry.Name(), err)}
		}
		var document spec
		if err := json.Unmarshal(data, &document); err != nil {
			return []string{fmt.Sprintf("decode %s: %v", entry.Name(), err)}
		}
		for apiPath, item := range document.Paths {
			for method, operation := range item {
				if !methods[method] {
					continue
				}
				location := fmt.Sprintf("%s %s %s", strings.ToUpper(method), apiPath, document.Info.Generation)
				if operation.Destructive == nil || operation.Pagination == "" || operation.Verb == "" || operation.Noun == "" {
					issues = append(issues, location+": missing required x-esper annotation")
					continue
				}
				generatedOperation, ok := byOperation[key(document.Info.Generation, strings.ToUpper(method), apiPath)]
				if !ok {
					issues = append(issues, location+": no generated operation")
					continue
				}
				if generatedOperation.Noun != operation.Noun || generatedOperation.Verb != operation.Verb || generatedOperation.Destructive != *operation.Destructive {
					issues = append(issues, location+": generated metadata differs from overlay")
				}
				command, _, err := root.Find(generatedOperation.Command)
				if err != nil || command == root {
					issues = append(issues, location+": generated command is unreachable")
					continue
				}
				for _, parameter := range operation.Parameters {
					if parameter.In == "path" && parameter.Name != operation.ScopeParent && operation.ScopeParent == "" {
						continue
					}
					if parameter.In == "path" && operation.ScopeParent != "" {
						continue
					}
					if parameter.In == "header" || parameter.In == "query" {
						if command.Flags().Lookup(kebab(parameter.Name)) == nil {
							issues = append(issues, location+": missing --"+kebab(parameter.Name))
						}
					}
				}
				if *operation.Destructive && root.PersistentFlags().Lookup("yes") == nil {
					issues = append(issues, location+": destructive command lacks --yes")
				}
			}
		}
	}
	return issues
}

func key(generation, method, apiPath string) string {
	return generation + "\x00" + method + "\x00" + apiPath
}
func kebab(value string) string {
	var result []rune
	for index, char := range value {
		if char == '_' {
			result = append(result, '-')
			continue
		}
		if index > 0 && char >= 'A' && char <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, char)
	}
	return strings.ToLower(string(result))
}

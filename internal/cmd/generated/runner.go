package generated

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"

	esperruntime "github.com/esper-io/esper-cli/internal/runtime"
	"github.com/spf13/cobra"
)

type Operation struct {
	Generation  string
	Command     []string
	Method      string
	Path        string
	Noun        string
	Verb        string
	Pagination  string
	Destructive bool
	ScopeParent string
	Parameters  []Parameter
	Body        *Body
}
type Parameter struct {
	Name, In, Type  string
	Required, Scope bool
	Enum            []string
}
type Body struct {
	MediaType  string
	Properties []Property
}
type Property struct {
	Name, Type, Format string
	Required, File     bool
	Enum               []string
}

var generatedOperations []Operation

func init() {
	if err := json.Unmarshal(generatedOperationsJSON, &generatedOperations); err != nil {
		panic(err)
	}
}

func Operations() []Operation {
	return append([]Operation(nil), generatedOperations...)
}

func AddCommands(root *cobra.Command, options *esperruntime.GlobalOptions) {
	groups := map[string][]Operation{}
	for _, operation := range generatedOperations {
		groups[strings.Join(operation.Command, "\x00")] = append(groups[strings.Join(operation.Command, "\x00")], operation)
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		addCommand(root, strings.Split(key, "\x00"), groups[key], options)
	}
}

func addCommand(root *cobra.Command, commandPath []string, operations []Operation, options *esperruntime.GlobalOptions) {
	parent := root
	for _, segment := range commandPath[:len(commandPath)-1] {
		child, _, err := parent.Find([]string{segment})
		if err != nil || child == parent {
			child = &cobra.Command{Use: segment}
			parent.AddCommand(child)
		}
		parent = child
	}
	verb := commandPath[len(commandPath)-1]
	command := &cobra.Command{Use: verb, Args: cobra.ArbitraryArgs, RunE: func(command *cobra.Command, args []string) error {
		return run(command, args, operations, options)
	}}
	addFlags(command, operations)
	parent.AddCommand(command)
}

func addFlags(command *cobra.Command, operations []Operation) {
	seen := map[string]bool{}
	hasUnscoped := false
	for _, operation := range operations {
		hasUnscoped = hasUnscoped || operation.ScopeParent == ""
	}
	for _, operation := range operations {
		for _, parameter := range operation.Parameters {
			if parameter.In == "path" && !parameter.Scope {
				continue
			}
			addStringFlag(command, parameterFlagName(parameter), parameter.Required && parameter.Scope && !hasUnscoped, parameter.Enum, seen)
		}
		if operation.Pagination == "limit-offset" || operation.Pagination == "apps-envelope" {
			addStringFlag(command, "limit", false, nil, seen)
			addStringFlag(command, "offset", false, nil, seen)
			if !seen["all"] {
				command.Flags().Bool("all", false, "fetch all result pages")
				seen["all"] = true
			}
		}
		if operation.Body == nil {
			continue
		}
		if operation.Body.MediaType == "application/json" {
			addStringFlag(command, "body", false, nil, seen)
		}
		for _, property := range operation.Body.Properties {
			addStringFlag(command, property.Name, false, property.Enum, seen)
		}
	}
}

func addStringFlag(command *cobra.Command, name string, required bool, values []string, seen map[string]bool) {
	name = kebab(name)
	if seen[name] {
		return
	}
	command.Flags().String(name, "", name)
	seen[name] = true
	if required {
		_ = command.MarkFlagRequired(name)
	}
	if len(values) > 0 {
		_ = command.RegisterFlagCompletionFunc(name, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return values, cobra.ShellCompDirectiveDefault
		})
	}
}

func run(command *cobra.Command, args []string, operations []Operation, options *esperruntime.GlobalOptions) error {
	operation, err := selectOperation(command, operations)
	if err != nil {
		return err
	}
	requestPath, err := replacePath(command, operation, args)
	if err != nil {
		return err
	}
	body, contentType, err := bodyFor(command, operation)
	if err != nil {
		return err
	}
	if operation.Destructive {
		ok, err := esperruntime.Confirm(command.InOrStdin(), command.ErrOrStderr(), requestPath, 1, options.Yes)
		if err != nil {
			return err
		}
		if !ok {
			return esperruntime.NewError(esperruntime.CategoryUsage, fmt.Errorf("operation cancelled"))
		}
	}
	store, err := esperruntime.NewStateStore()
	if err != nil {
		return esperruntime.NewError(esperruntime.CategoryAuth, err)
	}
	state, err := store.Load()
	if err != nil {
		return esperruntime.NewError(esperruntime.CategoryAuth, err)
	}
	credentials, err := esperruntime.ResolveCredentials(state.Config, options.Environment, options.APIKey)
	if err != nil {
		return err
	}
	query := make(map[string][]string)
	for _, parameter := range operation.Parameters {
		if parameter.In == "query" && command.Flags().Changed(kebab(parameter.Name)) {
			query[parameter.Name] = []string{flagString(command, parameter.Name)}
		}
	}
	client := esperruntime.NewHTTPClient(credentials)
	response, err := client.DoWithContentType(command.Context(), operation.Method, requestPath, query, body, contentType)
	if err != nil {
		return err
	}
	if options.JSON {
		return esperruntime.WriteJSON(command.OutOrStdout(), response)
	}
	return esperruntime.WriteHuman(command.OutOrStdout(), response)
}

func selectOperation(command *cobra.Command, operations []Operation) (Operation, error) {
	for _, operation := range operations {
		if operation.ScopeParent != "" && command.Flags().Changed(kebab(operation.ScopeParent)) {
			return operation, nil
		}
	}
	for _, operation := range operations {
		if operation.ScopeParent == "" {
			return operation, nil
		}
	}
	return operations[0], nil
}

func replacePath(command *cobra.Command, operation Operation, args []string) (string, error) {
	index := 0
	result := operation.Path
	for _, parameter := range operation.Parameters {
		if parameter.In != "path" {
			continue
		}
		name := "{" + parameter.Name + "}"
		value := ""
		if parameter.Scope {
			value = flagString(command, parameterFlagName(parameter))
		}
		if value == "" {
			if index >= len(args) {
				return "", esperruntime.NewError(esperruntime.CategoryUsage, fmt.Errorf("missing path parameter %s", parameter.Name))
			}
			value = args[index]
			index++
		}
		result = strings.ReplaceAll(result, name, value)
	}
	if index != len(args) {
		return "", esperruntime.NewError(esperruntime.CategoryUsage, fmt.Errorf("unexpected positional arguments"))
	}
	return result, nil
}

func bodyFor(command *cobra.Command, operation Operation) ([]byte, string, error) {
	if operation.Body == nil {
		return nil, "application/json", nil
	}
	bodyFlag := command.Flags().Changed("body")
	propertiesSet := false
	for _, property := range operation.Body.Properties {
		propertiesSet = propertiesSet || command.Flags().Changed(kebab(property.Name))
	}
	if bodyFlag && propertiesSet {
		return nil, "", esperruntime.NewError(esperruntime.CategoryUsage, fmt.Errorf("--body cannot be combined with property flags"))
	}
	if operation.Body.MediaType == "application/json" && bodyFlag {
		data, err := readJSONBody(flagString(command, "body"), command.InOrStdin())
		return data, "application/json", err
	}
	if operation.Body.MediaType == "multipart/form-data" {
		return multipartBody(command, operation.Body)
	}
	values := map[string]any{}
	for _, property := range operation.Body.Properties {
		if !command.Flags().Changed(kebab(property.Name)) {
			continue
		}
		values[property.Name] = typedValue(flagString(command, property.Name), property.Type)
	}
	if operation.Method != "PATCH" {
		for _, property := range operation.Body.Properties {
			if property.Required && !command.Flags().Changed(kebab(property.Name)) {
				return nil, "", esperruntime.NewError(esperruntime.CategoryUsage, fmt.Errorf("required flag --%s is not set", kebab(property.Name)))
			}
		}
	}
	if len(values) == 0 {
		return nil, "application/json", nil
	}
	data, err := json.Marshal(values)
	return data, "application/json", err
}

func multipartBody(command *cobra.Command, body *Body) ([]byte, string, error) {
	var output strings.Builder
	writer := multipart.NewWriter(&output)
	for _, property := range body.Properties {
		flag := kebab(property.Name)
		if !command.Flags().Changed(flag) {
			if property.Required {
				return nil, "", esperruntime.NewError(esperruntime.CategoryUsage, fmt.Errorf("required flag --%s is not set", flag))
			}
			continue
		}
		value := flagString(command, property.Name)
		if property.File {
			file, err := os.Open(value)
			if err != nil {
				return nil, "", esperruntime.NewError(esperruntime.CategoryUsage, fmt.Errorf("open --%s: %w", flag, err))
			}
			part, err := writer.CreateFormFile(property.Name, filepath.Base(value))
			if err == nil {
				_, err = io.Copy(part, file)
			}
			file.Close()
			if err != nil {
				return nil, "", fmt.Errorf("write --%s: %w", flag, err)
			}
			continue
		}
		if err := writer.WriteField(property.Name, value); err != nil {
			return nil, "", err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return []byte(output.String()), writer.FormDataContentType(), nil
}

func readJSONBody(value string, input io.Reader) ([]byte, error) {
	var data []byte
	var err error
	if value == "-" {
		data, err = io.ReadAll(input)
	} else if strings.HasPrefix(value, "@") {
		data, err = os.ReadFile(strings.TrimPrefix(value, "@"))
	} else {
		data = []byte(value)
	}
	if err != nil {
		return nil, esperruntime.NewError(esperruntime.CategoryUsage, fmt.Errorf("read --body: %w", err))
	}
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, esperruntime.NewError(esperruntime.CategoryUsage, fmt.Errorf("validate --body JSON: %w", err))
	}
	return data, nil
}
func typedValue(value, kind string) any {
	if kind == "boolean" {
		return value == "true"
	}
	if kind == "integer" {
		var number int64
		_, _ = fmt.Sscan(value, &number)
		return number
	}
	if kind == "number" {
		var number float64
		_, _ = fmt.Sscan(value, &number)
		return number
	}
	return value
}
func flagString(command *cobra.Command, name string) string {
	value, _ := command.Flags().GetString(kebab(name))
	return value
}

func parameterFlagName(parameter Parameter) string {
	if parameter.Scope {
		name := strings.TrimSuffix(parameter.Name, "_id")
		name = strings.TrimSuffix(name, "Id")
		return kebab(name)
	}
	return kebab(parameter.Name)
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

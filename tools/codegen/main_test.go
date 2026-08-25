package main

import "testing"

func TestExtractBodyRetainsRequirednessAndCollisionAutoFill(t *testing.T) {
	body := extractBody(map[string]requestMedia{
		"application/json": {Schema: schema{Type: "object", Properties: map[string]schema{
			"enterprise": {Type: "string"},
			"name":       {Type: "string"},
			"settings":   {Type: "object"},
		}, Required: []string{"enterprise", "settings"}}},
	}, true, []generatedParameter{{Name: "enterprise_id", In: "path", Scope: true, ScopeName: "enterprise"}}, nil)
	if !body.Required || !body.BodyOnly || len(body.Properties) != 0 {
		t.Fatalf("body metadata = %#v", body)
	}
	if len(body.AutoFill) != 1 || body.AutoFill[0].Name != "enterprise" || body.AutoFill[0].Parameter != "enterprise_id" {
		t.Fatalf("collision auto-fill = %#v", body.AutoFill)
	}
}

func TestExtractBodyRequiredEmptyObject(t *testing.T) {
	body := extractBody(map[string]requestMedia{"application/json": {Schema: schema{Type: "object"}}}, true, nil, nil)
	if !body.Required || !body.Empty {
		t.Fatalf("body metadata = %#v", body)
	}
}

func TestExtractBodyRequiredRootArrayIsNotEmptyObject(t *testing.T) {
	body := extractBody(map[string]requestMedia{"application/json": {Schema: schema{Type: "array"}}}, true, nil, nil)
	if !body.Required || body.Empty {
		t.Fatalf("body metadata = %#v", body)
	}
}

func TestScopeParameterNamesIncludesAncestorsOnly(t *testing.T) {
	scopes := scopeParameterNames("/stages/{stage_id}/runs/{run_id}/commands/{command_id}", "run")
	if scopes["stage_id"] != "stage" || scopes["run_id"] != "pipeline-run" || scopes["command_id"] != "" {
		t.Fatalf("scopes = %#v", scopes)
	}
}

func TestAliasesUseCanonicalCommand(t *testing.T) {
	operations := []generatedOperation{
		{Generation: "v0", Noun: "geofence", Verb: "get", OperationID: "canonical"},
		{Generation: "v0", Noun: "the-geofence", Verb: "get", OperationID: "alias", AliasOf: "canonical"},
	}
	assignCommands(operations)
	if operations[1].Noun != "geofence" || len(operations[1].Command) != 2 || operations[1].Command[0] != "geofence" {
		t.Fatalf("alias command = %#v", operations[1])
	}
}

func TestPipelineCommandAndOperationAlwaysUseSideFamilyPrefix(t *testing.T) {
	operations := []generatedOperation{
		{Generation: "pipelines-v0", Noun: "command", Verb: "list", ScopeParent: "target-run"},
		{Generation: "pipelines-v0", Noun: "operation", Verb: "get", ScopeParent: "stage-run"},
	}
	assignCommands(operations)
	if operations[0].Noun != "pipeline-command" || operations[1].Noun != "pipeline-operation" {
		t.Fatalf("pipeline nouns = %#v", operations)
	}
}

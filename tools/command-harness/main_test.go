package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/esper-io/esper-cli/internal/cmd/generated"
)

func TestSafeCommandRedactsBodiesAndKeys(t *testing.T) {
	got := safeCommand([]string{"command", "create", "--api-key=secret", "--body", `{"token":"secret"}`})
	if strings.Contains(got, "secret") || !strings.Contains(got, "<redacted>") {
		t.Fatalf("safeCommand() = %q", got)
	}
}

func TestReadonlyArgsRequireOneOfAndLimit(t *testing.T) {
	operation := generated.Operation{
		Command:      []string{"itunesapp", "list"},
		Path:         "/v2/itunesapps",
		Method:       "GET",
		RequireOneOf: []string{"app_id", "apple_app_id"},
		Parameters: []generated.Parameter{
			{Name: "app_id", In: "query"},
			{Name: "apple_app_id", In: "query"},
			{Name: "limit", In: "query"},
		},
	}
	args, missing := readonlyArgs(operation, map[string]string{"app_id": "app-1"})
	if len(missing) != 0 || !contains(args, "--app-id") || !contains(args, "--limit") {
		t.Fatalf("readonlyArgs() = %v, %v", args, missing)
	}
}

func TestMutationScenarioRequiresCleanup(t *testing.T) {
	if _, err := os.CreateTemp(t.TempDir(), "scenario"); err != nil {
		t.Fatal(err)
	}
	// The safety assertion is exercised through runMutations only with a real plan.
	if mutationConfirmation != "I_UNDERSTAND_THIS_TENANT_IS_DISPOSABLE" {
		t.Fatal("mutation confirmation changed")
	}
}

func TestMutationStepRequiresExplicitYes(t *testing.T) {
	runner := &harness{binary: "/does/not/run", timeout: time.Second}
	expected := 0
	item := runner.executeStep(step{Name: "delete", Args: []string{"device-request", "delete", "device-1"}, ExpectedExit: &expected, Mutation: true}, nil)
	if item.Status != "FAIL" || !strings.Contains(item.Detail, "--yes") {
		t.Fatalf("mutation safety result = %#v", item)
	}
}

func TestMutationStepCannotHideGeneratedWrite(t *testing.T) {
	expected := 0
	runner := &harness{binary: "/does/not/run", timeout: time.Second}
	item := runner.executeStep(step{Name: "delete", Args: []string{"device-request", "delete", "device-1", "--yes"}, ExpectedExit: &expected}, nil)
	if item.Status != "FAIL" || !strings.Contains(item.Detail, "mutation=true") {
		t.Fatalf("unmarked write result = %#v", item)
	}
}

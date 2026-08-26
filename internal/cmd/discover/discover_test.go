package discover

import (
	"strings"
	"testing"

	"github.com/esper-io/esper-cli/internal/cmd/generated"
)

func TestSearchMatchesPublicDocumentationURL(t *testing.T) {
	operations := []generated.Operation{
		{Command: []string{"geofence", "list"}, Method: "GET", Path: "/geofence/v1/geofences", Summary: "List all geofences", DocsSlugs: []string{"geofence_geofences"}},
		{Command: []string{"api", "v0", "geofence", "list"}, Method: "GET", Path: "/v0/enterprise/{enterprise_id}/geofence/", Summary: "List legacy geofences"},
	}
	matches := search(operations, "https://api.esper.io/openapi/geofence_geofences")
	if len(matches) != 1 || matches[0].Command != "geofence list" || matches[0].Docs[0] != "https://api.esper.io/openapi/geofence_geofences" {
		t.Fatalf("matches = %#v", matches)
	}
}

func TestSearchMatchesNestedBodyEnum(t *testing.T) {
	operations := []generated.Operation{{
		Command:     []string{"report-status", "create"},
		Method:      "POST",
		Path:        "/report/v0/reports",
		Summary:     "Create Report",
		OperationID: "createReportStatus",
		Body:        &generated.Body{Fields: []generated.BodyField{{Path: "filters.platform", Description: "Device platform", Enum: []string{"Android", "Apple"}}}},
	}}
	matches := search(operations, "Apple device report")
	if len(matches) != 1 || matches[0].Command != "report-status create" {
		t.Fatalf("matches = %#v", matches)
	}
}

func TestDocsSlugRejectsNonOpenAPIURLs(t *testing.T) {
	for _, value := range []string{"geofence_geofences", "https://api.esper.io/other/geofence_geofences"} {
		if got := docsSlug(value); got != "" {
			t.Fatalf("docsSlug(%q) = %q", value, got)
		}
	}
	if got := strings.Join(searchWords("Apple-device_report"), ","); got != "apple,device,report" {
		t.Fatalf("searchWords() = %q", got)
	}
}

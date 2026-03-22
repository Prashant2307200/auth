package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendErrorResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	SendErrorResponse(rr, http.StatusBadRequest, BAD_REQUEST, "invalid input")

	if got := rr.Code; got != http.StatusBadRequest {
		t.Fatalf("expected status %d got %d", http.StatusBadRequest, got)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json got %s", ct)
	}
	var er ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &er); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	if er.Code != BAD_REQUEST || er.Message != "invalid input" {
		t.Fatalf("unexpected body: %+v", er)
	}
}

func TestSendErrorResponseWithDetails(t *testing.T) {
	rr := httptest.NewRecorder()
	details := map[string]string{"field": "email"}
	SendErrorResponseWithDetails(rr, http.StatusConflict, CONFLICT, "already exists", details)

	if got := rr.Code; got != http.StatusConflict {
		t.Fatalf("expected status %d got %d", http.StatusConflict, got)
	}
	var er ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &er); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	if er.Details == "" {
		t.Fatalf("expected details to be present")
	}
	if !strings.Contains(er.Details, "field") {
		t.Fatalf("unexpected details value: %s", er.Details)
	}
}

func TestAllErrorCodesDefined(t *testing.T) {
	// sanity check that constants exist
	var _ ErrorCode = BAD_REQUEST
	var _ ErrorCode = UNAUTHORIZED
	var _ ErrorCode = FORBIDDEN
	var _ ErrorCode = NOT_FOUND
	var _ ErrorCode = CONFLICT
	var _ ErrorCode = INTERNAL_ERROR
	var _ ErrorCode = RATE_LIMITED
}

// Test that the details field is omitted from the JSON when empty
func TestDetailsOmittedWhenEmpty(t *testing.T) {
	rr := httptest.NewRecorder()
	SendErrorResponse(rr, http.StatusBadRequest, BAD_REQUEST, "no details")

	body := rr.Body.String()
	if strings.Contains(body, "\"details\"") {
		t.Fatalf("expected details to be omitted but found in body: %s", body)
	}
}

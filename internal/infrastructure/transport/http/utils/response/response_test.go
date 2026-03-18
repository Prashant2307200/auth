package response

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/utils"
	"github.com/Prashant2307200/auth-service/pkg/db"
)

func TestErrorToStatusMappings(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{err: nil, code: http.StatusOK},
		{err: utils.ErrInvalidInput, code: http.StatusBadRequest},
		{err: utils.ErrUnauthorized, code: http.StatusUnauthorized},
		{err: utils.ErrForbidden, code: http.StatusForbidden},
		{err: utils.ErrNotFound, code: http.StatusNotFound},
		{err: db.ErrNotFound, code: http.StatusNotFound},
		{err: errors.New("other"), code: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		if got := ErrorToStatus(tt.err); got != tt.code {
			t.Fatalf("ErrorToStatus(%v) = %d; want %d", tt.err, got, tt.code)
		}
	}
}

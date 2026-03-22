package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httputils "github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/stretchr/testify/require"
)

func TestBusinessHandler_Create_UnauthorizedEnvelope(t *testing.T) {
	mockBusinessRepo := &testutil.MockBusinessRepo{}
	mockUserRepo := &testutil.MockUserRepo{}
	uc := usecase.NewBusinessUseCase(mockBusinessRepo, mockUserRepo)
	h := &BusinessHandler{UC: uc}

	req := httptest.NewRequest(http.MethodPost, "/business/", nil)
	rr := httptest.NewRecorder()

	h.create(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)

	var er httputils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, httputils.UNAUTHORIZED, er.Code)
	require.Equal(t, "authentication required", er.Message)
}

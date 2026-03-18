package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/seeder"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
)

type DevHandler struct {
	UserRepo     interfaces.UserRepo
	BusinessRepo interfaces.BusinessRepo
}

func NewDevHandler(userRepo interfaces.UserRepo, businessRepo interfaces.BusinessRepo) *DevHandler {
	return &DevHandler{
		UserRepo:     userRepo,
		BusinessRepo: businessRepo,
	}
}

func (h *DevHandler) SeedDB(w http.ResponseWriter, r *http.Request) {
	if err := seeder.SeedAll(r.Context(), h.UserRepo, h.BusinessRepo); err != nil {
		response.WriteJson(w, response.ErrorToStatus(err), response.GeneralError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "seeded"})
}

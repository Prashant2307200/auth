package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/request"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/usecase"
)

type UserHandler struct {
	UC *usecase.UserUseCase
}

func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{UC: uc}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.getAll)
	mux.HandleFunc("POST /", h.create)
	mux.HandleFunc("GET /search", h.searchAll)
	mux.HandleFunc("GET /{id}/", h.getById)
	mux.HandleFunc("DELETE /{id}/", h.deleteById)
	mux.HandleFunc("PUT /{id}/", h.updateById)
}

func (h *UserHandler) getAll(w http.ResponseWriter, r *http.Request) {

	users, err := h.UC.GetUsers(r.Context())
	if err != nil {
		slog.Error("failed to get users", slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, users)
}

func (h *UserHandler) searchAll(w http.ResponseWriter, r *http.Request) {

	search := r.URL.Query().Get("q")

	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		if err == sql.ErrNoRows {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(errors.New("user not found with this search")))
		}
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	users, err := h.UC.SearchUsers(r.Context(), id, search)
	if err != nil {
		slog.Error("failed to get users", slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, users)
}

func (h *UserHandler) deleteById(w http.ResponseWriter, r *http.Request) {

	id, err := request.ParseId(r)
	if err != nil {
		slog.Error("failed to parse id", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	err = h.UC.DeleteUserById(r.Context(), id)
	if err != nil {
		slog.Error("failed to delete user", slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "user deleted successfully",
	})
}

func (h *UserHandler) updateById(w http.ResponseWriter, r *http.Request) {

	id, err := request.ParseId(r)
	if err != nil {
		slog.Error("failed to parse id", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	user, err := request.ParseJSON[entity.User](r)
	if err != nil {
		slog.Error("failed to parse json", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	err = h.UC.UpdateUserById(r.Context(), id, user)
	if err != nil {
		slog.Error("failed to update user", slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "user updated successfully",
	})
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {

	user, err := request.ParseJSON[entity.User](r)
	if err != nil {
		slog.Error("failed to parse json", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	err = h.UC.CreateUser(r.Context(), user)
	if err != nil {
		slog.Error("failed to create user", slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "user created successfully",
	})
}

func (h *UserHandler) getById(w http.ResponseWriter, r *http.Request) {

	id, err := request.ParseId(r)
	if err != nil {
		slog.Error("failed to parse id", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	user, err := h.UC.GetUserById(r.Context(), id)
	if err != nil {
		slog.Error("failed to get user", slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, user)
}
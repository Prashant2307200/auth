package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/request"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/usecase"
)

type BusinessHandler struct {
	UC *usecase.BusinessUseCase
}

type addBusinessUserRequest struct {
	UserID int64 `json:"user_id" validate:"required,gt=0"`
	Role   int   `json:"role" validate:"gte=0,lte=2"`
}

type createInviteRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  int    `json:"role" validate:"gte=0,lte=2"`
}

type addDomainRequest struct {
	Domain string `json:"domain" validate:"required,min=2"`
}

type verifyDomainRequest struct {
	VerificationToken string `json:"verification_token" validate:"required"`
}

type toggleDomainAutoJoinRequest struct {
	AutoJoinEnabled bool `json:"auto_join_enabled"`
}

func NewBusinessHandler(uc *usecase.BusinessUseCase) *BusinessHandler {
	return &BusinessHandler{UC: uc}
}

func (h *BusinessHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /", h.create)
	mux.HandleFunc("GET /my", h.getMyBusinesses)
	mux.HandleFunc("GET /{id}/", h.getByID)
	mux.HandleFunc("PUT /{id}/", h.update)
	mux.HandleFunc("DELETE /{id}/", h.delete)
	mux.HandleFunc("POST /{id}/users/", h.addUser)
	mux.HandleFunc("DELETE /{id}/users/{userId}/", h.removeUser)
	mux.HandleFunc("GET /{id}/users/", h.getUsers)
	mux.HandleFunc("POST /{id}/invites/", h.createInvite)
	mux.HandleFunc("GET /{id}/invites/", h.listInvites)
	mux.HandleFunc("DELETE /{id}/invites/{inviteId}/", h.revokeInvite)
	mux.HandleFunc("POST /{id}/domains/", h.addDomain)
	mux.HandleFunc("POST /{id}/domains/verify/", h.verifyDomain)
	mux.HandleFunc("PUT /{id}/domains/{domainId}/", h.toggleDomainAutoJoin)
}

func (h *BusinessHandler) create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	business, err := request.ParseJSON[entity.Business](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := response.ValidationError(business); err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	createdBusiness, err := h.UC.CreateBusiness(r.Context(), userID, business)
	if err != nil {
		response.WriteDomainError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusCreated, "business created successfully", createdBusiness)
}

func (h *BusinessHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	business, err := h.UC.GetBusinessByID(r.Context(), id)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, err)
		return
	}

	response.WriteJson(w, http.StatusOK, business)
}

func (h *BusinessHandler) update(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	id, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	business, err := request.ParseJSON[entity.Business](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := response.ValidationError(business); err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.UC.UpdateBusiness(r.Context(), userID, id, business); err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, "business updated successfully", nil)
}

func (h *BusinessHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	id, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.UC.DeleteBusiness(r.Context(), userID, id); err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "business deleted successfully", nil)
}

func (h *BusinessHandler) addUser(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	businessID, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	payload, err := request.ParseJSON[addBusinessUserRequest](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := response.ValidationError(payload); err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.UC.AddUserToBusiness(r.Context(), requesterID, businessID, payload.UserID, payload.Role); err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, "user added to business successfully", nil)
}

func (h *BusinessHandler) removeUser(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	businessID, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	rawUserID := r.PathValue("userId")
	userID, err := strconv.ParseInt(rawUserID, 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errors.New("userId must be a valid integer"))
		return
	}

	if err := h.UC.RemoveUserFromBusiness(r.Context(), requesterID, businessID, userID); err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "user removed from business successfully", nil)
}

func (h *BusinessHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	businessID, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	users, err := h.UC.GetBusinessUsers(r.Context(), requesterID, businessID)
	if err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}
	response.WriteJson(w, http.StatusOK, users)
}

func (h *BusinessHandler) getMyBusinesses(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	businesses, err := h.UC.GetUserBusinesses(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get my businesses", slog.Any("error", err))
		response.WriteDomainError(w, err)
		return
	}
	response.WriteJson(w, http.StatusOK, businesses)
}

func (h *BusinessHandler) createInvite(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	businessID, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	payload, err := request.ParseJSON[createInviteRequest](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := response.ValidationError(payload); err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	invite, err := h.UC.CreateInvite(r.Context(), requesterID, businessID, payload.Email, payload.Role)
	if err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, "invite created successfully", invite)
}

func (h *BusinessHandler) listInvites(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	businessID, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	invites, err := h.UC.ListInvites(r.Context(), requesterID, businessID)
	if err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}
	response.WriteJson(w, http.StatusOK, invites)
}

func (h *BusinessHandler) revokeInvite(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	businessID, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	rawInviteID := r.PathValue("inviteId")
	inviteID, err := strconv.ParseInt(rawInviteID, 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errors.New("inviteId must be a valid integer"))
		return
	}
	if err := h.UC.RevokeInvite(r.Context(), requesterID, businessID, inviteID); err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "invite revoked successfully", nil)
}

func (h *BusinessHandler) addDomain(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	businessID, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	payload, err := request.ParseJSON[addDomainRequest](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := response.ValidationError(payload); err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	domain, err := h.UC.AddDomain(r.Context(), requesterID, businessID, payload.Domain)
	if err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, "domain added successfully", domain)
}

func (h *BusinessHandler) verifyDomain(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	businessID, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	payload, err := request.ParseJSON[verifyDomainRequest](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := response.ValidationError(payload); err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	domain, err := h.UC.VerifyDomain(r.Context(), requesterID, businessID, payload.VerificationToken)
	if err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "domain verified successfully", domain)
}

func (h *BusinessHandler) toggleDomainAutoJoin(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	businessID, err := request.ParseId(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	rawDomainID := r.PathValue("domainId")
	domainID, err := strconv.ParseInt(rawDomainID, 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errors.New("domainId must be a valid integer"))
		return
	}
	payload, err := request.ParseJSON[toggleDomainAutoJoinRequest](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.UC.ToggleDomainAutoJoin(r.Context(), requesterID, businessID, domainID, payload.AutoJoinEnabled); err != nil {
		response.WriteError(w, http.StatusForbidden, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, "domain auto-join updated successfully", nil)
}

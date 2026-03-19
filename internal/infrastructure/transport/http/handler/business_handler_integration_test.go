package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/Prashant2307200/auth-service/pkg/db"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBusinessHandler_CreateBusiness_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetBySlug", mock.Anything, "acme-co").Return(nil, nil)
	mockBusiness.On("CreateWithOwner", mock.Anything, mock.AnythingOfType("*entity.Business"), int64(11)).Return(int64(100), nil)
	mockBusiness.On("GetById", mock.Anything, int64(100)).Return(&entity.Business{ID: 100, Name: "Acme Co", Slug: "acme-co", Email: "owner@acme.co"}, nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	body := `{"name":"Acme Co","slug":"acme-co","email":"owner@acme.co"}`
	req := httptest.NewRequest(http.MethodPost, "/businesses", bytes.NewBufferString(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), 11))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.create(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_CreateBusiness_Unauthorized(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses", bytes.NewBufferString(`{"name":"Acme"}`))
	rr := httptest.NewRecorder()

	h.create(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBusinessHandler_CreateBusiness_MalformedJSON(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses", bytes.NewBufferString(`{"name":`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 11))
	rr := httptest.NewRecorder()

	h.create(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBusinessHandler_GetBusiness_NotFound(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetById", mock.Anything, int64(404)).Return(nil, db.ErrNotFound)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/businesses?id=404", nil)
	rr := httptest.NewRecorder()

	h.getByID(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_UpdateBusiness_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(50), int64(12)).Return(usecase.BusinessRoleAdmin, nil)
	mockBusiness.On("Update", mock.Anything, int64(50), mock.AnythingOfType("*entity.Business")).Return(nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	body := `{"name":"Acme Updated","slug":"acme-updated","email":"updated@acme.co"}`
	req := httptest.NewRequest(http.MethodPut, "/businesses?id=50", bytes.NewBufferString(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), 12))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.update(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_DeleteBusiness_Unauthorized(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/businesses?id=50", nil)
	rr := httptest.NewRecorder()

	h.delete(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBusinessHandler_AddUserToBusiness_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(51), int64(13)).Return(usecase.BusinessRoleAdmin, nil)
	mockUser.On("GetById", mock.Anything, int64(99)).Return(testutil.CreateTestUserWithID(99), nil)
	mockBusiness.On("AddUser", mock.Anything, int64(51), int64(99), 1).Return(nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	body := `{"user_id":99,"role":1}`
	req := httptest.NewRequest(http.MethodPost, "/businesses?id=51", bytes.NewBufferString(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), 13))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.addUser(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockBusiness.AssertExpectations(t)
	mockUser.AssertExpectations(t)
}

func TestBusinessHandler_AddUserToBusiness_Unauthorized(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses?id=51", bytes.NewBufferString(`{"user_id":99,"role":1}`))
	rr := httptest.NewRecorder()

	h.addUser(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBusinessHandler_RemoveUserFromBusiness_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(52), int64(14)).Return(usecase.BusinessRoleOwner, nil)
	mockBusiness.On("GetUserRole", mock.Anything, int64(52), int64(98)).Return(usecase.BusinessRoleMember, nil)
	mockBusiness.On("RemoveUser", mock.Anything, int64(52), int64(98)).Return(nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/businesses?id=52", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 14))
	req.SetPathValue("userId", "98")
	rr := httptest.NewRecorder()

	h.removeUser(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_GetBusinessUsers_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	users := []*entity.User{testutil.CreateTestUserWithID(20), testutil.CreateTestUserWithID(21)}
	mockBusiness.On("GetUserRole", mock.Anything, int64(53), int64(15)).Return(usecase.BusinessRoleMember, nil)
	mockBusiness.On("GetUsers", mock.Anything, int64(53)).Return(users, nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/businesses?id=53", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 15))
	rr := httptest.NewRecorder()

	h.getUsers(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_CreateInvite_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(54), int64(16)).Return(usecase.BusinessRoleAdmin, nil)
	mockBusiness.On("CreateInvite", mock.Anything, mock.AnythingOfType("*entity.BusinessInvite")).Return(int64(700), nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	body := `{"email":"invitee@acme.co","role":1}`
	req := httptest.NewRequest(http.MethodPost, "/businesses?id=54", bytes.NewBufferString(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), 16))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.createInvite(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_ListInvites_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	invites := []*entity.BusinessInvite{testutil.CreateTestInvite(55, "member@acme.co", "token-1", entity.InviteStatusPending, time.Now().Add(24*time.Hour))}
	mockBusiness.On("GetUserRole", mock.Anything, int64(55), int64(17)).Return(usecase.BusinessRoleAdmin, nil)
	mockBusiness.On("ListInvites", mock.Anything, int64(55)).Return(invites, nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/businesses?id=55", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 17))
	rr := httptest.NewRecorder()

	h.listInvites(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_DeleteInvite_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(56), int64(18)).Return(usecase.BusinessRoleAdmin, nil)
	mockBusiness.On("RevokeInvite", mock.Anything, int64(801), int64(56)).Return(nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/businesses?id=56", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 18))
	req.SetPathValue("inviteId", "801")
	rr := httptest.NewRecorder()

	h.revokeInvite(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_CreateDomain_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(57), int64(19)).Return(usecase.BusinessRoleAdmin, nil)
	mockBusiness.On("GetDomain", mock.Anything, int64(57), "acme.co").Return(nil, errors.New("not found"))
	mockBusiness.On("CreateDomain", mock.Anything, mock.AnythingOfType("*entity.BusinessDomain")).Return(int64(901), nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses?id=57", bytes.NewBufferString(`{"domain":"acme.co"}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 19))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.addDomain(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_VerifyDomain_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	domain := &entity.BusinessDomain{ID: 902, BusinessID: 58, Domain: "acme.co", VerificationToken: "verify-token"}
	mockBusiness.On("GetUserRole", mock.Anything, int64(58), int64(20)).Return(usecase.BusinessRoleAdmin, nil)
	mockBusiness.On("GetDomainByVerificationToken", mock.Anything, "verify-token").Return(domain, nil)
	mockBusiness.On("VerifyDomain", mock.Anything, int64(902)).Return(nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses?id=58", bytes.NewBufferString(`{"verification_token":"verify-token"}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 20))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.verifyDomain(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_UpdateDomainAutoJoin_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(59), int64(21)).Return(usecase.BusinessRoleAdmin, nil)
	mockBusiness.On("UpdateDomainAutoJoin", mock.Anything, int64(903), int64(59), true).Return(nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPut, "/businesses?id=59", bytes.NewBufferString(`{"auto_join_enabled":true}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 21))
	req.SetPathValue("domainId", "903")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.toggleDomainAutoJoin(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_GetMyBusinesses_Unauthorized(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/businesses/my", nil)
	rr := httptest.NewRecorder()

	h.getMyBusinesses(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBusinessHandler_GetMyBusinesses_Error(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserBusinesses", mock.Anything, int64(22)).Return(nil, errors.New("db down"))

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/businesses/my", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 22))
	rr := httptest.NewRecorder()

	h.getMyBusinesses(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_DeleteBusiness_Forbidden(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(60), int64(23)).Return(usecase.BusinessRoleAdmin, nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/businesses?id=60", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 23))
	rr := httptest.NewRecorder()

	h.delete(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_AddUserToBusiness_Forbidden(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(61), int64(24)).Return(usecase.BusinessRoleMember, nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses?id=61", bytes.NewBufferString(`{"user_id":99,"role":1}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 24))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.addUser(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_RemoveUserFromBusiness_InvalidUserID(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/businesses?id=62", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 25))
	req.SetPathValue("userId", "not-int")
	rr := httptest.NewRecorder()

	h.removeUser(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBusinessHandler_CreateInvite_ValidationError(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses?id=63", bytes.NewBufferString(`{"email":"bad-email","role":1}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 26))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.createInvite(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBusinessHandler_ListInvites_Unauthorized(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/businesses?id=64", nil)
	rr := httptest.NewRecorder()

	h.listInvites(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBusinessHandler_DeleteInvite_InvalidInviteID(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/businesses?id=65", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 27))
	req.SetPathValue("inviteId", "bad")
	rr := httptest.NewRecorder()

	h.revokeInvite(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBusinessHandler_CreateDomain_ValidationError(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses?id=66", bytes.NewBufferString(`{"domain":""}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 28))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.addDomain(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBusinessHandler_VerifyDomain_Forbidden(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	mockBusiness.On("GetUserRole", mock.Anything, int64(67), int64(29)).Return(usecase.BusinessRoleMember, nil)

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses?id=67", bytes.NewBufferString(`{"verification_token":"abc"}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 29))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.verifyDomain(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_UpdateDomainAutoJoin_InvalidDomainID(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPut, "/businesses?id=68", bytes.NewBufferString(`{"auto_join_enabled":true}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 30))
	req.SetPathValue("domainId", "bad")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.toggleDomainAutoJoin(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestBusinessHandler_CreateBusiness_ValidationError(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := NewBusinessHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/businesses", bytes.NewBufferString(`{"name":"A","slug":"a","email":"bad"}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), 31))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.create(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

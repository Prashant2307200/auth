package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_GetById_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "profile_pic", "role", "created_at", "updated_at"}).AddRow(1, "u1", "e@e.com", "p", "pic", 0, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, profile_pic, role, created_at, updated_at ")).WithArgs(1).WillReturnRows(rows)

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	got, err := r.GetById(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), got.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByEmail_Empty(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	r, err := NewUserRepo(db)
	require.NoError(t, err)

	_, err = r.GetByEmail(context.Background(), "")
	require.Error(t, err)
}

func TestUserRepo_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	u := &entity.User{Username: "u", Email: "e@e.com", Password: "p", ProfilePic: "pic", Role: 0}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (username, email, password, profile_pic, role, created_at, updated_at)")).WithArgs(u.Username, u.Email, u.Password, u.ProfilePic, u.Role).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	id, err := r.Create(context.Background(), u)
	require.NoError(t, err)
	require.Equal(t, int64(42), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_UpdatePassword_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2")).WithArgs("h", 99).WillReturnResult(sqlmock.NewResult(0, 0))

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	err = r.UpdatePassword(context.Background(), 99, "h")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

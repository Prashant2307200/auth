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

func TestUserRepo_List_EmptyAndMultiple(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, profile_pic, role, created_at, updated_at ")).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password", "profile_pic", "role", "created_at", "updated_at"}))

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	users, err := r.List(context.Background())
	require.NoError(t, err)
	require.Len(t, users, 0)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "profile_pic", "role", "created_at", "updated_at"}).AddRow(1, "a", "a@a.com", "p", "pic", 0, now, now).AddRow(2, "b", "b@b.com", "p2", "pic2", 0, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, profile_pic, role, created_at, updated_at ")).WillReturnRows(rows)

	users, err = r.List(context.Background())
	require.NoError(t, err)
	require.Len(t, users, 2)
	require.Equal(t, "a", users[0].Username)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_UpdateById_DeleteById_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	u := &entity.User{Username: "u", Email: "e@e.com", Password: "p", ProfilePic: "pic", Role: 0}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).WithArgs(u.Username, u.Email, u.Password, u.ProfilePic, u.Role, 99).WillReturnResult(sqlmock.NewResult(0, 0))

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	err = r.UpdateById(context.Background(), 99, u)
	require.Error(t, err)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).WithArgs(100).WillReturnResult(sqlmock.NewResult(0, 0))
	err = r.DeleteById(context.Background(), 100)
	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

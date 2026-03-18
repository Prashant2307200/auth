package repository

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_UpdateById_DeleteById_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	u := &entity.User{Username: "u", Email: "e@e.com", Password: "p", ProfilePic: "pic", Role: 0}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).WithArgs(u.Username, u.Email, u.Password, u.ProfilePic, u.Role, 99).WillReturnResult(sqlmock.NewResult(0, 1))

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	err = r.UpdateById(context.Background(), 99, u)
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).WithArgs(100).WillReturnResult(sqlmock.NewResult(0, 1))
	err = r.DeleteById(context.Background(), 100)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_Create_ErrorVariant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	u := &entity.User{Username: "a", Email: "a@a.com", Password: "p", ProfilePic: "pic", Role: 0}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (username, email, password, profile_pic, role, created_at, updated_at)")).WithArgs(u.Username, u.Email, u.Password, u.ProfilePic, u.Role).WillReturnError(sqlmock.ErrCancelled)

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	_, err = r.Create(context.Background(), u)
	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

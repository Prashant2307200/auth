package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_GetByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "username", "email", "password", "profile_pic", "role", "created_at", "updated_at"}).AddRow(2, "bob", "b@b.com", "h", "pic", 0, time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, profile_pic, role, created_at, updated_at ")).WithArgs("b@b.com").WillReturnRows(rows)

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	u, err := r.GetByEmail(context.Background(), "b@b.com")
	require.NoError(t, err)
	require.Equal(t, int64(2), u.ID)
	require.Equal(t, "b@b.com", u.Email)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByEmail_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, password, profile_pic, role, created_at, updated_at ")).WithArgs("nope").WillReturnError(sqlmock.ErrCancelled)

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	_, err = r.GetByEmail(context.Background(), "nope")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

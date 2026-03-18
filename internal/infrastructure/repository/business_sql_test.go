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

func TestBusinessRepo_Create_GetById(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	b := &entity.Business{Name: "Acme", Slug: "acme", Email: "a@a.com", OwnerID: 1}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO businesses (name, slug, email, owner_id, signup_policy, created_at, updated_at)")).WithArgs(b.Name, b.Slug, b.Email, b.OwnerID, "closed").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, slug, email, owner_id, COALESCE(signup_policy, 'closed'), created_at, updated_at")).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "email", "owner_id", "signup_policy", "created_at", "updated_at"}).AddRow(7, "Acme", "acme", "a@a.com", 1, "closed", time.Now(), time.Now()))

	r, err := NewBusinessRepo(db)
	require.NoError(t, err)

	id, err := r.Create(context.Background(), b)
	require.NoError(t, err)
	require.Equal(t, int64(7), id)

	got, err := r.GetById(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "acme", got.Slug)
	require.NoError(t, mock.ExpectationsWereMet())
}

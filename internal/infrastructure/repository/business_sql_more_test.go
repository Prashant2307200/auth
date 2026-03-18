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

func TestBusinessRepo_GetBySlug_Update_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, slug, email, owner_id, COALESCE(signup_policy, 'closed'), created_at, updated_at")).WithArgs("acme").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "email", "owner_id", "signup_policy", "created_at", "updated_at"}).AddRow(8, "Acme", "acme", "a@a.com", 1, "closed", time.Now(), time.Now()))

	r, err := NewBusinessRepo(db)
	require.NoError(t, err)

	got, err := r.GetBySlug(context.Background(), "acme")
	require.NoError(t, err)
	require.Equal(t, "acme", got.Slug)

	upd := &entity.Business{Name: "New", Slug: "new", Email: "n@n.com", SignupPolicy: "closed"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE businesses")).WithArgs(upd.Name, upd.Slug, upd.Email, upd.SignupPolicy, 99).WillReturnResult(sqlmock.NewResult(0, 0))
	err = r.Update(context.Background(), 99, upd)
	require.Error(t, err)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM businesses WHERE id = $1")).WithArgs(100).WillReturnResult(sqlmock.NewResult(0, 0))
	err = r.Delete(context.Background(), 100)
	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

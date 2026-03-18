package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestBusinessRepo_Update_Delete_GetUsers_GetUserRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	r, err := NewBusinessRepo(db)
	require.NoError(t, err)

	b1 := &entity.Business{Name: "name", Slug: "slug", Email: "e@e.com", SignupPolicy: "closed"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE businesses")).WithArgs(b1.Name, b1.Slug, b1.Email, b1.SignupPolicy, 10).WillReturnResult(sqlmock.NewResult(0, 1))
	err = r.Update(context.Background(), 10, b1)
	require.NoError(t, err)

	b2 := &entity.Business{Name: "name2", Slug: "slug2", Email: "e2@e.com", SignupPolicy: "closed"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE businesses")).WithArgs(b2.Name, b2.Slug, b2.Email, b2.SignupPolicy, 11).WillReturnResult(sqlmock.NewResult(0, 0))
	err = r.Update(context.Background(), 11, b2)
	require.Error(t, err)

	// Delete success
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM businesses WHERE id = $1")).WithArgs(20).WillReturnResult(sqlmock.NewResult(0, 1))
	err = r.Delete(context.Background(), 20)
	require.NoError(t, err)

	// Delete not found
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM businesses WHERE id = $1")).WithArgs(21).WillReturnResult(sqlmock.NewResult(0, 0))
	err = r.Delete(context.Background(), 21)
	require.Error(t, err)

	// GetUsers success
	cols := []string{"id", "username", "email", "password", "profile_pic", "role", "created_at", "updated_at"}
	now := time.Now()
	rows := sqlmock.NewRows(cols).AddRow(int64(5), "u1", "u1@e.com", "p", "pic", 0, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.id, u.username, u.email, u.password, u.profile_pic, u.role, u.created_at, u.updated_at")).WithArgs(30).WillReturnRows(rows)

	users, err := r.GetUsers(context.Background(), 30)
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, int64(5), users[0].ID)
	require.NotNil(t, users[0].BusinessID)
	require.Equal(t, int64(30), *users[0].BusinessID)

	// GetUserRole success
	rowsRole := sqlmock.NewRows([]string{"role"}).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT role")).WithArgs(40, 50).WillReturnRows(rowsRole)
	role, err := r.GetUserRole(context.Background(), 40, 50)
	require.NoError(t, err)
	require.Equal(t, 2, role)

	// GetUserRole not found -> Expect sql.ErrNoRows
	mock.ExpectQuery(regexp.QuoteMeta("SELECT role")).WithArgs(41, 51).WillReturnError(sql.ErrNoRows)
	_, err = r.GetUserRole(context.Background(), 41, 51)
	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

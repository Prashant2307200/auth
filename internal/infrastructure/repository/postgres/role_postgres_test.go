package repository

import (
	"context"
	// driver import removed; use sqlmock result constructors instead
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestRolePostgres_Create_Get_List_Update_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now()

	// Expect create
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO roles (business_id, name, permissions, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id")).WithArgs(int64(10), "admin", "[]").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))

	rp, err := NewRolePostgres(db)
	require.NoError(t, err)

	id, err := rp.Create(context.Background(), &entity.Role{BusinessID: 10, Name: "admin", Permissions: []string{}})
	require.NoError(t, err)
	require.Equal(t, int64(100), id)

	// Expect get by id
	rows := sqlmock.NewRows([]string{"id", "business_id", "name", "permissions", "created_at", "updated_at"}).AddRow(100, 10, "admin", "", now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, name, permissions, created_at, updated_at FROM roles WHERE id = $1")).WithArgs(int64(100)).WillReturnRows(rows)

	r, err := rp.GetByID(context.Background(), 100)
	require.NoError(t, err)
	require.Equal(t, int64(100), r.ID)

	// Expect list
	listRows := sqlmock.NewRows([]string{"id", "business_id", "name", "permissions", "created_at", "updated_at"}).AddRow(100, 10, "admin", "", now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, name, permissions, created_at, updated_at FROM roles WHERE business_id = $1 ORDER BY name ASC")).WithArgs(int64(10)).WillReturnRows(listRows)

	list, err := rp.ListByBusiness(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, list, 1)

	// Expect update
	// Expect Update to receive JSON serialized permissions
	mock.ExpectExec(regexp.QuoteMeta("UPDATE roles SET name = $1, permissions = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND business_id = $4")).WithArgs("owner", "[\"p\"]", int64(100), int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))

	err = rp.Update(context.Background(), &entity.Role{ID: 100, BusinessID: 10, Name: "owner", Permissions: []string{"p"}})
	require.NoError(t, err)

	// Expect delete
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM roles WHERE id = $1")).WithArgs(int64(100)).WillReturnResult(sqlmock.NewResult(0, 1))

	err = rp.Delete(context.Background(), 100)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRolePostgres_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, name, permissions, created_at, updated_at FROM roles WHERE id = $1")).WithArgs(int64(999)).WillReturnError(sqlmock.ErrCancelled)

	rp, err := NewRolePostgres(db)
	require.NoError(t, err)

	_, err = rp.GetByID(context.Background(), 999)
	require.Error(t, err)
}

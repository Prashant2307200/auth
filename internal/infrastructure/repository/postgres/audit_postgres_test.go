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

func TestAuditPostgres_Log_GetByID_List_Export(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo, err := NewAuditPostgres(db)
	require.NoError(t, err)

	now := time.Now()

	// Log: expect INSERT
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO audit_logs (business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at, updated_at)")).WithArgs(int64(10), int64(20), "user.created", "user", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "1.2.3.4", "ua").WillReturnResult(sqlmock.NewResult(1, 1))

	a := &entity.AuditLog{
		BusinessID: 10,
		UserID:     20,
		Action:     "user.created",
		EntityType: "user",
		IPAddress:  "1.2.3.4",
		UserAgent:  "ua",
	}
	err = repo.Log(context.Background(), a)
	require.NoError(t, err)

	// GetByID: expect SELECT and return row
	rows := sqlmock.NewRows([]string{"id", "business_id", "user_id", "action", "entity_type", "entity_id", "old_values", "new_values", "ip_address", "user_agent", "created_at"}).AddRow(1, 10, 20, "user.created", "user", nil, "{}", "{}", "1.2.3.4", "ua", now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE id = $1")).WithArgs(int64(1)).WillReturnRows(rows)

	got, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), got.ID)
	require.Equal(t, int64(10), got.BusinessID)
	require.Equal(t, int64(20), got.UserID)

	// ListByBusiness pagination: expect args limit, offset
	rows2 := sqlmock.NewRows([]string{"id", "business_id", "user_id", "action", "entity_type", "entity_id", "old_values", "new_values", "ip_address", "user_agent", "created_at"})
	for i := 0; i < 3; i++ {
		rows2.AddRow(int64(i+2), 10, 20+i, "action", "user", nil, "{}", "{}", "1.2.3.4", "ua", now)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE business_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3")).WithArgs(int64(10), 3, 0).WillReturnRows(rows2)

	list, err := repo.ListByBusiness(context.Background(), 10, 3, 0)
	require.NoError(t, err)
	require.Len(t, list, 3)

	// ListByUser pagination
	rows3 := sqlmock.NewRows([]string{"id", "business_id", "user_id", "action", "entity_type", "entity_id", "old_values", "new_values", "ip_address", "user_agent", "created_at"})
	rows3.AddRow(int64(99), 10, 20, "action", "user", nil, "{}", "{}", "1.2.3.4", "ua", now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE business_id = $1 AND user_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4")).WithArgs(int64(10), int64(20), 1, 0).WillReturnRows(rows3)

	byUser, err := repo.ListByUser(context.Background(), 10, 20, 1, 0)
	require.NoError(t, err)
	require.Len(t, byUser, 1)

	// Export returns all rows (no limit)
	rows4 := sqlmock.NewRows([]string{"id", "business_id", "user_id", "action", "entity_type", "entity_id", "old_values", "new_values", "ip_address", "user_agent", "created_at"})
	rows4.AddRow(int64(1), 10, 20, "a", "user", nil, "{}", "{}", "1.2.3.4", "ua", now)
	rows4.AddRow(int64(2), 10, 21, "b", "user", nil, "{}", "{}", "1.2.3.4", "ua", now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE business_id = $1 ORDER BY created_at ASC")).WithArgs(int64(10)).WillReturnRows(rows4)

	exp, err := repo.Export(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, exp, 2)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuditPostgres_Immutability_NoUpdateDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo, err := NewAuditPostgres(db)
	require.NoError(t, err)

	// Expect only INSERT and SELECT queries; if code issues UPDATE/DELETE tests will fail due to unexpected query
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO audit_logs (business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at, updated_at)")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE id = $1")).WillReturnRows(sqlmock.NewRows([]string{"id", "business_id", "user_id", "action", "entity_type", "entity_id", "old_values", "new_values", "ip_address", "user_agent", "created_at"}).AddRow(1, 1, 1, "a", "user", nil, "{}", "{}", "", "", time.Now()))

	// Call Log and GetByID; the repository must not perform any UPDATE or DELETE operations
	err = repo.Log(context.Background(), &entity.AuditLog{BusinessID: 1, UserID: 1, Action: "a", EntityType: "user"})
	require.NoError(t, err)
	_, err = repo.GetByID(context.Background(), 1)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

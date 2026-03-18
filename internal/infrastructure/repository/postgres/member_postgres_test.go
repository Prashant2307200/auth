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

func TestMemberPostgres_CRUD(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mp, err := NewMemberPostgres(db)
	require.NoError(t, err)

	now := time.Now()

	// Create
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO business_members (business_id, user_id, email, role_id, status, invited_by, invited_at, created_at, updated_at)")).WithArgs(int64(1), sqlmock.AnyArg(), "e@example.com", int64(2), entity.MemberStatusPending, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(10, now, now))

	invitedBy := int64(3)
	m := &entity.BusinessMember{BusinessID: 1, Email: "e@example.com", RoleID: 2, Status: entity.MemberStatusPending, InvitedBy: &invitedBy, InvitedAt: now}
	err = mp.Create(context.Background(), m)
	require.NoError(t, err)
	require.Equal(t, int64(10), m.ID)

	// GetByID
	rows := sqlmock.NewRows([]string{"id", "business_id", "user_id", "email", "role_id", "status", "invited_by", "invited_at", "accepted_at", "created_at", "updated_at"}).AddRow(10, 1, nil, "e@example.com", 2, entity.MemberStatusPending, 3, now, nil, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, user_id, email, role_id, status, invited_by, invited_at, accepted_at, created_at, updated_at FROM business_members WHERE id = $1")).WithArgs(int64(10)).WillReturnRows(rows)

	got, err := mp.GetByID(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, int64(10), got.ID)

	// ListByBusiness
	listRows := sqlmock.NewRows([]string{"id", "business_id", "user_id", "email", "role_id", "status", "invited_by", "invited_at", "accepted_at", "created_at", "updated_at"}).AddRow(10, 1, nil, "e@example.com", 2, entity.MemberStatusPending, 3, now, nil, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, user_id, email, role_id, status, invited_by, invited_at, accepted_at, created_at, updated_at FROM business_members WHERE business_id = $1 ORDER BY created_at ASC")).WithArgs(int64(1)).WillReturnRows(listRows)

	list, err := mp.ListByBusiness(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, list, 1)

	// Update
	mock.ExpectExec(regexp.QuoteMeta("UPDATE business_members SET user_id = $1, email = $2, role_id = $3, status = $4, invited_by = $5, invited_at = $6, accepted_at = $7, updated_at = NOW() WHERE id = $8 AND business_id = $9")).WithArgs(sqlmock.AnyArg(), "e@example.com", int64(2), entity.MemberStatusActive, int64(3), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(10), int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	got.Status = entity.MemberStatusActive
	err = mp.Update(context.Background(), got)
	require.NoError(t, err)

	// Delete
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM business_members WHERE id = $1")).WithArgs(int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))
	err = mp.Delete(context.Background(), 10)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMemberPostgres_ListByUser_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mp, err := NewMemberPostgres(db)
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, business_id, user_id, email, role_id, status, invited_by, invited_at, accepted_at, created_at, updated_at FROM business_members WHERE user_id = $1 ORDER BY created_at ASC")).WithArgs(int64(999)).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	list, err := mp.ListByUser(context.Background(), 999)
	require.NoError(t, err)
	require.Len(t, list, 0)
	require.NoError(t, mock.ExpectationsWereMet())
}

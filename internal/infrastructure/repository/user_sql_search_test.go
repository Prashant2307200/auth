package repository

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_Search_EmptyAndMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, profile_pic")).WithArgs(int64(1), "", "%"+""+"%", "%"+""+"%").WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "profile_pic"}))

	r, err := NewUserRepo(db)
	require.NoError(t, err)

	users, err := r.Search(context.Background(), 1, "")
	require.NoError(t, err)
	require.Len(t, users, 0)

	rows := sqlmock.NewRows([]string{"id", "username", "email", "profile_pic"}).AddRow(3, "alice", "a@a.com", "pic")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, profile_pic")).WithArgs(int64(1), "ali", "%ali%", "%ali%").WillReturnRows(rows)

	users, err = r.Search(context.Background(), 1, "ali")
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, "alice", users[0].Username)

	require.NoError(t, mock.ExpectationsWereMet())
}

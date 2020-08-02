package user_test

import (
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sf9v/nero/sort"
	"github.com/sf9v/nero/test/integration/gen/user"
)

func TestSorts(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		sfs := []user.SortFunc{
			user.IDAsc(), user.IDDesc(),
		}

		sb := &sort.Sorts{}
		for _, sf := range sfs {
			sf(sb)
		}

		qb := sq.Select("id").From("users")
		for _, s := range sb.All() {
			qb = addSorts(qb, s)
		}

		got, _, err := qb.ToSql()
		require.NoError(t, err)
		expect := "SELECT id FROM users ORDER BY id ASC, id DESC"
		assert.Equal(t, expect, got)
	})

	t.Run("Email", func(t *testing.T) {
		sfs := []user.SortFunc{
			user.EmailAsc(), user.EmailDesc(),
		}

		sb := &sort.Sorts{}
		for _, sf := range sfs {
			sf(sb)
		}

		qb := sq.Select("email").From("users")
		for _, s := range sb.All() {
			qb = addSorts(qb, s)
		}

		got, _, err := qb.ToSql()
		require.NoError(t, err)
		expect := "SELECT email FROM users ORDER BY email ASC, email DESC"
		assert.Equal(t, expect, got)
	})

	t.Run("Name", func(t *testing.T) {
		sfs := []user.SortFunc{
			user.NameAsc(), user.NameDesc(),
		}

		sb := &sort.Sorts{}
		for _, sf := range sfs {
			sf(sb)
		}

		qb := sq.Select("name").From("users")
		for _, s := range sb.All() {
			qb = addSorts(qb, s)
		}

		got, _, err := qb.ToSql()
		require.NoError(t, err)
		expect := "SELECT name FROM users ORDER BY name ASC, name DESC"
		assert.Equal(t, expect, got)
	})

	t.Run("UpdatedAt", func(t *testing.T) {
		sfs := []user.SortFunc{
			user.UpdatedAtAsc(), user.UpdatedAtDesc(),
		}

		sb := &sort.Sorts{}
		for _, sf := range sfs {
			sf(sb)
		}

		qb := sq.Select("updated_at").From("users")
		for _, s := range sb.All() {
			qb = addSorts(qb, s)
		}

		got, _, err := qb.ToSql()
		require.NoError(t, err)
		expect := "SELECT updated_at FROM users ORDER BY updated_at ASC, updated_at DESC"
		assert.Equal(t, expect, got)
	})

	t.Run("CreatedAt", func(t *testing.T) {
		sfs := []user.SortFunc{
			user.CreatedAtAsc(), user.CreatedAtDesc(),
		}

		sb := &sort.Sorts{}
		for _, sf := range sfs {
			sf(sb)
		}

		qb := sq.Select("created_at").From("users")
		for _, s := range sb.All() {
			qb = addSorts(qb, s)
		}

		got, _, err := qb.ToSql()
		require.NoError(t, err)
		expect := "SELECT created_at FROM users ORDER BY created_at ASC, created_at DESC"
		assert.Equal(t, expect, got)
	})
}

func addSorts(sb sq.SelectBuilder, s *sort.Sort) sq.SelectBuilder {
	switch s.Direction {
	case sort.Asc:
		return sb.OrderBy(fmt.Sprintf("%s ASC", s.Field))
	case sort.Desc:
		return sb.OrderBy(fmt.Sprintf("%s DESC", s.Field))
	}

	return sb
}

package gen

import (
	"fmt"
	"strings"
	"testing"

	gen "github.com/sf9v/nero/gen/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newPGRepo(t *testing.T) {
	schema, err := buildSchema(new(gen.Example))
	require.NoError(t, err)
	require.NotNil(t, schema)

	pgRepo := newPGRepository(schema)
	expect := strings.TrimSpace(`
type PGRepository struct {
	db *sql.DB
}

var _ = Repository(&PGRepository{})

func NewPGRepository(db *sql.DB) *PGRepository {
	return &PGRepository{
		db: db,
	}
}

func (s *PGRepository) Tx(ctx context.Context) (nero.Tx, error) {
	return s.db.BeginTx(ctx, &sql.TxOptions{})
}

func (s *PGRepository) Create(ctx context.Context, c *Creator) (int64, error) {
	tx, err := s.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil && tx != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Wrapf(err, "rollback error: %v", rollbackErr)
			}
			return
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "commit error")
		}
	}()

	id, err := s.CreateTx(ctx, tx, c)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *PGRepository) Query(ctx context.Context, q *Queryer) ([]*internal.Example, error) {
	tx, err := s.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil && tx != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Wrapf(err, "rollback error: %v", rollbackErr)
			}
			return
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "commit error")
		}
	}()

	list, err := s.QueryTx(ctx, tx, q)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *PGRepository) Update(ctx context.Context, u *Updater) (int64, error) {
	tx, err := s.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil && tx != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Wrapf(err, "rollback error: %v", rollbackErr)
			}
			return
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "commit error")
		}
	}()

	rowsAffected, err := s.UpdateTx(ctx, tx, u)
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (s *PGRepository) Delete(ctx context.Context, d *Deleter) (int64, error) {
	tx, err := s.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil && tx != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Wrapf(err, "rollback error: %v", rollbackErr)
			}
			return
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "commit error")
		}
	}()

	rowsAffected, err := s.DeleteTx(ctx, tx, d)
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (s *PGRepository) CreateTx(ctx context.Context, tx nero.Tx, c *Creator) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	qb := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Insert(c.collection).
		Columns(c.columns...).
		Values(c.name, c.updatedAt).
		Suffix("RETURNING \"id\"").
		RunWith(txx)
	var id int64
	err := qb.QueryRowContext(ctx).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *PGRepository) QueryTx(ctx context.Context, tx nero.Tx, q *Queryer) ([]*internal.Example, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errors.New("expecting tx to be *sql.Tx")
	}

	pb := &predicate.Builder{}
	for _, pf := range q.pfs {
		pf(pb)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	sb := psql.Select(q.columns...).From(q.collection)
	for _, p := range pb.Predicates() {
		switch p.Op {
		case predicate.Eq:
			sb = sb.Where(squirrel.Eq{
				p.Field: p.Val,
			})
		case predicate.NotEq:
			sb = sb.Where(squirrel.NotEq{
				p.Field: p.Val,
			})
		case predicate.Gt:
			sb = sb.Where(squirrel.Gt{
				p.Field: p.Val,
			})
		case predicate.GtOrEq:
			sb = sb.Where(squirrel.GtOrEq{
				p.Field: p.Val,
			})
		case predicate.Lt:
			sb = sb.Where(squirrel.Lt{
				p.Field: p.Val,
			})
		case predicate.LtOrEq:
			sb = sb.Where(squirrel.LtOrEq{
				p.Field: p.Val,
			})
		}
	}

	if q.limit > 0 {
		sb = sb.Limit(q.limit)
	}

	if q.offset > 0 {
		sb = sb.Offset(q.offset)
	}

	sql, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}

	stmnt, err := txx.PrepareContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	rows, err := stmnt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*internal.Example{}
	for rows.Next() {
		var item internal.Example
		err = rows.Scan(
			&item.ID,
			&item.Name,
			&item.UpdatedAt,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		list = append(list, &item)
	}

	return list, nil
}

func (s *PGRepository) UpdateTx(ctx context.Context, tx nero.Tx, u *Updater) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	pb := &predicate.Builder{}
	for _, pf := range u.pfs {
		pf(pb)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	sb := psql.Update(u.collection).Set("name", u.name).Set("updated_at", u.updatedAt)
	for _, p := range pb.Predicates() {
		switch p.Op {
		case predicate.Eq:
			sb = sb.Where(squirrel.Eq{
				p.Field: p.Val,
			})
		case predicate.NotEq:
			sb = sb.Where(squirrel.NotEq{
				p.Field: p.Val,
			})
		case predicate.Gt:
			sb = sb.Where(squirrel.Gt{
				p.Field: p.Val,
			})
		case predicate.GtOrEq:
			sb = sb.Where(squirrel.GtOrEq{
				p.Field: p.Val,
			})
		case predicate.Lt:
			sb = sb.Where(squirrel.Lt{
				p.Field: p.Val,
			})
		case predicate.LtOrEq:
			sb = sb.Where(squirrel.LtOrEq{
				p.Field: p.Val,
			})
		}
	}

	sql, args, err := sb.ToSql()
	if err != nil {
		return 0, err
	}

	stmnt, err := txx.PrepareContext(ctx, sql)
	if err != nil {
		return 0, err
	}

	res, err := stmnt.ExecContext(ctx, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (s *PGRepository) DeleteTx(ctx context.Context, tx nero.Tx, d *Deleter) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	pb := &predicate.Builder{}
	for _, pf := range d.pfs {
		pf(pb)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	sb := psql.Delete(d.collection)
	for _, p := range pb.Predicates() {
		switch p.Op {
		case predicate.Eq:
			sb = sb.Where(squirrel.Eq{
				p.Field: p.Val,
			})
		case predicate.NotEq:
			sb = sb.Where(squirrel.NotEq{
				p.Field: p.Val,
			})
		case predicate.Gt:
			sb = sb.Where(squirrel.Gt{
				p.Field: p.Val,
			})
		case predicate.GtOrEq:
			sb = sb.Where(squirrel.GtOrEq{
				p.Field: p.Val,
			})
		case predicate.Lt:
			sb = sb.Where(squirrel.Lt{
				p.Field: p.Val,
			})
		case predicate.LtOrEq:
			sb = sb.Where(squirrel.LtOrEq{
				p.Field: p.Val,
			})
		}
	}

	sql, args, err := sb.ToSql()
	if err != nil {
		return 0, err
	}

	stmnt, err := txx.PrepareContext(ctx, sql)
	if err != nil {
		return 0, err
	}

	res, err := stmnt.ExecContext(ctx, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
`)

	got := strings.TrimSpace(fmt.Sprintf("%#v", pgRepo))
	assert.Equal(t, expect, got)
}

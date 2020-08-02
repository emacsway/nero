package pg

import (
	"fmt"
	"strings"
	"testing"

	gen "github.com/sf9v/nero/gen/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPGRepoC(t *testing.T) {
	schema, err := gen.BuildSchema(new(gen.Example))
	require.NoError(t, err)
	require.NotNil(t, schema)

	pgRepo := NewPGRepoC(schema)
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

func (pgr *PGRepository) Tx(ctx context.Context) (nero.Tx, error) {
	return pgr.db.BeginTx(ctx, nil)
}

func (pgr *PGRepository) Create(ctx context.Context, c *Creator) (int64, error) {
	tx, err := pgr.Tx(ctx)
	if err != nil {
		return 0, err
	}

	id, err := pgr.CreateTx(ctx, tx, c)
	if err != nil {
		return 0, rollback(tx, err)
	}

	return id, tx.Commit()
}

func (pgr *PGRepository) CreateMany(ctx context.Context, cs ...*Creator) error {
	tx, err := pgr.Tx(ctx)
	if err != nil {
		return err
	}

	err = pgr.CreateManyTx(ctx, tx, cs...)
	if err != nil {
		return rollback(tx, err)
	}

	return tx.Commit()
}

func (pgr *PGRepository) CreateTx(ctx context.Context, tx nero.Tx, c *Creator) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	qb := squirrel.Insert(c.collection).
		Columns(c.columns...).
		Values(c.name, c.updatedAt).
		Suffix("RETURNING \"id\"").
		PlaceholderFormat(squirrel.Dollar).
		RunWith(txx)
	var id int64
	err := qb.QueryRowContext(ctx).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (pgr *PGRepository) CreateManyTx(ctx context.Context, tx nero.Tx, cs ...*Creator) error {
	if len(cs) == 0 {
		return nil
	}

	txx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("expecting tx to be *sql.Tx")
	}

	qb := squirrel.Insert(cs[0].collection).
		Columns(cs[0].columns...)
	for _, c := range cs {
		qb = qb.Values(c.name, c.updatedAt)
	}

	qb = qb.Suffix("RETURNING \"id\"").
		PlaceholderFormat(squirrel.Dollar)
	_, err := qb.RunWith(txx).ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (pgr *PGRepository) Query(ctx context.Context, q *Queryer) ([]*internal.Example, error) {
	tx, err := pgr.Tx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := pgr.QueryTx(ctx, tx, q)
	if err != nil {
		return nil, rollback(tx, err)
	}

	return list, tx.Commit()
}

func (pgr *PGRepository) QueryOne(ctx context.Context, q *Queryer) (*internal.Example, error) {
	tx, err := pgr.Tx(ctx)
	if err != nil {
		return nil, err
	}

	item, err := pgr.QueryOneTx(ctx, tx, q)
	if err != nil {
		return nil, rollback(tx, err)
	}

	return item, tx.Commit()
}

func (pgr *PGRepository) QueryTx(ctx context.Context, tx nero.Tx, q *Queryer) ([]*internal.Example, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errors.New("expecting tx to be *sql.Tx")
	}

	qb := pgr.buildSelect(q)
	rows, err := qb.RunWith(txx).QueryContext(ctx)
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

func (pgr *PGRepository) QueryOneTx(ctx context.Context, tx nero.Tx, q *Queryer) (*internal.Example, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errors.New("expecting tx to be *sql.Tx")
	}

	qb := pgr.buildSelect(q)
	row := qb.RunWith(txx).QueryRowContext(ctx)

	var item internal.Example
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.UpdatedAt,
		&item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (pgr *PGRepository) buildSelect(q *Queryer) squirrel.SelectBuilder {
	qb := squirrel.Select(q.columns...).
		From(q.collection).
		PlaceholderFormat(squirrel.Dollar)

	pb := &predicate.Predicates{}
	for _, pf := range q.pfs {
		pf(pb)
	}
	for _, p := range pb.All() {
		switch p.Op {
		case predicate.Eq:
			qb = qb.Where(squirrel.Eq{
				p.Field: p.Val,
			})
		case predicate.NotEq:
			qb = qb.Where(squirrel.NotEq{
				p.Field: p.Val,
			})
		case predicate.Gt:
			qb = qb.Where(squirrel.Gt{
				p.Field: p.Val,
			})
		case predicate.GtOrEq:
			qb = qb.Where(squirrel.GtOrEq{
				p.Field: p.Val,
			})
		case predicate.Lt:
			qb = qb.Where(squirrel.Lt{
				p.Field: p.Val,
			})
		case predicate.LtOrEq:
			qb = qb.Where(squirrel.LtOrEq{
				p.Field: p.Val,
			})
		}
	}

	sb := &sort.Sorts{}
	for _, sf := range q.sfs {
		sf(sb)
	}
	for _, s := range sb.All() {
		switch s.Direction {
		case sort.Asc:
			qb = qb.OrderBy(fmt.Sprintf("%s ASC", s.Field))
		case sort.Desc:
			qb = qb.OrderBy(fmt.Sprintf("%s DESC", s.Field))
		}
	}

	if q.limit > 0 {
		qb = qb.Limit(q.limit)
	}

	if q.offset > 0 {
		qb = qb.Offset(q.offset)
	}

	return qb
}

func (pgr *PGRepository) Update(ctx context.Context, u *Updater) (int64, error) {
	tx, err := pgr.Tx(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := pgr.UpdateTx(ctx, tx, u)
	if err != nil {
		return 0, rollback(tx, err)
	}

	return rowsAffected, tx.Commit()
}

func (pgr *PGRepository) UpdateTx(ctx context.Context, tx nero.Tx, u *Updater) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	pb := &predicate.Predicates{}
	for _, pf := range u.pfs {
		pf(pb)
	}

	qb := squirrel.Update(u.collection).
		Set("name", u.name).
		Set("updated_at", u.updatedAt).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(txx)
	for _, p := range pb.All() {
		switch p.Op {
		case predicate.Eq:
			qb = qb.Where(squirrel.Eq{
				p.Field: p.Val,
			})
		case predicate.NotEq:
			qb = qb.Where(squirrel.NotEq{
				p.Field: p.Val,
			})
		case predicate.Gt:
			qb = qb.Where(squirrel.Gt{
				p.Field: p.Val,
			})
		case predicate.GtOrEq:
			qb = qb.Where(squirrel.GtOrEq{
				p.Field: p.Val,
			})
		case predicate.Lt:
			qb = qb.Where(squirrel.Lt{
				p.Field: p.Val,
			})
		case predicate.LtOrEq:
			qb = qb.Where(squirrel.LtOrEq{
				p.Field: p.Val,
			})
		}
	}

	res, err := qb.ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (pgr *PGRepository) Delete(ctx context.Context, d *Deleter) (int64, error) {
	tx, err := pgr.Tx(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := pgr.DeleteTx(ctx, tx, d)
	if err != nil {
		return 0, rollback(tx, err)
	}

	return rowsAffected, tx.Commit()
}

func (pgr *PGRepository) DeleteTx(ctx context.Context, tx nero.Tx, d *Deleter) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	pb := &predicate.Predicates{}
	for _, pf := range d.pfs {
		pf(pb)
	}

	qb := squirrel.Delete(d.collection).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(txx)
	for _, p := range pb.All() {
		switch p.Op {
		case predicate.Eq:
			qb = qb.Where(squirrel.Eq{
				p.Field: p.Val,
			})
		case predicate.NotEq:
			qb = qb.Where(squirrel.NotEq{
				p.Field: p.Val,
			})
		case predicate.Gt:
			qb = qb.Where(squirrel.Gt{
				p.Field: p.Val,
			})
		case predicate.GtOrEq:
			qb = qb.Where(squirrel.GtOrEq{
				p.Field: p.Val,
			})
		case predicate.Lt:
			qb = qb.Where(squirrel.Lt{
				p.Field: p.Val,
			})
		case predicate.LtOrEq:
			qb = qb.Where(squirrel.LtOrEq{
				p.Field: p.Val,
			})
		}
	}

	res, err := qb.ExecContext(ctx)
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

package sqlite

import (
	"fmt"
	"strings"
	"testing"

	gen "github.com/sf9v/nero/gen/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLiteRepoCExample(t *testing.T) {
	schema, err := gen.BuildSchema(new(gen.Example))
	require.NoError(t, err)
	require.NotNil(t, schema)

	sqliteRepo := NewSQLiteRepoC(schema)
	expect := strings.TrimSpace(`
type SQLiteRepository struct {
	db *sql.DB
}

var _ = Repository(&SQLiteRepository{})

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (sqlir *SQLiteRepository) Tx(ctx context.Context) (nero.Tx, error) {
	return sqlir.db.BeginTx(ctx, nil)
}

func (sqlir *SQLiteRepository) Create(ctx context.Context, c *Creator) (int64, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return 0, err
	}

	id, err := sqlir.CreateTx(ctx, tx, c)
	if err != nil {
		return 0, rollback(tx, err)
	}

	return id, tx.Commit()
}

func (sqlir *SQLiteRepository) CreateMany(ctx context.Context, cs ...*Creator) error {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return err
	}

	err = sqlir.CreateManyTx(ctx, tx, cs...)
	if err != nil {
		return rollback(tx, err)
	}

	return tx.Commit()
}

func (sqlir *SQLiteRepository) CreateTx(ctx context.Context, tx nero.Tx, c *Creator) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	qb := squirrel.Insert(c.collection).
		Columns(c.columns...).
		Values(c.name, c.updatedAt).
		RunWith(txx)
	res, err := qb.ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (sqlir *SQLiteRepository) CreateManyTx(ctx context.Context, tx nero.Tx, cs ...*Creator) error {
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

	_, err := qb.RunWith(txx).ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (sqlir *SQLiteRepository) Query(ctx context.Context, q *Queryer) ([]*internal.Example, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := sqlir.QueryTx(ctx, tx, q)
	if err != nil {
		return nil, rollback(tx, err)
	}

	return list, tx.Commit()
}

func (sqlir *SQLiteRepository) QueryOne(ctx context.Context, q *Queryer) (*internal.Example, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return nil, err
	}

	item, err := sqlir.QueryOneTx(ctx, tx, q)
	if err != nil {
		return nil, rollback(tx, err)
	}

	return item, tx.Commit()
}

func (sqlir *SQLiteRepository) QueryTx(ctx context.Context, tx nero.Tx, q *Queryer) ([]*internal.Example, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errors.New("expecting tx to be *sql.Tx")
	}

	qb := sqlir.buildSelect(q)
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

func (sqlir *SQLiteRepository) QueryOneTx(ctx context.Context, tx nero.Tx, q *Queryer) (*internal.Example, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errors.New("expecting tx to be *sql.Tx")
	}

	qb := sqlir.buildSelect(q)
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

func (sqlir *SQLiteRepository) buildSelect(q *Queryer) squirrel.SelectBuilder {
	qb := squirrel.Select(q.columns...).
		From(q.collection)

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

func (sqlir *SQLiteRepository) Update(ctx context.Context, u *Updater) (int64, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := sqlir.UpdateTx(ctx, tx, u)
	if err != nil {
		return 0, rollback(tx, err)
	}

	return rowsAffected, tx.Commit()
}

func (sqlir *SQLiteRepository) UpdateTx(ctx context.Context, tx nero.Tx, u *Updater) (int64, error) {
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

func (sqlir *SQLiteRepository) Delete(ctx context.Context, d *Deleter) (int64, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := sqlir.DeleteTx(ctx, tx, d)
	if err != nil {
		return 0, rollback(tx, err)
	}

	return rowsAffected, tx.Commit()
}

func (sqlir *SQLiteRepository) DeleteTx(ctx context.Context, tx nero.Tx, d *Deleter) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	pb := &predicate.Predicates{}
	for _, pf := range d.pfs {
		pf(pb)
	}

	qb := squirrel.Delete(d.collection).
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

	got := strings.TrimSpace(fmt.Sprintf("%#v", sqliteRepo))
	assert.Equal(t, expect, got)
}

func TestNewSQLiteRepoCExample2(t *testing.T) {
	schema, err := gen.BuildSchema(new(gen.Example2))
	require.NoError(t, err)
	require.NotNil(t, schema)

	sqliteRepo := NewSQLiteRepoC(schema)
	expect := strings.TrimSpace(`
type SQLiteRepository struct {
	db *sql.DB
}

var _ = Repository(&SQLiteRepository{})

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (sqlir *SQLiteRepository) Tx(ctx context.Context) (nero.Tx, error) {
	return sqlir.db.BeginTx(ctx, nil)
}

func (sqlir *SQLiteRepository) Create(ctx context.Context, c *Creator) (string, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return "", err
	}

	id, err := sqlir.CreateTx(ctx, tx, c)
	if err != nil {
		return "", rollback(tx, err)
	}

	return id, tx.Commit()
}

func (sqlir *SQLiteRepository) CreateMany(ctx context.Context, cs ...*Creator) error {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return err
	}

	err = sqlir.CreateManyTx(ctx, tx, cs...)
	if err != nil {
		return rollback(tx, err)
	}

	return tx.Commit()
}

func (sqlir *SQLiteRepository) CreateTx(ctx context.Context, tx nero.Tx, c *Creator) (string, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return "", errors.New("expecting tx to be *sql.Tx")
	}

	qb := squirrel.Insert(c.collection).
		Columns(c.columns...).
		Values(c.name, c.updatedAt).
		RunWith(txx)
	res, err := qb.ExecContext(ctx)
	if err != nil {
		return "", err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(id, 10), nil
}

func (sqlir *SQLiteRepository) CreateManyTx(ctx context.Context, tx nero.Tx, cs ...*Creator) error {
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

	_, err := qb.RunWith(txx).ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (sqlir *SQLiteRepository) Query(ctx context.Context, q *Queryer) ([]*internal.Example2, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return nil, err
	}

	list, err := sqlir.QueryTx(ctx, tx, q)
	if err != nil {
		return nil, rollback(tx, err)
	}

	return list, tx.Commit()
}

func (sqlir *SQLiteRepository) QueryOne(ctx context.Context, q *Queryer) (*internal.Example2, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return nil, err
	}

	item, err := sqlir.QueryOneTx(ctx, tx, q)
	if err != nil {
		return nil, rollback(tx, err)
	}

	return item, tx.Commit()
}

func (sqlir *SQLiteRepository) QueryTx(ctx context.Context, tx nero.Tx, q *Queryer) ([]*internal.Example2, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errors.New("expecting tx to be *sql.Tx")
	}

	qb := sqlir.buildSelect(q)
	rows, err := qb.RunWith(txx).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*internal.Example2{}
	for rows.Next() {
		var item internal.Example2
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

func (sqlir *SQLiteRepository) QueryOneTx(ctx context.Context, tx nero.Tx, q *Queryer) (*internal.Example2, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errors.New("expecting tx to be *sql.Tx")
	}

	qb := sqlir.buildSelect(q)
	row := qb.RunWith(txx).QueryRowContext(ctx)

	var item internal.Example2
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

func (sqlir *SQLiteRepository) buildSelect(q *Queryer) squirrel.SelectBuilder {
	qb := squirrel.Select(q.columns...).
		From(q.collection)

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

func (sqlir *SQLiteRepository) Update(ctx context.Context, u *Updater) (int64, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := sqlir.UpdateTx(ctx, tx, u)
	if err != nil {
		return 0, rollback(tx, err)
	}

	return rowsAffected, tx.Commit()
}

func (sqlir *SQLiteRepository) UpdateTx(ctx context.Context, tx nero.Tx, u *Updater) (int64, error) {
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

func (sqlir *SQLiteRepository) Delete(ctx context.Context, d *Deleter) (int64, error) {
	tx, err := sqlir.Tx(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := sqlir.DeleteTx(ctx, tx, d)
	if err != nil {
		return 0, rollback(tx, err)
	}

	return rowsAffected, tx.Commit()
}

func (sqlir *SQLiteRepository) DeleteTx(ctx context.Context, tx nero.Tx, d *Deleter) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	pb := &predicate.Predicates{}
	for _, pf := range d.pfs {
		pf(pb)
	}

	qb := squirrel.Delete(d.collection).
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

	got := strings.TrimSpace(fmt.Sprintf("%#v", sqliteRepo))
	assert.Equal(t, expect, got)
}

package template

import "github.com/sf9v/nero"

// PostgresTemplate is the template for generating a postgres repository
type PostgresTemplate struct {
	filename string
}

var _ nero.Templater = (*PostgresTemplate)(nil)

// NewPostgresTemplate returns a new PostgresTemplate
func NewPostgresTemplate() *PostgresTemplate {
	return &PostgresTemplate{
		filename: "postgres.go",
	}
}

// WithFilename overrides the default filename
func (t *PostgresTemplate) WithFilename(filename string) *PostgresTemplate {
	t.filename = filename
	return t
}

// Filename returns the filename
func (t *PostgresTemplate) Filename() string {
	return t.filename
}

// Content returns the template content
func (t *PostgresTemplate) Content() string {
	return postgresTmpl
}

const postgresTmpl = `
// Code generated by nero, DO NOT EDIT.
package {{.Pkg}}

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"io"
	"strings"
	"log"
	"os"
	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sf9v/nero"
	"github.com/sf9v/nero/aggregate"
	"github.com/sf9v/nero/comparison"
	"github.com/sf9v/nero/sort"
	{{range $import := .SchemaImports -}}
		"{{$import}}"
	{{end -}}
	{{range $import := .ColumnImports -}}
		"{{$import}}"
	{{end -}}
)

// PostgresRepository implements the Repository interface
type PostgresRepository struct {
	db  *sql.DB
	logger nero.Logger
	debug bool
}

var _ Repository = (*PostgresRepository)(nil)

// NewPostgresRepository is a factory for PostgresRepository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// Debug enables debug mode
func (pg *PostgresRepository) Debug() *PostgresRepository {	
	return &PostgresRepository{
		db:  pg.db,	
		debug: true,
		logger: log.New(os.Stdout, "nero: ", 0),
	}
}

// WithLogger overrides the default logger
func (pg *PostgresRepository) WithLogger(logger nero.Logger) *PostgresRepository {	
	pg.logger = logger
	return pg
}

// Tx creates begins a new transaction
func (pg *PostgresRepository) Tx(ctx context.Context) (nero.Tx, error) {
	return pg.db.BeginTx(ctx, nil)
}

// Create creates a new {{.Type.Name}}
func (pg *PostgresRepository) Create(ctx context.Context, c *Creator) ({{type .Ident.Type.V}}, error) {
	return pg.create(ctx, pg.db, c)
}

// CreateTx creates a new {{.Type.Name}} inside a transaction
func (pg *PostgresRepository) CreateTx(ctx context.Context, tx nero.Tx, c *Creator) ({{type .Ident.Type.V}}, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return {{zero .Ident.Type.V}}, errors.New("expecting tx to be *sql.Tx")
	}

	return pg.create(ctx, txx, c)
}

func (pg *PostgresRepository) create(ctx context.Context, runner nero.SQLRunner, c *Creator) ({{type .Ident.Type.V}}, error) {
	columns := []string{}
	values := []interface{}{}
	{{range $col := .Cols }}
		{{if ne $col.Auto true}}
			if c.{{$col.Identifier}} != {{zero $col.Type.V}} {
				columns = append(columns, "\"{{$col.Name}}\"")
				{{if and ($col.IsArray) (ne $col.IsValueScanner true) -}}
					values = append(values, pq.Array(c.{{$col.Identifier}}))
				{{else -}}
					values = append(values, c.{{$col.Identifier}})
				{{end -}}
			}
		{{end}}
	{{end}}

	qb := squirrel.Insert("\"{{.Collection}}\"").
		Columns(columns...).
		Values(values...).
		Suffix("RETURNING \"{{.Ident.Name}}\"").
		PlaceholderFormat(squirrel.Dollar).
		RunWith(runner)
	if pg.debug {
		sql, args, err := qb.ToSql()
		pg.logger.Printf("method: Create, stmt: %q, args: %v, error: %v", sql, args, err)
	}

	var {{.Ident.Identifier}} {{type .Ident.Type.V}}
	err := qb.QueryRowContext(ctx).Scan(&{{.Ident.Identifier}})
	if err != nil {
		return {{zero .Ident.Type.V}}, err
	}

	return {{.Ident.Identifier}}, nil
}

// CreateMany creates many {{.Type.Name}}
func (pg *PostgresRepository) CreateMany(ctx context.Context, cs ...*Creator) error {
	return pg.createMany(ctx, pg.db, cs...)
}

// CreateManyTx creates many {{.Type.Name}} inside a transaction
func (pg *PostgresRepository) CreateManyTx(ctx context.Context, tx nero.Tx, cs ...*Creator) error {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("expecting tx to be *sql.Tx")
	}

	return pg.createMany(ctx, txx, cs...)
}

func (pg *PostgresRepository) createMany(ctx context.Context, runner nero.SQLRunner, cs ...*Creator) error {
	if len(cs) == 0 {
		return nil
	}

	columns := []string{
		{{range $col := .Cols -}}
			{{if ne $col.Auto true -}}
				"\"{{$col.Name}}\"",
			{{end -}}
		{{end -}}
	}
	qb := squirrel.Insert("\"{{.Collection}}\"").Columns(columns...)
	for _, c := range cs {
		qb = qb.Values(
			{{range $col := .Cols -}}
				{{if ne $col.Auto true -}}
					{{if and ($col.IsArray) (ne $col.IsValueScanner true) -}}
						pq.Array(c.{{$col.Identifier}}),
					{{else -}}
						c.{{$col.Identifier}},
					{{end -}}
				{{end -}}
			{{end -}}
		)
	}

	qb = qb.Suffix("RETURNING \"{{.Ident.Name}}\"").
		PlaceholderFormat(squirrel.Dollar)
	if pg.debug {
		sql, args, err := qb.ToSql()
		pg.logger.Printf("method: CreateMany, stmt: %q, args: %v, error: %v", sql, args, err)
	}

	_, err := qb.RunWith(runner).ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Query queries many {{.Type.Name}}
func (pg *PostgresRepository) Query(ctx context.Context, q *Queryer) ([]*{{type .Type.V}}, error) {
	return pg.query(ctx, pg.db, q)
}

// QueryTx queries many {{.Type.Name}} inside a transaction
func (pg *PostgresRepository) QueryTx(ctx context.Context, tx nero.Tx, q *Queryer) ([]*{{type .Type.V}}, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errors.New("expecting tx to be *sql.Tx")
	}

	return pg.query(ctx, txx, q)
}

func (pg *PostgresRepository) query(ctx context.Context, runner nero.SQLRunner, q *Queryer) ([]*{{type .Type.V}}, error) {
	qb := pg.buildSelect(q)	
	if pg.debug {
		sql, args, err := qb.ToSql()
		pg.logger.Printf("method: Query, stmt: %q, args: %v, error: %v", sql, args, err)
	}

	rows, err := qb.RunWith(runner).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	{{plural (lowerCamel .Type.Name)}} := []*{{type .Type.V}}{}
	for rows.Next() {
		var {{lowerCamel .Type.Name}} {{type .Type.V}}
		err = rows.Scan(
			{{range $col := .Cols -}}
				{{if and ($col.IsArray) (ne $col.IsValueScanner true) -}}
					pq.Array(&{{lowerCamel $.Type.Name}}.{{$col.Field}}),
				{{else -}}
					&{{lowerCamel $.Type.Name}}.{{$col.Field}},
				{{end -}}
			{{end -}}
		)
		if err != nil {
			return nil, err
		}

		{{plural (lowerCamel .Type.Name)}} = append({{plural (lowerCamel .Type.Name)}}, &{{lowerCamel .Type.Name}})
	}

	return {{plural (lowerCamel .Type.Name)}}, nil
}

// QueryOne queries one {{.Type.Name}}
func (pg *PostgresRepository) QueryOne(ctx context.Context, q *Queryer) (*{{type .Type.V}}, error) {
	return pg.queryOne(ctx, pg.db, q)
}

// QueryOneTx queries one {{.Type.Name}} inside a transaction
func (pg *PostgresRepository) QueryOneTx(ctx context.Context, tx nero.Tx, q *Queryer) (*{{type .Type.V}}, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return nil, errors.New("expecting tx to be *sql.Tx")
	}

	return pg.queryOne(ctx, txx, q)
}

func (pg *PostgresRepository) queryOne(ctx context.Context, runner nero.SQLRunner, q *Queryer) (*{{type .Type.V}}, error) {
	qb := pg.buildSelect(q)
	if pg.debug {
		sql, args, err := qb.ToSql()
		pg.logger.Printf("method: QueryOne, stmt: %q, args: %v, error: %v", sql, args, err)
	}

	var {{lowerCamel .Type.Name}} {{type .Type.V}}
	err := qb.RunWith(runner).
		QueryRowContext(ctx).
		Scan(
			{{range $col := .Cols -}}
				{{if and ($col.IsArray) (ne $col.IsValueScanner true) -}}
					pq.Array(&{{lowerCamel $.Type.Name}}.{{$col.Field}}),
				{{else -}}
					&{{lowerCamel $.Type.Name}}.{{$col.Field}},
				{{end -}}	
			{{end -}}
		)
	if err != nil {
		return {{zero .Type.V}}, err
	}

	return &{{lowerCamel .Type.Name}}, nil
}

func (pg *PostgresRepository) buildSelect(q *Queryer) squirrel.SelectBuilder {
	columns := []string{
		{{range $col := .Cols -}}
			"\"{{$col.Name}}\"",
		{{end -}}
	}
	qb := squirrel.Select(columns...).
		From("\"{{.Collection}}\"").
		PlaceholderFormat(squirrel.Dollar)

	pfs := q.pfs
	pb := &comparison.Predicates{}
	for _, pf := range pfs {
		pf(pb)
	}
	` + predsBldrBlock + `

	sfs := q.sfs
	sorts := &sort.Sorts{}
	for _, sf := range sfs {
		sf(sorts)
	}
	for _, s := range sorts.All() {
		col := fmt.Sprintf("%q", s.Col)
		switch s.Direction {
		case sort.Asc:
			qb = qb.OrderBy(col + " ASC")
		case sort.Desc:
			qb = qb.OrderBy(col + " DESC")
		}
	}

	if q.limit > 0 {
		qb = qb.Limit(uint64(q.limit))
	}

	if q.offset > 0 {
		qb = qb.Offset(uint64(q.offset))
	}

	return qb
}

// Update updates {{.Type.Name}}
func (pg *PostgresRepository) Update(ctx context.Context, u *Updater) (int64, error) {
	return pg.update(ctx, pg.db, u)
}

// UpdateTx updates {{.Type.Name}} inside a transaction
func (pg *PostgresRepository) UpdateTx(ctx context.Context, tx nero.Tx, u *Updater) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	return pg.update(ctx, txx, u)
}

func (pg *PostgresRepository) update(ctx context.Context, runner nero.SQLRunner, u *Updater) (int64, error) {
	qb := squirrel.Update("\"{{.Collection}}\"").
		PlaceholderFormat(squirrel.Dollar)	
	{{range $col := .Cols}}
		{{if ne $col.Auto true}}
			if u.{{$col.Identifier}} != {{zero $col.Type.V}} {
				{{if and ($col.IsArray) (ne $col.IsValueScanner true) -}}
					qb = qb.Set("\"{{$col.Name}}\"", pq.Array(u.{{$col.Identifier}}))
				{{else -}}
					qb = qb.Set("\"{{$col.Name}}\"", u.{{$col.Identifier}})
				{{end -}}
			}
		{{end}}
	{{end}}

	pfs := u.pfs
	pb := &comparison.Predicates{}
	for _, pf := range pfs {
		pf(pb)
	}
	` + predsBldrBlock + `

	if pg.debug {
		sql, args, err := qb.ToSql()
		pg.logger.Printf("method: Update, stmt: %q, args: %v, error: %v", sql, args, err)
	}

	res, err := qb.RunWith(runner).ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// Delete deletes {{.Type.Name}}
func (pg *PostgresRepository) Delete(ctx context.Context, d *Deleter) (int64, error) {
	return pg.delete(ctx, pg.db, d)
}

// Delete deletes {{.Type.Name}} inside a transaction
func (pg *PostgresRepository) DeleteTx(ctx context.Context, tx nero.Tx, d *Deleter) (int64, error) {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return 0, errors.New("expecting tx to be *sql.Tx")
	}

	return pg.delete(ctx, txx, d)
}

func (pg *PostgresRepository) delete(ctx context.Context, runner nero.SQLRunner, d *Deleter) (int64, error) {
	qb := squirrel.Delete("\"{{.Collection}}\"").
		PlaceholderFormat(squirrel.Dollar)

	pfs := d.pfs
	pb := &comparison.Predicates{}
	for _, pf := range pfs {
		pf(pb)
	}
	` + predsBldrBlock + `

	if pg.debug {
		sql, args, err := qb.ToSql()
		pg.logger.Printf("method: Delete, stmt: %q, args: %v, error: %v", sql, args, err)
	}

	res, err := qb.RunWith(runner).ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// Aggregate runs aggregate operations
func (pg *PostgresRepository) Aggregate(ctx context.Context, a *Aggregator) error {
	return pg.aggregate(ctx, pg.db, a)
}

// Aggregate runs aggregate operations inside a transaction
func (pg *PostgresRepository) AggregateTx(ctx context.Context, tx nero.Tx, a *Aggregator) error {
	txx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("expecting tx to be *sql.Tx")
	}

	return pg.aggregate(ctx, txx, a)
}

func (pg *PostgresRepository) aggregate(ctx context.Context, runner nero.SQLRunner, a *Aggregator) error {
	aggs := &aggregate.Aggregates{}
	for _, aggf := range a.aggfs {
		aggf(aggs)
	}
	cols := []string{}
	for _, agg := range aggs.All() {
		col := agg.Col
		qcol := fmt.Sprintf("%q", col)
		switch agg.Fn {
		case aggregate.Avg:
			cols = append(cols, "AVG("+qcol+") avg_"+col)
		case aggregate.Count:
			cols = append(cols, "COUNT("+qcol+") count_"+col)
		case aggregate.Max:
			cols = append(cols, "MAX("+qcol+") max_"+col)
		case aggregate.Min:
			cols = append(cols, "MIN("+qcol+") min_"+col)
		case aggregate.Sum:
			cols = append(cols, "SUM("+qcol+") sum_"+col)
		case aggregate.None:
			cols = append(cols, qcol)
		}
	}

	qb := squirrel.Select(cols...).From("\"{{.Collection}}\"").
		PlaceholderFormat(squirrel.Dollar)

	groups := []string{}
	for _, group := range a.groups {
		groups = append(groups, fmt.Sprintf("%q", group.String()))
	}
	qb = qb.GroupBy(groups...)

	pfs := a.pfs
	pb := &comparison.Predicates{}
	for _, pf := range pfs {
		pf(pb)
	}
	` + predsBldrBlock + `

	sfs := a.sfs
	sorts := &sort.Sorts{}
	for _, sf := range sfs {
		sf(sorts)
	}
	for _, s := range sorts.All() {
		col := fmt.Sprintf("%q", s.Col)
		switch s.Direction {
		case sort.Asc:
			qb = qb.OrderBy(col + " ASC")
		case sort.Desc:
			qb = qb.OrderBy(col + " DESC")
		}
	}

	if pg.debug {
		sql, args, err := qb.ToSql()
		pg.logger.Printf("method: Aggregate, stmt: %q, args: %v, error: %v", sql, args, err)
	}

	rows, err := qb.RunWith(runner).QueryContext(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()

	v := reflect.ValueOf(a.v).Elem()
	t := reflect.TypeOf(v.Interface()).Elem()
	if t.NumField() != len(cols) {
		return errors.New("aggregate columns and destination struct field count should match")
	}

	for rows.Next() {
		ve := reflect.New(t).Elem()
		dest := make([]interface{}, ve.NumField())
		for i := 0; i < ve.NumField(); i++ {
			dest[i] = ve.Field(i).Addr().Interface()
		}

		err = rows.Scan(dest...)
		if err != nil {
			return err
		}

		v.Set(reflect.Append(v, ve))
	}

	return nil
}
`

const predsBldrBlock = `
	for _, p := range pb.All() {
		switch p.Op {
		case comparison.Eq:
			col, ok := p.Arg.(Column)
			if ok {
				qb = qb.Where(fmt.Sprintf("%q = %q", p.Col, col.String()))
			} else {
				qb = qb.Where(fmt.Sprintf("%q = ?", p.Col), p.Arg)
			}
		case comparison.NotEq:
			col, ok := p.Arg.(Column)
			if ok {
				qb = qb.Where(fmt.Sprintf("%q <> %q", p.Col, col.String()))
			} else {	
				qb = qb.Where(fmt.Sprintf("%q <> ?", p.Col), p.Arg)
			}
		case comparison.Gt:
			col, ok := p.Arg.(Column)
			if ok {
				qb = qb.Where(fmt.Sprintf("%q > %q", p.Col, col.String()))
			} else {
				qb = qb.Where(fmt.Sprintf("%q > ?", p.Col), p.Arg)
			}
		case comparison.GtOrEq:
			col, ok := p.Arg.(Column)
			if ok {
				qb = qb.Where(fmt.Sprintf("%q >= %q", p.Col, col.String()))
			} else {
				qb = qb.Where(fmt.Sprintf("%q >= ?", p.Col), p.Arg)
			}
		case comparison.Lt:
			col, ok := p.Arg.(Column)
			if ok {
				qb = qb.Where(fmt.Sprintf("%q < %q", p.Col, col.String()))
			} else {
				qb = qb.Where(fmt.Sprintf("%q < ?", p.Col), p.Arg)
			}
		case comparison.LtOrEq:
			col, ok := p.Arg.(Column)
			if ok {
				qb = qb.Where(fmt.Sprintf("%q <= %q", p.Col, col.String()))
			} else {
				qb = qb.Where(fmt.Sprintf("%q <= ?", p.Col), p.Arg)
			}
		case comparison.IsNull:
			qb = qb.Where(fmt.Sprintf("%q IS NULL", p.Col))
		case comparison.IsNotNull:
			qb = qb.Where(fmt.Sprintf("%q IS NOT NULL", p.Col))
		case comparison.In, comparison.NotIn:
			args := p.Arg.([]interface{})
			if len(args) == 0 {
				continue
			}
			qms := []string{}
			for range args {
				qms = append(qms, "?")
			}
			fmtStr := "%q IN (%s)"
			if p.Op == comparison.NotIn {
				fmtStr = "%q NOT IN (%s)"
			}
			plchldr := strings.Join(qms, ",")
			qb = qb.Where(fmt.Sprintf(fmtStr, p.Col, plchldr), args...)
		}
	}
`

package gen

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/jinzhu/inflection"

	gen "github.com/sf9v/nero/gen/internal"
)

func newRepositoryFile(schema *gen.Schema) (*bytes.Buffer, error) {
	tmpl, err := template.New("repository.tmpl").Funcs(template.FuncMap{
		"plural": inflection.Plural,
		"type": func(v interface{}) string {
			return fmt.Sprintf("%T", v)
		},
	}).Parse(repositoryTmpl)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, schema)
	return buf, err
}

const repositoryTmpl = `
// Code generated by nero, DO NOT EDIT.
package {{.Pkg}}

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sf9v/nero"
	{{range $import := .SchemaImports -}}
		"{{$import}}"
	{{end -}}
	{{range $import := .ColumnImports -}}
		"{{$import}}"
	{{end -}}
)

// Repository is a repository for {{.Type.Name}}
type Repository interface {
	// Tx begins a new transaction
	Tx(context.Context) (nero.Tx, error)
	// Create creates a new {{.Type.Name}}
	Create(context.Context, *Creator) (id {{type .Ident.Type.V}}, err error)
	// CreateTx creates a new type .Type.Name}} inside a transaction
	CreateTx(context.Context, nero.Tx, *Creator) (id {{type .Ident.Type.V}}, err error)
	// CreateMany creates many {{.Type.Name}}
	CreateMany(context.Context, ...*Creator) error
	// CreateManyTx creates many {{.Type.Name}} inside a transaction
	CreateManyTx(context.Context, nero.Tx, ...*Creator) error
	// Query queries many {{.Type.Name}}
	Query(context.Context, *Queryer) ([]{{type .Type.V}}, error)
	// QueryTx queries many {{.Type.V}} inside a transaction
	QueryTx(context.Context, nero.Tx, *Queryer) ([]{{type .Type.V}}, error)
	// QueryOne queries one {{.Type.Name}}
	QueryOne(context.Context, *Queryer) ({{type .Type.V}}, error)
	// QueryOneTx queries one {{.Type.Name}} inside a transaction
	QueryOneTx(context.Context, nero.Tx, *Queryer) ({{type .Type.V}}, error)
	// Update updates {{.Type.Name}}
	Update(context.Context, *Updater) (rowsAffected int64, err error)
	// UpdateTx updates {{.Type.Name}} inside a transaction
	UpdateTx(context.Context, nero.Tx, *Updater) (rowsAffected int64, err error)
	// Delete deletes {{.Type.Name}}
	Delete(context.Context, *Deleter) (rowsAffected int64, err error)
	// Delete deletes {{.Type.Name}} inside a transaction
	DeleteTx(context.Context, nero.Tx, *Deleter) (rowsAffected int64, err error)
	// Aggregate performs aggregate query
	Aggregate(context.Context, *Aggregator) error
	// Aggregate performs aggregate query inside a transaction
	AggregateTx(context.Context, nero.Tx, *Aggregator) error
}

// Creator is a create builder for {{.Type.Name}}
type Creator struct {
	{{range $col := .Cols -}}
		{{if ne $col.Auto true -}}
		{{$col.Identifier}} {{type $col.Type.V}}
		{{end -}}
	{{end -}}
}

// NewCreator is a factory for Creator
func NewCreator() *Creator {
	return &Creator{}
}

{{range $col := .Cols}}
	{{if ne $col.Auto true -}}
		// {{$col.Field}} is a setter for {{$col.Identifier}}
		func (c *Creator) {{$col.Field}}({{$col.Identifier}} {{type $col.Type.V}}) *Creator {
			c.{{$col.Identifier}} = {{$col.Identifier}}
			return c
		}
	{{end -}}
{{end -}}

// Queryer is a query builder for {{.Type.Name}}
type Queryer struct {
	limit  uint
	offset uint
	pfs    []PredFunc
	sfs    []SortFunc
}

// NewQueryer is a factory for Queryer
func NewQueryer() *Queryer {
	return &Queryer{}
}

// Where adds predicates to the query
func (q *Queryer) Where(pfs ...PredFunc) *Queryer {
	q.pfs = append(q.pfs, pfs...)
	return q
}

// Sort adds sorting expressions to the query
func (q *Queryer) Sort(sfs ...SortFunc) *Queryer {
	q.sfs = append(q.sfs, sfs...)
	return q
}

// Limit adds limit clause to the query
func (q *Queryer) Limit(limit uint) *Queryer {
	q.limit = limit
	return q
}

// Offset adds offset clause to the query
func (q *Queryer) Offset(offset uint) *Queryer {
	q.offset = offset
	return q
}

// Updater is an update builder for {{.Type.Name}}
type Updater struct {
	{{range $col := .Cols -}}
		{{if ne $col.Auto true -}}
		{{$col.Identifier}} {{type $col.Type.V}}
		{{end -}}
	{{end -}}
	pfs []PredFunc
}

// NewUpdater is a factory for Updater
func NewUpdater() *Updater {
	return &Updater{}
}

{{range $col := .Cols}}
	{{if ne $col.Auto true -}}
		// {{$col.Field}} is a setter for {{$col.Identifier}}
		func (c *Updater) {{$col.Field}}({{$col.Identifier}} {{type $col.Type.V}}) *Updater {
			c.{{$col.Identifier}} = {{$col.Identifier}}
			return c
		}
	{{end -}}
{{end -}}

// Where adds predicates to the update builder
func (u *Updater) Where(pfs ...PredFunc) *Updater {
	u.pfs = append(u.pfs, pfs...)
	return u
}

// Deleter is a delete builder for {{.Type.Name}}
type Deleter struct {
	pfs []PredFunc
}

// NewDeleter is a factory for Deleter
func NewDeleter() *Deleter {
	return &Deleter{}
}

// Where adds predicates to the delete builder
func (d *Deleter) Where(pfs ...PredFunc) *Deleter {
	d.pfs = append(d.pfs, pfs...)
	return d
}

// Aggregator is an aggregate builder for {{.Type.Name}}
type Aggregator struct {
	v      interface{}
	aggfs  []AggFunc
	pfs    []PredFunc
	sfs    []SortFunc
	groups []Column
}

// NewAggregator is a factory for Aggregator
// 'v' argument must be an array of struct
func NewAggregator(v interface{}) *Aggregator {
	return &Aggregator{
		v: v,
	}
}

// Aggregate adds aggregate functions to the aggregate builder
func (a *Aggregator) Aggregate(aggfs ...AggFunc) *Aggregator {
	a.aggfs = append(a.aggfs, aggfs...)
	return a
}

// Where adds predicates to the aggregate builder
func (a *Aggregator) Where(pfs ...PredFunc) *Aggregator {
	a.pfs = append(a.pfs, pfs...)
	return a
}

// Sort adds sorting expressions to the aggregate builder
func (a *Aggregator) Sort(sfs ...SortFunc) *Aggregator {
	a.sfs = append(a.sfs, sfs...)
	return a
}

// Group adds grouping clause to the aggregate builder
func (a *Aggregator) Group(cols ...Column) *Aggregator {
	a.groups = append(a.groups, cols...)
	return a
}

// rollback performs a rollback
func rollback(tx nero.Tx, err error) error {
	rerr := tx.Rollback()
	if rerr != nil {
		err = errors.Wrapf(err, "rollback error: %v", rerr)
	}
	return err
}
`

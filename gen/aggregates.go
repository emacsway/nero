package gen

import (
	"bytes"
	"text/template"

	"github.com/sf9v/nero/aggregate"
	gen "github.com/sf9v/nero/gen/internal"
)

func newAggregatesFile(schema *gen.Schema) (*bytes.Buffer, error) {
	v := struct {
		Functions []aggregate.Function
		Schema    *gen.Schema
	}{
		Functions: []aggregate.Function{
			aggregate.Avg, aggregate.Count,
			aggregate.Max, aggregate.Min,
			aggregate.Sum, aggregate.None,
		},
		Schema: schema,
	}

	tmpl, err := template.New("aggregates.tmpl").
		Parse(aggregatesTmpl)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, v)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

const aggregatesTmpl = `
// Code generated by nero, DO NOT EDIT.
package {{.Schema.Pkg}}

import (
	"github.com/sf9v/nero/aggregate"
)

// AggFunc is an aggregate function
type AggFunc func(*aggregate.Aggregates)

{{range $fn := .Functions}}
// {{$fn.String}} is a {{$fn.Desc}} aggregate function
func {{$fn.String}}(col Column) AggFunc {
	return func(a *aggregate.Aggregates) {
		a.Add(&aggregate.Aggregate{
			Col: col.String(),
			Fn: aggregate.{{$fn.String}},
		})
	}
}
{{end}}
`

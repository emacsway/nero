package gen

import (
	"github.com/dave/jennifer/jen"
	gen "github.com/sf9v/nero/gen/internal"
)

func newMeta(schema *gen.Schema) *jen.Statement {
	stmnt := jen.Const().Defs(
		jen.Id("collection").Op("=").Lit(schema.Coln),
	).Line()

	stmnt = stmnt.Type().Id("Column").Int().Line()
	// column stringer
	stmnt = stmnt.Func().Params(jen.Id("c").Id("Column")).
		Id("String").Params().Params(jen.String()).Block(
		jen.Switch(jen.Id("c")).BlockFunc(func(g *jen.Group) {
			for _, col := range schema.Cols {
				colTypeName := col.CamelName()
				if len(col.StructField) > 0 {
					colTypeName = col.StructField
				}
				colTypeName = camel("Column" + "_" + colTypeName)
				g.Case(jen.Id(colTypeName)).Block(
					jen.Return(jen.Lit(col.Name)),
				)
			}
		}),
		jen.Return(jen.Lit("invalid"))).Line()

	// column type names
	stmnt = stmnt.Const().DefsFunc(func(g *jen.Group) {
		for i, col := range schema.Cols {
			colTypeName := col.CamelName()
			if len(col.StructField) > 0 {
				colTypeName = col.StructField
			}
			colTypeName = camel("Column" + "_" + colTypeName)
			if i == 0 {
				g.Id(colTypeName).Id("Column").Op("=").Iota()
				continue
			}

			g.Id(colTypeName)
		}
	})

	return stmnt
}
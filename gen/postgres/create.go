package postgres

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
	gen "github.com/sf9v/nero/gen/internal"
	"github.com/sf9v/nero/jenx"
)

func newCreateBlock(schema *gen.Schema) *jen.Statement {
	ident := schema.Ident
	identv := ident.Type.V()
	return jen.Func().Params(rcvrParamC).Id("Create").
		Params(
			jen.Id("ctx").Add(ctxC),
			jen.Id("c").Op("*").Id("Creator"),
		).
		Params(jenx.Type(identv), jen.Error()).
		BlockFunc(func(g *jen.Group) {
			g.List(jen.Id("tx"), jen.Err()).Op(":=").
				Add(rcvrIDC).Dot("Tx").Call(ctxIDC)
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jenx.Zero(identv), jen.Err())).
				Line()

			g.List(jen.Id(ident.LowerCamelName()), jen.Err()).Op(":=").
				Add(rcvrIDC).Dot("CreateTx").Call(
				ctxIDC,
				jen.Id("tx"),
				jen.Id("c"),
			)
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jenx.Zero(identv), txRollbackC),
			).Line()

			g.Return(
				jen.Id(ident.LowerCamelName()),
				txCommitC,
			)
		})
}

func newCreateManyBlock() *jen.Statement {
	return jen.Func().Params(rcvrParamC).Id("CreateMany").
		Params(
			jen.Id("ctx").Add(ctxC),
			jen.Id("cs").Op("...").Op("*").Id("Creator"),
		).
		Params(jen.Error()).
		BlockFunc(func(g *jen.Group) {
			g.List(jen.Id("tx"), jen.Err()).Op(":=").
				Add(rcvrIDC).Dot("Tx").Call(ctxIDC)
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Err())).Line()
			g.List(jen.Err()).Op("=").Add(rcvrIDC).Dot("CreateManyTx").
				Call(ctxIDC, jen.Id("tx"), jen.Id("cs").Op("..."))
			g.If(jen.Err().Op("!=").Nil()).
				Block(jen.Return(txRollbackC)).Line()
			g.Return(txCommitC)
		})
}

func newCreateTxBlock(schema *gen.Schema) *jen.Statement {
	ident := schema.Ident
	identv := ident.Type.V()
	return jen.Func().Params(rcvrParamC).Id("CreateTx").
		Params(
			jen.Id("ctx").Add(ctxC),
			jen.Id("tx").Add(txC),
			jen.Id("c").Op("*").Id("Creator"),
		).
		Params(jenx.Type(identv), jen.Error()).
		BlockFunc(func(g *jen.Group) {
			// assert tx type
			g.List(jen.Id("txx"), jen.Id("ok")).Op(":=").
				Id("tx").Assert(jen.Op("*").Qual("database/sql", "Tx"))
			g.If(jen.Op("!").Id("ok")).Block(
				jen.Return(
					jenx.Zero(identv),
					jen.Qual(errPkg, "New").Call(jen.Lit("expecting tx to be *sql.Tx")),
				)).Line()

			// quote column names
			g.Id("columns").Op(":=").Index().String().Values()
			g.Id("values").Op(":=").Index().Interface().Values()
			for _, col := range schema.Cols {
				if col.Auto {
					continue
				}
				field := col.LowerCamelName()
				if len(col.StructField) > 0 {
					field = strcase.ToLowerCamel(col.StructField)
				}
				colv := col.Type.V()
				g.If(jen.Id("c").Dot(field).
					Op("!=").Add(jenx.Zero(colv))).
					Block(
						jen.Id("columns").Op("=").Append(
							jen.Id("columns"),
							jen.Lit(fmt.Sprintf("%q", col.Name)),
						),
						jen.Id("values").Op("=").Append(
							jen.Id("values"),
							jen.Id("c").Dot(field),
						),
					)
			}

			g.Line()

			// query builder
			g.Id("qb").Op(":=").Qual(sqPkg, "Insert").
				Call(jen.Lit(fmt.Sprintf("%q", schema.Collection))).
				Add(jenx.Dotln("Columns")).
				Call(jen.Id("columns").Op("...")).
				Add(jenx.Dotln("Values")).
				Call(jen.Id("values").Op("...")).
				Add(jenx.Dotln("Suffix")).
				Call(jen.Lit(fmt.Sprintf("RETURNING %q", ident.Name))).
				Add(jenx.Dotln("PlaceholderFormat")).
				Call(jen.Qual(sqPkg, "Dollar")).
				Add(jenx.Dotln("RunWith")).
				Call(jen.Id("txx"))
			// debug
			g.Add(newDebugLogBlock("Create")).Line().Line()

			g.Var().Id(ident.LowerCamelName()).Add(jenx.Type(identv))
			g.Err().Op(":=").Id("qb").Dot("QueryRowContext").
				Call(ctxIDC).Dot("Scan").Call(jen.Op("&").Id(ident.LowerCamelName()))
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jenx.Zero(identv), jen.Err()),
			).Line()

			g.Return(jen.Id(ident.LowerCamelName()), jen.Nil())
		})
}

func newCreateManyTxBlock(schema *gen.Schema) *jen.Statement {
	ident := schema.Ident
	return jen.Func().Params(rcvrParamC).Id("CreateManyTx").
		Params(
			jen.Id("ctx").Add(ctxC),
			jen.Id("tx").Add(txC),
			jen.Id("cs").Op("...").Op("*").Id("Creator"),
		).
		Params(jen.Error()).
		BlockFunc(func(g *jen.Group) {
			g.If(jen.Len(jen.Id("cs")).Op("==").Lit(0)).Block(
				jen.Return(jen.Nil()),
			).Line()

			// assert tx type
			g.List(jen.Id("txx"), jen.Id("ok")).Op(":=").
				Id("tx").Assert(jen.Op("*").Qual("database/sql", "Tx"))
			g.If(jen.Op("!").Id("ok")).Block(jen.Return(
				jen.Qual(errPkg, "New").Call(jen.Lit("expecting tx to be *sql.Tx")),
			)).Line()

			g.Id("columns").Op(":=").Index().String().
				ValuesFunc(func(g *jen.Group) {
					for _, col := range schema.Cols {
						if col.Auto {
							continue
						}
						g.Lit(fmt.Sprintf("%q", col.Name))
					}
				})

			// query builder
			g.Id("qb").Op(":=").Qual(sqPkg, "Insert").
				Call(jen.Lit(fmt.Sprintf("%q", schema.Collection))).
				Dot("Columns").Call(jen.Id("columns").Op("..."))

			g.For(jen.List(jen.Id("_"), jen.Id("c")).
				Op(":=").Range().Id("cs"),
			).BlockFunc(func(g *jen.Group) {
				g.Id("qb").Op("=").Id("qb").Dot("Values").
					CallFunc(func(g *jen.Group) {
						for _, col := range schema.Cols {
							if col.Auto {
								continue
							}
							field := col.LowerCamelName()
							if len(col.StructField) > 0 {
								field = strcase.ToLowerCamel(col.StructField)
							}
							g.Id("c").Dot(field)
						}
					})
			}).Line()

			g.Id("qb").Op("=").Id("qb").Dot("Suffix").
				Call(jen.Lit(fmt.Sprintf("RETURNING %q", ident.Name))).
				Op(".").Line().Id("PlaceholderFormat").
				Call(jen.Qual(sqPkg, "Dollar"))

			// debug
			g.Add(newDebugLogBlock("CreateMany")).Line().Line()

			g.List(jen.Id("_"), jen.Err()).Op(":=").Id("qb").
				Dot("RunWith").Call(jen.Id("txx")).
				Dot("ExecContext").Call(ctxIDC)
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Err())).Line()

			g.Return(jen.Nil())
		})
}

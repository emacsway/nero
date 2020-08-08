package postgres

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
	gen "github.com/sf9v/nero/gen/internal"
	"github.com/sf9v/nero/jenx"
)

func newUpdateBlock() *jen.Statement {
	return jen.Func().Params(rcvrParamC).Id("Update").
		Params(
			jen.Id("ctx").Add(ctxC),
			jen.Id("u").Op("*").Id("Updater"),
		).
		Params(jen.Int64(), jen.Error()).
		BlockFunc(func(g *jen.Group) {
			g.List(jen.Id("tx"), jen.Err()).Op(":=").
				Add(rcvrIDC).Dot("Tx").Call(ctxIDC)
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Lit(0), jen.Err())).Line()

			g.List(jen.Id("rowsAffected"), jen.Err()).Op(":=").
				Add(rcvrIDC).Dot("UpdateTx").Call(
				ctxIDC, jen.Id("tx"), jen.Id("u"))
			g.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Lit(0), txRollbackC),
			).Line()

			g.Return(jen.Id("rowsAffected"), txCommitC)
		})
}

func newUpdateTxBlock(schema *gen.Schema) *jen.Statement {
	return jen.Func().Params(rcvrParamC).Id("UpdateTx").
		Params(
			jen.Id("ctx").Add(ctxC),
			jen.Id("tx").Add(txC),
			jen.Id("u").Op("*").Id("Updater"),
		).
		Params(jen.Int64(), jen.Error()).
		BlockFunc(func(g *jen.Group) {
			// assert tx type
			g.List(jen.Id("txx"), jen.Id("ok")).Op(":=").
				Id("tx").Assert(jen.Op("*").Qual("database/sql", "Tx"))
			g.If(jen.Op("!").Id("ok")).Block(
				jen.Return(
					jen.Lit(0),
					jen.Qual(errPkg, "New").
						Call(jen.Lit("expecting tx to be *sql.Tx")),
				)).Line()

			ifErr := jen.If(jen.Err().Op("!=").Nil()).Block(
				jen.Return(jen.Lit(0), jen.Err()))

			// query builder
			g.Id("qb").Op(":=").Qual(sqPkg, "Update").
				Call(jen.Lit(fmt.Sprintf("%q", schema.Collection))).
				Dot("PlaceholderFormat").Call(jen.Qual(sqPkg, "Dollar"))

			for _, col := range schema.Cols {
				if col.Auto {
					continue
				}

				field := col.LowerCamelName()
				if len(col.StructField) > 0 {
					field = strcase.ToLowerCamel(col.StructField)
				}

				colv := col.Type.V()
				g.If(jen.Id("u").Dot(field).
					Op("!=").Add(jenx.Zero(colv))).
					Block(jen.Id("qb").Op("=").Id("qb").Dot("Set").
						Call(
							jen.Lit(fmt.Sprintf("%q", col.Name)),
							jen.Id("u").Dot(field),
						))
			}

			g.Line()

			g.Id("pfs").Op(":=").Id("u").Dot("pfs")
			g.Add(newPredicatesBlock()).Line()

			// debug
			g.Add(newDebugLogBlock("Update")).Line().Line()

			g.List(jen.Id("res"), jen.Err()).Op(":=").Id("qb").
				Dot("RunWith").Call(jen.Id("txx")).
				Dot("ExecContext").Call(ctxIDC)
			g.Add(ifErr).Line()

			g.List(jen.Id("rowsAffected"), jen.Err()).Op(":=").
				Id("res").Dot("RowsAffected").Call()
			g.Add(ifErr).Line()

			g.Return(jen.Id("rowsAffected"), jen.Nil())
		})
}

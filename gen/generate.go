package gen

import (
	"bytes"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"

	"github.com/sf9v/nero"
	gen "github.com/sf9v/nero/gen/internal"
)

const (
	pkgPath = "github.com/sf9v/nero"
	header  = "Code generated by nero, DO NOT EDIT."
)

type file struct {
	name string
	jf   *jen.File
}

// OutFile is an output file
type OutFile struct {
	Name string
	*bytes.Buffer
}

// Generate generates the repository and returns the output files
func Generate(s nero.Schemaer) ([]*OutFile, error) {
	schema, err := buildSchema(s)
	if err != nil {
		return nil, err
	}

	pkgName := strings.ToLower(schema.Pkg)
	pkgFile := jen.NewFile(pkgName)
	pkgFile.Const().Defs(
		jen.Id("collection").Op("=").Lit(schema.Coln),
	)

	files := []*file{{name: "meta.go", jf: pkgFile}}

	predsFile := jen.NewFile(pkgName)
	predsFile.Add(newPredicates(schema))
	files = append(files, &file{
		name: "predicates.go",
		jf:   predsFile,
	})

	repoFile := jen.NewFile(pkgName)
	repoFile.Add(newRepository(schema))
	files = append(files, &file{
		name: "repository.go",
		jf:   repoFile,
	})

	creatorFile := jen.NewFile(pkgName)
	creatorFile.Add(newCreator(schema))
	files = append(files, &file{
		name: "creator.go",
		jf:   creatorFile,
	})

	queryerFile := jen.NewFile(pkgName)
	queryerFile.Add(newQueryer(schema))
	files = append(files, &file{
		name: "queryer.go",
		jf:   queryerFile,
	})

	updaterFile := jen.NewFile(pkgName)
	updaterFile.Add(newUpdater(schema))
	files = append(files, &file{
		name: "updater.go",
		jf:   updaterFile,
	})

	deleterFile := jen.NewFile(pkgName)
	deleterFile.Add(newDeleter())
	files = append(files, &file{
		name: "deleter.go",
		jf:   deleterFile,
	})

	// sqlite repository implementation
	sqliteRepoFile := jen.NewFile(pkgName)
	sqliteRepoFile.Anon("github.com/mattn/go-sqlite3")
	sqliteRepoFile.Add(newSQLiteRepo(schema))
	files = append(files, &file{
		name: "sqlite_repository.go",
		jf:   sqliteRepoFile,
	})

	outFiles := []*OutFile{}

	for _, file := range files {
		buff := &bytes.Buffer{}
		file.jf.PackageComment(header)
		err = file.jf.Render(buff)
		if err != nil {
			return nil, errors.Wrap(err, "render file")
		}

		outFiles = append(outFiles, &OutFile{
			Name:   file.name,
			Buffer: buff,
		})
	}

	return outFiles, nil
}

func buildSchema(s nero.Schemaer) (*gen.Schema, error) {
	ns := s.Schema()
	schema := &gen.Schema{
		Coln: ns.Collection,
		Typ:  gen.NewTyp(s),
		Cols: []*gen.Col{},
		Pkg:  ns.Pkg,
	}

	identCnt := 0
	for _, co := range ns.Columns {
		col := &gen.Col{
			Name:      co.Name,
			FieldName: toCamel(co.Name),
			Typ:       gen.NewTyp(co.T),
			Ident:     co.IsIdent,
			Auto:      co.IsAuto,
		}

		if len(co.FieldName) > 0 {
			col.FieldName = co.FieldName
		}

		if co.IsIdent {
			schema.Ident = col
			identCnt++
		}

		schema.Cols = append(schema.Cols, col)
	}

	if identCnt == 0 {
		return nil, errors.New("at least one ident column is required")
	}

	if identCnt > 1 {
		return nil, errors.New("only one ident column is allowed")
	}

	return schema, nil
}

func toCamel(s string) string {
	return strcase.ToCamel(s)
}
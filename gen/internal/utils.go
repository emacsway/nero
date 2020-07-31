package internal

import (
	"reflect"

	"github.com/dave/jennifer/jen"
)

func GetTypeC(typ *Typ) jen.Code {
	c := &jen.Statement{}
	if typ.Nillabe {
		c = c.Op("*")
	}

	// built-in types
	if typ.PkgPath == "" {
		switch typ.T.Kind() {
		case reflect.Int:
			return c.Int()
		case reflect.Int32:
			return c.Int32()
		case reflect.Int64:
			return c.Int64()
		case reflect.Uint:
			return c.Uint()
		case reflect.Uint32:
			return c.Uint32()
		case reflect.Uint64:
			return c.Uint64()
		case reflect.Float32:
			return c.Float32()
		case reflect.Float64:
			return c.Float64()
		case reflect.String:
			return c.String()
		}
	}

	return c.Qual(typ.PkgPath, typ.Name)
}

func GetZeroValC(typ *Typ) jen.Code {
	if typ.Nillabe {
		return jen.Nil()
	}

	// built-in types
	if typ.PkgPath == "" {
		switch typ.T.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return jen.Lit(0)
		case reflect.String:
			return jen.Lit("")
		}
	}

	return jen.Qual(typ.PkgPath, typ.Name).Op("{").Op("}")
}
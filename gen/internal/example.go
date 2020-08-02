package internal

import (
	"time"

	"github.com/sf9v/nero"
)

// Example is an example type
type Example struct {
	ID        int64
	Name      string
	UpdatedAt *time.Time
	CreatedAt *time.Time
}

// Schema implements nero.Schemaer
func (e *Example) Schema() *nero.Schema {
	return &nero.Schema{
		Pkg:        "example",
		Collection: "examples",
		Columns: []*nero.Column{
			nero.NewColumn("id", e.ID).Field("ID").Ident().Auto(),
			nero.NewColumn("name", e.Name),
			nero.NewColumn("updated_at", e.UpdatedAt),
			nero.NewColumn("created_at", e.CreatedAt).Auto(),
		},
	}
}

// Example2 is an example type that has a string id
type Example2 struct {
	ID        string
	Name      string
	UpdatedAt *time.Time
	CreatedAt *time.Time
}

// Schema implements nero.Schemaer
func (e *Example2) Schema() *nero.Schema {
	return &nero.Schema{
		Pkg:        "example2",
		Collection: "examples",
		Columns: []*nero.Column{
			nero.NewColumn("id", e.ID).Field("ID").Ident().Auto(),
			nero.NewColumn("name", e.Name),
			nero.NewColumn("updated_at", e.UpdatedAt),
			nero.NewColumn("created_at", e.CreatedAt).Auto(),
		},
	}
}

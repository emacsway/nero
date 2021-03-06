package customtypes

import (
	"github.com/segmentio/ksuid"
	"github.com/sf9v/nero"
)

// Custom demonstrates the use of many different field types
type Custom struct {
	ID             int64
	S              string
	MapStrStr      map[string]string
	MapStrPtrStr   map[string]*string
	MapInt64Str    map[int64]string
	MapInt64PtrStr map[int64]*string
	MapStrItem     map[string]Item
	MapStrPtrItem  map[string]*Item
	Item           Item
	PtrItem        *Item
	Items          []Item
	PtrItems       []*Item
	KSID           ksuid.KSUID
	NullColumn     *string
}

// Item is an example struct embedded in Custom struct
type Item struct {
	Name string
}

// Schema implements nero.Schemaer
func (c *Custom) Schema() *nero.Schema {
	return &nero.Schema{
		Pkg:        "user",
		Collection: "users",
		Columns: []*nero.Column{
			nero.NewColumn("id", c.ID).
				StructField("ID").Ident().Auto(),
			nero.NewColumn("ksid", c.KSID).StructField("KSID"),
			nero.NewColumn("s", c.S),
			nero.NewColumn("map_str_str", c.MapStrStr),
			nero.NewColumn("map_str_ptr_str", c.MapStrPtrStr),
			nero.NewColumn("map_int64_str", c.MapInt64Str),
			nero.NewColumn("map_int64_ptr_str", c.MapInt64PtrStr),
			nero.NewColumn("map_str_item", c.MapStrItem),
			nero.NewColumn("map_str_ptr_item", c.MapStrPtrItem),
			nero.NewColumn("item", c.Item),
			nero.NewColumn("ptr_item", c.PtrItem),
			nero.NewColumn("items", c.Items),
			nero.NewColumn("ptr_items", c.PtrItems),
			nero.NewColumn("null_column", c.NullColumn).Nullable(),
		},
	}
}

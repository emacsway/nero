package nero

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestColumn(t *testing.T) {
	cfg := NewColumn("id", int64(0)).
		Auto().Ident().StructField("ID").
		Cfg()
	assert.Equal(t, "id", cfg.Name)
	_, ok := cfg.T.(int64)
	assert.True(t, ok)
	assert.True(t, cfg.Ident)
	assert.True(t, cfg.Auto)
	assert.Equal(t, "ID", cfg.StructField)

	now := time.Now()
	cfg = NewColumn("updated_at", &now).Nullable().Cfg()
	assert.True(t, cfg.Nullable)

	cfg = NewColumn("comparable", "").ColumnComparable().Cfg()
	assert.True(t, cfg.ColumnComparable)
}

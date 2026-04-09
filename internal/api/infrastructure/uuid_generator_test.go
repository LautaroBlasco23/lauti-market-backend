package infrastructure

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestGenerate_ReturnsNonEmpty(t *testing.T) {
	g := NewUUIDGenerator()
	result := g.Generate()
	assert.NotEmpty(t, result)
}

func TestGenerate_Unique(t *testing.T) {
	g := NewUUIDGenerator()
	a := g.Generate()
	b := g.Generate()
	assert.NotEqual(t, a, b)
}

func TestGenerate_ValidUUIDFormat(t *testing.T) {
	g := NewUUIDGenerator()
	result := g.Generate()
	assert.Regexp(t, uuidV4Pattern, result)
}

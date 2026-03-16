package infrastructure

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

type invalidStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestValidate_ValidStruct(t *testing.T) {
	s := validStruct{Name: "Alice", Email: "alice@example.com"}
	err := Validate(s)
	assert.NoError(t, err)
}

func TestValidate_InvalidStruct(t *testing.T) {
	s := invalidStruct{Name: "", Email: "not-an-email"}
	err := Validate(s)
	require.Error(t, err)
}

func TestFieldErrors_ExtractsFields(t *testing.T) {
	s := invalidStruct{Name: "", Email: "not-an-email"}
	err := Validate(s)
	require.Error(t, err)

	fe := FieldErrors(err)
	require.NotNil(t, fe)
	assert.Contains(t, fe, "name")
	assert.Equal(t, "required", fe["name"])
	assert.Contains(t, fe, "email")
	assert.Equal(t, "email", fe["email"])
}

func TestFieldErrors_NonValidationError(t *testing.T) {
	err := errors.New("some generic error")
	fe := FieldErrors(err)
	assert.Nil(t, fe)
}

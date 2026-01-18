package infrastructure

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

var V = validator.New()

func Validate(i any) error {
	return V.Struct(i)
}

func FieldErrors(err error) map[string]string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil // or empty map, depending on your contract
	}

	fe := make(map[string]string, len(ve))
	for _, e := range ve {
		fe[strings.ToLower(e.Field())] = e.Tag()
	}
	return fe
}

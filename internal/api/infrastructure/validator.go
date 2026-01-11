package infrastructure

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var V = validator.New()

func Validate(i any) error {
	return V.Struct(i)
}

func FieldErrors(err error) map[string]string {
	fe := map[string]string{}
	for _, e := range err.(validator.ValidationErrors) {
		fe[strings.ToLower(e.Field())] = e.Tag()
	}
	return fe
}

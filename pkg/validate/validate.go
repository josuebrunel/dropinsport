package validate

import "github.com/go-playground/validator/v10"

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func Struct(s any) error {
	return validate.Struct(s)
}

func Var(v any, tag string) error {
	return validate.Var(v, tag)
}

package utils

import "github.com/go-playground/validator/v10"

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i interface{}) validator.ValidationErrors {
	err := v.validator.Struct(i)
	if err == nil {
		return nil
	}

	errs := err.(validator.ValidationErrors)

	return errs
}

func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

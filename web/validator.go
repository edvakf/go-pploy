package web

import "gopkg.in/go-playground/validator.v9"

// https://echo.labstack.com/guide/request#validate-data

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

var Validator = CustomValidator{validator: validator.New()}

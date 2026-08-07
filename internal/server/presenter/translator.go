package presenter

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
)

func NewTranslator() ut.Translator {
	en := en.New()
	uni := ut.New(en)
	return uni.GetFallback()
}

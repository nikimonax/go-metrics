package presenter

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
)

var translator ut.Translator

func getTranslator() ut.Translator {
	if translator == nil {
		en := en.New()
		uni := ut.New(en)
		translator = uni.GetFallback()
	}

	return translator
}

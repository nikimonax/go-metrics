package presenter

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nikimonax/go-metrics/internal/lib/httpextra"
	"github.com/nikimonax/go-metrics/internal/model"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type JsonErrorPresenter struct {
	translator ut.Translator
	sugar      *zap.SugaredLogger
}

// RenderError implements [ErrorPresenter].
func (presenter *JsonErrorPresenter) Render(
	w http.ResponseWriter, err error, code int,
) {
	payload := model.NewErrorResponse()

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		payload.Errors = validationErrs.Translate(presenter.translator)
	} else {
		payload.Errors["main"] = err.Error()
	}

	w.Header().Set(httpextra.HDRContentType, httpextra.MIMEJSON)
	w.WriteHeader(code)

	err = json.NewEncoder(w).Encode(payload)

	if err == nil {
		return
	}

	if presenter.sugar == nil {
		return
	}

	presenter.sugar.Errorw(
		"failed to write http response",
		"err", err,
	)
}

func NewJsonErrorPresenter(logger *zap.Logger) ErrorPresenter {
	var sugar *zap.SugaredLogger

	if logger != nil {
		sugar = logger.Sugar()
	}

	return &JsonErrorPresenter{
		translator: getTranslator(),
		sugar:      sugar,
	}
}

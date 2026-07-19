package presenter

import "net/http"

type PlainTextErrorPresenter struct{}

// RenderError implements [ErrorPresenter].
func (presenter *PlainTextErrorPresenter) Render(
	w http.ResponseWriter, err error, code int,
) {
	http.Error(w, err.Error(), code)
}

func NewPlainTextErrorPresenter() ErrorPresenter {
	return new(PlainTextErrorPresenter)
}

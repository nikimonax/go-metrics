// Команда staticlint запускает набор статических анализаторов
// (стандартные vet-подобные, staticcheck, errcheck, noctx и собственный
// exitcheck) через multichecker.
//
// Использование:
//
//	staticlint [flags] [packages]
//
// Примеры:
//
//	staticlint ./...
//	staticlint -enable=SA1000,ST1000 ./...
//	staticlint -help
//
// Список доступных флагов и анализаторов выводится при запуске
// с флагом -help.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/nikimonax/go-metrics/tools/staticlint"
)

func main() {
	multichecker.Main(staticlint.Build()...)
}

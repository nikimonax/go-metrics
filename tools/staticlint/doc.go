// Package staticlint собирает набор статических анализаторов для multichecker.
//
// В набор входят:
//
//   - стандартные анализаторы из golang.org/x/tools/go/analysis/passes
//     (аналоги проверок `go vet`: appends, assign, bools, copylock, loopclosure,
//     printf, shadow, stringintconv, structtag, tests, timeformat, unmarshal
//     и другие);
//   - staticcheck - анализаторы категорий SA (staticcheck), S (simple),
//     QF (quickfix), а также unused (U1000);
//   - stylecheck - анализаторы категории ST, кроме исключённых
//     (ST1000, ST1020, ST1021, ST1022 - проверки комментариев и докстрингов);
//   - сторонние анализаторы: errcheck (непроверенные ошибки) и
//     noctx (http-запросы без context.Context);
//   - собственный анализатор exitcheck (запрет вызова os.Exit в функцих main).

package staticlint

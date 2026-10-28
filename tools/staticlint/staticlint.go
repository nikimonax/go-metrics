package staticlint

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/gofix"
	"golang.org/x/tools/go/analysis/passes/hostport"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inline"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/modernize"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"honnef.co/go/tools/unused"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/sonatard/noctx"

	"github.com/nikimonax/go-metrics/tools/exitcheck"
)

var standartAnalyzers = []*analysis.Analyzer{
	appends.Analyzer,
	asmdecl.Analyzer,
	assign.Analyzer,
	atomic.Analyzer,
	atomicalign.Analyzer,
	bools.Analyzer,
	buildssa.Analyzer,
	buildtag.Analyzer,
	cgocall.Analyzer,
	composite.Analyzer,
	copylock.Analyzer,
	ctrlflow.Analyzer,
	deepequalerrors.Analyzer,
	defers.Analyzer,
	directive.Analyzer,
	errorsas.Analyzer,
	// fieldalignment.Analyzer excluded,
	findcall.Analyzer,
	framepointer.Analyzer,
	gofix.Analyzer,
	hostport.Analyzer,
	httpmux.Analyzer,
	httpresponse.Analyzer,
	ifaceassert.Analyzer,
	inline.Analyzer,
	inspect.Analyzer,
	loopclosure.Analyzer,
	lostcancel.Analyzer,
	modernize.AnyAnalyzer,
	nilfunc.Analyzer,
	nilness.Analyzer,
	pkgfact.Analyzer,
	printf.Analyzer,
	reflectvaluecompare.Analyzer,
	shadow.Analyzer,
	shift.Analyzer,
	sigchanyzer.Analyzer,
	slog.Analyzer,
	sortslice.Analyzer,
	stdmethods.Analyzer,
	stdversion.Analyzer,
	stringintconv.Analyzer,
	structtag.Analyzer,
	testinggoroutine.Analyzer,
	tests.Analyzer,
	timeformat.Analyzer,
	unmarshal.Analyzer,
	unreachable.Analyzer,
	unsafeptr.Analyzer,
	unusedresult.Analyzer,
	unusedwrite.Analyzer,
	usesgenerics.Analyzer,
	waitgroup.Analyzer,
}

var excludedStyleChecks = map[string]struct{}{
	// Incorrect or missing package comment
	"ST1000": {},
	// The documentation of an exported function should start with the function's name
	"ST1020": {},
	// The documentation of an exported type should start with type's name
	"ST1021": {},
	// The documentation of an exported variable or constant should start with variable's name
	"ST1022": {},
}

var maxCap = len(standartAnalyzers) +
	len(staticcheck.Analyzers) +
	len(simple.Analyzers) +
	len(stylecheck.Analyzers) +
	len(quickfix.Analyzers) +
	4 // unused, errcheck, noctx, exitcheck

func Build() []*analysis.Analyzer {
	analyzers := make([]*analysis.Analyzer, 0, maxCap)

	// standart
	analyzers = append(analyzers, standartAnalyzers...)

	// staticcheck - SA
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	// staticcheck - S
	for _, v := range simple.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	// staticcheck - ST
	for _, v := range stylecheck.Analyzers {
		if _, ok := excludedStyleChecks[v.Analyzer.Name]; !ok {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// staticcheck - QF
	for _, v := range quickfix.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	// staticcheck - unused
	analyzers = append(analyzers, unused.Analyzer.Analyzer)

	// other - errcheck + noctx
	analyzers = append(analyzers, errcheck.Analyzer, noctx.Analyzer)

	// custom - exitcheck
	analyzers = append(analyzers, exitcheck.Analyzer)

	return analyzers
}

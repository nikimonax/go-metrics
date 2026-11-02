package pkg1

import "github.com/nikimonax/go-metrics/tools/structreset/testdata/pkg2"

// generate:reset
type ResettableStruct struct {
	pkg2.ResettableEmbed

	Chan    chan struct{}
	Func    func()
	I, J    int
	String  string
	StringP *string
	Slice   []int
	Array   [3]string
	Map     map[string]int
	Child   *ResettableStruct
}

/*
generate:reset
*/
type EmptyStruct struct{}

// want +1 `Using "generate:reset" is not applicable here`
// generate:reset
func NotResettable() {}

func main() {
	// want +1 `Using "generate:reset" is not applicable here`
	// generate:reset
	type InlineStruct struct {
		str string
	}
}

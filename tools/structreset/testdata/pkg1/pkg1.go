package pkg1

import "github.com/nikimonax/go-metrics/tools/structreset/testdata/pkg2"

// generate:reset
type ResetableStruct struct {
	pkg2.ResetableEmbed

	Chan    chan struct{}
	Func    func()
	I, J    int
	String  string
	StringP *string
	Slice   []int
	Array   [3]string
	Map     map[string]int
	Child   *ResetableStruct
}

/*
generate:reset
*/
type EmptyStruct struct{}

// want +1 `Using "generate:reset" is not applicable here`
// generate:reset
func NotResetable() {}

func main() {
	// want +1 `Using "generate:reset" is not applicable here`
	// generate:reset
	type InlineStruct struct {
		str string
	}
}

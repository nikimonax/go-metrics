package test

import (
	"testing"

	"github.com/nikimonax/go-metrics/tools/structreset/testdata/pkg1"
	"github.com/nikimonax/go-metrics/tools/structreset/testdata/pkg2"
	"github.com/stretchr/testify/assert"
)

const markerValue = 999

type customResetable int

func (r *customResetable) Reset() {
	*r = markerValue
}

func Test(t *testing.T) {
	cr := customResetable(5)
	ch := make(chan struct{})
	f := func() {}
	s := "World"

	resetable1 := pkg1.ResetableStruct{
		String:  "Hello",
		StringP: &s,
		Slice:   []int{1, 2, 3, 4},
		Array:   [3]string{"1", "2", "3"},
		Map:     map[string]int{"1": 1, "2": 2},
	}

	resetable2 := pkg1.ResetableStruct{
		ResetableEmbed: pkg2.ResetableEmbed{
			EmbedChild: &cr,
		},

		Chan:  ch,
		Func:  f,
		I:     42,
		J:     67,
		Child: &resetable1,
	}

	var hidden any = &resetable2

	resetable, ok := hidden.(interface{ Reset() })

	if !ok {
		t.Fatal(`"Reset() method was not generated"`)
	}

	resetable.Reset()

	assert.Equal(t, "", resetable1.String)
	assert.Equal(t, &s, resetable1.StringP)
	assert.Equal(t, "", *resetable1.StringP)
	assert.Len(t, resetable1.Slice, 0)
	assert.Equal(t, [3]string{"", "", ""}, resetable1.Array)
	assert.Equal(t, map[string]int{}, resetable1.Map)
	assert.NotNil(t, resetable2.EmbedChild)
	assert.Equal(t, customResetable(markerValue), cr)
	assert.Equal(t, ch, resetable2.Chan)
	assert.NotNil(t, resetable2.Func)
	assert.Equal(t, 0, resetable2.I)
	assert.Equal(t, 0, resetable2.J)
	assert.Equal(t, &resetable1, resetable2.Child)
}

package test

import (
	"testing"

	"github.com/nikimonax/go-metrics/tools/structreset/testdata/pkg1"
	"github.com/nikimonax/go-metrics/tools/structreset/testdata/pkg2"
	"github.com/stretchr/testify/assert"
)

const markerValue = 999

type customResettable int

func (r *customResettable) Reset() {
	*r = markerValue
}

func Test(t *testing.T) {
	cr := customResettable(5)
	ch := make(chan struct{})
	f := func() {}
	s := "World"

	resettable1 := pkg1.ResettableStruct{
		String:  "Hello",
		StringP: &s,
		Slice:   []int{1, 2, 3, 4},
		Array:   [3]string{"1", "2", "3"},
		Map:     map[string]int{"1": 1, "2": 2},
	}

	resettable2 := pkg1.ResettableStruct{
		ResettableEmbed: pkg2.ResettableEmbed{
			EmbedChild: &cr,
		},

		Chan:  ch,
		Func:  f,
		I:     42,
		J:     67,
		Child: &resettable1,
	}

	var hidden any = &resettable2

	resettable, ok := hidden.(interface{ Reset() })

	if !ok {
		t.Fatal(`"Reset() method was not generated"`)
	}

	resettable.Reset()

	assert.Equal(t, "", resettable1.String)
	assert.Equal(t, &s, resettable1.StringP)
	assert.Equal(t, "", *resettable1.StringP)
	assert.Len(t, resettable1.Slice, 0)
	assert.Equal(t, [3]string{"", "", ""}, resettable1.Array)
	assert.Equal(t, map[string]int{}, resettable1.Map)
	assert.NotNil(t, resettable2.EmbedChild)
	assert.Equal(t, customResettable(markerValue), cr)
	assert.Equal(t, ch, resettable2.Chan)
	assert.NotNil(t, resettable2.Func)
	assert.Equal(t, 0, resettable2.I)
	assert.Equal(t, 0, resettable2.J)
	assert.Equal(t, &resettable1, resettable2.Child)
}

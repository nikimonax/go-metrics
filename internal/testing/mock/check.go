package mock

import (
	"reflect"

	"github.com/stretchr/testify/mock"
)

var (
	Anything       = mock.Anything
	AnythingOfType = mock.AnythingOfType
	IsType         = mock.IsType
	MatchedBy      = mock.MatchedBy
)

func IsImplements(i any) func(any) bool {
	interfaceType := reflect.TypeOf(i).Elem()

	return func(obj any) bool {
		objType := reflect.TypeOf(obj)
		return objType.Implements(interfaceType)
	}
}

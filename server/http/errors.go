package http

import "fmt"

func newErrMsgInvalidMethod(got string, want string) string {
	return fmt.Sprintf("invalid http method: got '%s', want '%s'", got, want)
}

func newErrMsgInvalidContentType(got string, want string) string {
	return fmt.Sprintf("invalid content type: got '%s', want '%s'", got, want)
}

func newErrMsgParamNotProvided(location string, name string) string {
	return fmt.Sprintf("%s parameter '%s' not provided", location, name)
}

func newErrMsgParamNotValid(name string) string {
	return fmt.Sprintf("parameter '%s' not valid", name)
}

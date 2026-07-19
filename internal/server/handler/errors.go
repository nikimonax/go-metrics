package handler

import "fmt"

func newErrMsgParamNotProvided(location string, name string) string {
	return fmt.Sprintf("%s parameter '%s' not provided", location, name)
}

func newErrMsgParamNotValid(name string) string {
	return fmt.Sprintf("parameter '%s' not valid", name)
}

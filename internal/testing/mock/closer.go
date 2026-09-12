package mock

import "io"

type MockCloser struct {
	reader    io.Reader
	closed    int
	closeErrs []error
}

var (
	_ io.Reader = (*MockCloser)(nil)
	_ io.Closer = (*MockCloser)(nil)
)

func (mc *MockCloser) Read(p []byte) (int, error) {
	return mc.reader.Read(p)
}

func (mc *MockCloser) Close() error {
	mc.closed++

	var err error

	if len(mc.closeErrs) > 0 {
		err = mc.closeErrs[0]
		mc.closeErrs = mc.closeErrs[1:]
	}

	return err
}

func (mc *MockCloser) Closed() bool {
	return mc.closed > 0
}

func (mc *MockCloser) ClosedNumber() int {
	return mc.closed
}

func (mc *MockCloser) SetCloseErrs(errs ...error) {
	mc.closeErrs = errs
}

func (mc *MockCloser) AddCloseErrs(errs ...error) {
	mc.closeErrs = append(mc.closeErrs, errs...)
}

func NewMockCloser(reader io.Reader) *MockCloser {
	return &MockCloser{reader: reader}
}

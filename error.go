package goeureka

type Error struct {
	ErrorNo  int
	ErrorMsg string
}

func ErrorNew(err string) *Error {
	return &Error{ErrorNo: -1, ErrorMsg: err}
}
func (e *Error) Error() string {
	return e.ErrorMsg
}

package echoerror

type HTTPError interface {
	error
	Code() int
}

type httpError struct {
	code int
	err  error
}

func (e *httpError) Error() string {
	return e.err.Error()
}

func (e *httpError) Code() int {
	return e.code
}

func NewHttp(code int, err error) HTTPError {
	return &httpError{code, err}
}

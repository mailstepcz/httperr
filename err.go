package httperr

import (
	"database/sql"
	"errors"
	"net/http"
)

// HTTPError is an error convertible into an HTTP error.
type HTTPError interface {
	error
	HTTPStatus() int
}

// HTTPErrorEnvelope is an error convertible into an HTTP error.
type HTTPErrorEnvelope struct {
	err    error
	status int
}

// New creates a new error convertible into an HTTP error.
func New(err string, status int) HTTPErrorEnvelope {
	return Wrap(errors.New(err), status)
}

// Wrap wraps an error into one convertible into an HTTP error.
func Wrap(err error, status int) HTTPErrorEnvelope {
	return HTTPErrorEnvelope{
		err:    err,
		status: status,
	}
}

func (err HTTPErrorEnvelope) Error() string {
	return err.err.Error()
}

func (err HTTPErrorEnvelope) Unwrap() error {
	return err.err
}

// HTTPStatus is the corresponding HTTP error code.
func (err HTTPErrorEnvelope) HTTPStatus() int {
	return err.status
}

var _ HTTPError = HTTPErrorEnvelope{}

// HTTPStatus gets the corresponding HTTP status.
func HTTPStatus(err error) int {
	if c, ok := getHTTPStatus(err); ok {
		return c
	}

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

func getHTTPStatus(err error) (int, bool) {
	if err, ok := err.(HTTPError); ok {
		return err.HTTPStatus(), true
	}

	if err, ok := err.(wrappedErr); ok {
		return getHTTPStatus(err.Unwrap())
	}

	if err, ok := err.(wrappedErrs); ok {
		errs := err.Unwrap()
		statuses := make([]int, 0, len(errs))
		for _, err := range errs {
			if c, ok := getHTTPStatus(err); ok {
				statuses = append(statuses, c)
			}
		}
		if len(statuses) > 1 {
			panic("more wrapped errors provide an HTTP status") // what should we do here?
		}
		if len(statuses) == 0 {
			return 0, false
		}
		return statuses[0], true
	}

	return 0, false
}

type wrappedErr interface {
	error
	Unwrap() error
}

type wrappedErrs interface {
	error
	Unwrap() []error
}

package httperr

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/mailstepcz/serr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
func New(err string, status int, attrs ...serr.Attributed) HTTPErrorEnvelope {
	return Wrap("", errors.New(err), status, attrs...)
}

// Wrap wraps an error into one convertible into an HTTP error.
func Wrap(msg string, err error, status int, attrs ...serr.Attributed) HTTPErrorEnvelope {
	if len(attrs) > 0 {
		err = serr.Wrap(msg, err, attrs...)
	} else if msg != "" {
		err = fmt.Errorf("%s: %w", msg, err)
	}
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

	if status, ok := status.FromError(err); ok {
		return grpcCodeToStatusCode(status.Code())
	}

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

func grpcCodeToStatusCode(code codes.Code) int {
	switch code {
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.InvalidArgument:
		return http.StatusBadRequest
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

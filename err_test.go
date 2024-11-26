package httperr

import (
	"errors"
	"net/http"
	"testing"

	"github.com/mailstepcz/serr"
	"github.com/stretchr/testify/require"
)

func TestHTTPStatus(t *testing.T) {
	req := require.New(t)

	dummyErr := Wrap("", errors.ErrUnsupported, http.StatusNotImplemented)

	req.Equal(http.StatusNotImplemented, HTTPStatus(dummyErr))

	req.Equal("unsupported operation", dummyErr.Error())

	req.True(errors.Is(dummyErr, errors.ErrUnsupported))
}

func TestHTTPWrappedError(t *testing.T) {
	req := require.New(t)

	dummyErr := serr.Wrap("wrapped", Wrap("", errors.ErrUnsupported, http.StatusNotImplemented))

	req.Equal(http.StatusNotImplemented, HTTPStatus(dummyErr))

	req.Equal("wrapped: unsupported operation", dummyErr.Error())

	req.True(errors.Is(dummyErr, errors.ErrUnsupported))
}

func TestHTTPWrappedErrors(t *testing.T) {
	req := require.New(t)

	dummyErr := errors.Join(errors.New("some error"), serr.Wrap("wrapped", Wrap("", errors.ErrUnsupported, http.StatusNotImplemented)))

	req.Equal(http.StatusNotImplemented, HTTPStatus(dummyErr))

	req.Equal("some error\nwrapped: unsupported operation", dummyErr.Error())

	req.True(errors.Is(dummyErr, errors.ErrUnsupported))
}

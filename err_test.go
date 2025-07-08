package httperr

import (
	"errors"
	"net/http"
	"testing"

	"github.com/mailstepcz/grpcerr"
	"github.com/mailstepcz/serr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
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

func TestSerrWrappedGrpcError(t *testing.T) {
	req := require.New(t)

	baseError := grpcerr.New("first error", codes.NotFound)
	wrappedError := serr.Wrap("wrapped error", baseError)

	req.Equal(http.StatusNotFound, HTTPStatus(wrappedError))

	req.Equal("wrapped error: first error", wrappedError.Error())

	req.ErrorIs(wrappedError, baseError)
}

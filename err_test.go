package httperr

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/mailstepcz/grpcerr"
	"github.com/mailstepcz/serr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	req.EqualError(wrappedError, "wrapped error: first error")

	req.ErrorIs(wrappedError, baseError)
}

func TestHTTPStatusMapping(t *testing.T) {
	tcs := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "explicit envelope status wins",
			err:  Wrap("", errors.ErrUnsupported, http.StatusNotImplemented),
			want: http.StatusNotImplemented,
		},
		{
			name: "unrelated error",
			err:  errors.New("boom"),
			want: http.StatusInternalServerError,
		},
		{
			name: "sql.ErrNoRows",
			err:  sql.ErrNoRows,
			want: http.StatusNotFound,
		},
		{
			name: "gRPC status carrying codes.NotFound",
			err:  status.Error(codes.NotFound, "not found"),
			want: http.StatusNotFound,
		},
		{
			name: "grpcerr error tagged codes.Canceled",
			err:  grpcerr.New("caller went away", codes.Canceled),
			want: StatusClientClosedRequest,
		},
		{
			name: "gRPC status carrying codes.Canceled",
			err:  status.Error(codes.Canceled, "context canceled"),
			want: StatusClientClosedRequest,
		},
		{
			name: "serr-wrapped gRPC status carrying codes.Canceled",
			err:  serr.Wrap("getting users by ids", status.Error(codes.Canceled, "context canceled")),
			want: StatusClientClosedRequest,
		},
		{
			name: "plain context.Canceled",
			err:  context.Canceled,
			want: StatusClientClosedRequest,
		},
		{
			name: "serr-wrapped context.Canceled",
			err:  serr.Wrap("downloading file", context.Canceled),
			want: StatusClientClosedRequest,
		},
		{
			name: "gRPC status carrying codes.DeadlineExceeded is not a cancellation",
			err:  status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
			want: http.StatusInternalServerError,
		},
		{
			name: "context.DeadlineExceeded is not a cancellation",
			err:  context.DeadlineExceeded,
			want: http.StatusInternalServerError,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, HTTPStatus(tc.err))
		})
	}
}

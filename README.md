# httperr
Go support for HTTP error codes

## Status mapping

`HTTPStatus` resolves the status in this order:

1. `HTTPError.HTTPStatus()` on the error or anywhere in its `Unwrap` chain, or a
   `grpcerr.Convertible` code translated via the table below.
2. A gRPC status found in the chain, translated via the table below.
3. `sql.ErrNoRows` → `404`.
4. `serr.IsCanceled(err)` → `499`.
5. Fallback → `500`.

| gRPC code | HTTP |
| --- | --- |
| `NotFound` | 404 |
| `Unauthenticated` | 401 |
| `PermissionDenied` | 403 |
| `Unimplemented` | 501 |
| `InvalidArgument` | 400 |
| `AlreadyExists` | 409 |
| `FailedPrecondition` | 400 |
| `Canceled` | 499 |
| anything else | 500 |

### 499 — the caller went away

`Canceled` maps to `StatusClientClosedRequest` (499, nginx's non-standard code),
not to 400. A cancelled request is not a bad request: the client disconnected
before the answer was ready, and nobody is left to read the response. Keeping the
two apart lets monitoring ignore cancellations without also ignoring genuine
validation failures.

Step 4 catches the cancellations that never became a gRPC status — a cancelled
Postgres statement (SQLSTATE `57014`), a `context.Canceled` from an HTTP or S3
client — which would otherwise fall through to 500.

`context.DeadlineExceeded` is deliberately **not** treated as a cancellation and
still yields 500; a timeout is a real signal worth reporting.

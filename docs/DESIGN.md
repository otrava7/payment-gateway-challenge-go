# Design considerations & assumptions

This document captures the key design decisions behind the payment gateway and
the assumptions made where the requirements left room for interpretation. For
how to build, run and use the service, see the [README](../README.md).

## Architecture & layering

The code is split into thin, single-responsibility layers:

```
HTTP request → api → service → bank (acquiring bank, HTTP)
                       └──────→ repository (storage)
```

- **`api`** owns HTTP concerns only: routing, decoding, and mapping a domain
  outcome to a status code and body. It contains no payment rules.
- **`service`** owns the business logic: validation and the orchestration of
  *validate → authorize → store*.
- **`bank`** owns the acquiring-bank protocol: the HTTP call and the translation
  between our domain model and the bank's wire format.
- **`repository`** owns storage.

Each boundary is a **consumer-defined interface** (`api.PaymentsService`,
`service.AcquiringBank`). The consumer declares the narrow contract it needs, the
concrete type satisfies it structurally, and tests substitute a mock. This keeps
dependencies pointing inward and makes every layer testable in isolation.

## Where validation lives

Payment validation lives in the **service**, not the HTTP layer. A failed rule is
a *domain outcome* (the payment is **Rejected** and not forwarded to the bank),
not a transport error — so the service is the single source of truth for "what
makes a valid payment", and the handler simply maps the result to HTTP.

The one exception kept at the edge is a request body that cannot be decoded at
all: that is a genuine transport error, so the handler rejects it directly.

## Payment outcomes

Three outcomes, matching the requirements:

| Outcome | Meaning | HTTP | Stored? |
| --- | --- | --- | --- |
| `Authorized` | Bank authorized the payment | `201` | Yes |
| `Declined` | Bank declined the payment | `201` | Yes |
| `Rejected` | Failed validation, never sent to the bank | `400` | No |

`Authorized` and `Declined` are both *created payments* — real outcomes of a bank
call — so both are stored and retrievable. `Rejected` means no payment could be
created, so there is nothing to retrieve. `PaymentStatus` is modelled as a typed
enum in `models` rather than bare strings, so the values live in one place.

## Error handling

A failure to reach the acquiring bank (network error, non-200 response) is
**distinct from both a decline and a rejection**. The request was valid, so it is
neither `Rejected` (which implies invalid input) nor a created payment. The
service returns a plain error, which the handler maps to `500` while logging the
underlying cause — the caller gets an opaque message, operators get the detail.

## Observability

- **Structured logging** uses the standard library `log/slog` (JSON), so no
  logging dependency is introduced.
- **Correlation**: `chi`'s `RequestID` middleware assigns each request an id, and
  a custom slog handler stamps it onto every log line — including those emitted
  deeper in the stack — so one payment can be traced end to end.
- **Context propagation**: `context.Context` flows from the HTTP request through
  the service to the bank's HTTP call (`http.NewRequestWithContext`). This is
  threaded through *all* service methods for consistency, enabling cancellation,
  deadlines, and future tracing.
- **PCI-DSS**: the full card number and CVV are never logged or returned. Only the
  last four digits, status, amount and currency appear in logs and responses.

## Identifiers

Payment ids are random v4 GUIDs (`google/uuid`). The requirements allow any
format; a GUID avoids enumerable ids and needs no coordination.

## Storage

Storage is a simple in-memory repository, which is sufficient for the challenge.
It is intentionally behind the service so it could be replaced with a real,
durable store without touching the business logic. Being in-memory, payments do
not survive a restart and the implementation is single-process.

## Testing strategy

- **Unit tests** cover each layer in isolation using mocks for the collaborating
  interface (mock service for handlers, mock bank for the service). They run with
  the race detector and need no external dependencies.
- **End-to-end tests** (`internal/e2e`, behind the `e2e` build tag) drive the
  whole stack over real HTTP against the mountebank bank simulator, covering the
  authorized, declined, rejected and not-found paths.

The build tag keeps the default `go test ./...` fast and dependency-free, while
`make e2e` runs the simulator-backed suite.

## Build & versioning

`version`/`commit`/`date` in `main` are `-ldflags -X` injection points. On a
tagged release GoReleaser stamps them; `make build` stamps the same values from
git for local and CI builds, so a running binary always reports its identity (and
emits it on startup). Unstamped `go run` builds correctly show the `dev` defaults.

## Assumptions

Where the requirements were open to interpretation:

- **Supported currencies** are `USD`, `GBP`, `EUR` — the requirement is "validate
  against no more than 3 currency codes", and these three are used.
- **Amount sign is not constrained.** The requirement only states the amount must
  be required and an integer, so the gateway does not reject zero or negative
  amounts — whether such an amount is acceptable is left to the acquiring bank.
  Refunds are treated as a separate, future operation rather than a negative
  payment.
- **`Rejected` HTTP shape**: a rejected payment returns `400` with a
  `payment_status` of `Rejected` and an `error` message describing the failed
  rule. The requirements define `Rejected` as an outcome but not its exact HTTP
  representation.
- The acquiring bank is reached over HTTP at `ACQUIRING_BANK_URL` (default
  `http://localhost:8080`, the simulator).

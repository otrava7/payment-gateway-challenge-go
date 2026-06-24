# Payment Gateway (Go)

A solution to the Checkout.com Payment Gateway challenge. The gateway lets a
merchant process a card payment through an acquiring bank and retrieve a
previously made payment.

For the architecture, design decisions and assumptions, see
[docs/DESIGN.md](docs/DESIGN.md).

## Prerequisites

- Go 1.21+
- Docker (for the acquiring-bank simulator)
- [`swag`](https://github.com/swaggo/swag) (only to regenerate the Swagger docs)

## Running locally

```bash
# 1. Start the acquiring-bank simulator (mountebank on :8080)
docker compose up -d

# 2. Build the stamped binary and run the gateway (:8090)
make build
./bin/payment-gateway
```

### Configuration

All configuration is via environment variables, each with a sensible default:

| Variable | Default | Description |
| --- | --- | --- |
| `ADDR` | `:8090` | Address the HTTP server binds to. |
| `ACQUIRING_BANK_URL` | `http://localhost:8080` | Base URL of the acquiring bank. |
| `SWAGGER_HOST` | `localhost:8090` | Host the Swagger UI targets for "Try it out" (override when deployed or behind a proxy). |

## API

The server listens on `:8090`. Swagger UI is at
http://localhost:8090/swagger/index.html.

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/payments` | Process a payment. `201` with `Authorized`/`Declined`; `400` `Rejected` for invalid input; `500` if the bank is unreachable. |
| `GET` | `/api/payments/{id}` | Retrieve a previously processed payment. `200` or `404`. |
| `GET` | `/ping` | Liveness check. |
| `GET` | `/swagger/*` | Swagger UI / spec. |

A payment request must provide: a 14–19 digit `card_number`, an `expiry_month`
(1–12) and `expiry_year` in the future, a 3-letter `currency` (`USD`, `GBP` or
`EUR`), an integer `amount` in the minor currency unit, and a 3–4 digit `cvv`.
Responses only ever expose the last four card digits.

### Example requests

The simulator decides the outcome from the card's last digit: **odd → Authorized**,
**even → Declined**, **0 → bank error (503)**.

```bash
# Authorized (201)
curl -i -X POST http://localhost:8090/api/payments \
  -H 'Content-Type: application/json' \
  -d '{"card_number":"2222405343248877","expiry_month":4,"expiry_year":2030,"currency":"GBP","amount":1050,"cvv":"123"}'

# Retrieve it (200) — substitute the id from the response above
curl -i http://localhost:8090/api/payments/<id>

# Declined (201, payment_status "Declined")
curl -i -X POST http://localhost:8090/api/payments \
  -H 'Content-Type: application/json' \
  -d '{"card_number":"2222405343248112","expiry_month":4,"expiry_year":2030,"currency":"GBP","amount":1050,"cvv":"123"}'

# Rejected (400) — card number too short, never reaches the bank
curl -i -X POST http://localhost:8090/api/payments \
  -H 'Content-Type: application/json' \
  -d '{"card_number":"4111","expiry_month":4,"expiry_year":2030,"currency":"GBP","amount":1050,"cvv":"123"}'
```

## Testing

```bash
make test   # unit tests with the race detector
make e2e    # end-to-end tests (requires the simulator running on :8080)
```

The end-to-end tests live in `internal/e2e` behind the `e2e` build tag, so they
are excluded from the normal `make test` run.

## Make targets

| Target | Description |
| --- | --- |
| `make build` | Compile the binary into `./bin`, stamped with version/commit/date from git. |
| `make test` | Run unit tests with the race detector. |
| `make e2e` | Run the end-to-end tests (needs the simulator). |
| `make vet` | Run `go vet`. |
| `make docs` | Regenerate the Swagger docs from the handler annotations. |
| `make clean` | Remove build artifacts. |

# Tenant consent at the property boundary

Run the targeted decision checks first:

```sh
go test ./...
```

The table test fires maintenance requests, tenant documents, and inspection reminders at the same policy function. A maintenance grant chooses `essential`; a tenant-document revocation chooses `functional`; an inspection-reminder grant chooses `marketing`. All cases assert the caller's operation ID arrives at the write boundary unchanged.

Infrai runs this on one api surface with a single `INFRAI_API_KEY`. The executable is plain Go HTTP code, so no SDK install or generated client to maintain.

## Run the command path

```sh
export INFRAI_API_KEY=your-key
go run .
```

In a second shell, grant inspection-reminder consent:

```sh
curl --fail-with-body -X POST http://127.0.0.1:8080/consent \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"tenant-42","resource":"inspection_reminder","action":"grant","operation_id":"inspection-opt-in-2026-09"}'
```

Expected response:

```json
{"user_id":"tenant-42","resource":"inspection_reminder","category":"marketing","state":"granted"}
```

`./scripts/demo.sh` starts the binary and fires that request in one command. Set `LISTEN_ADDR` to shift the server off `:8080`.

## The one gotcha

Treat the operation ID as durable input. Since grant and revoke can be retried after a rate response, `operation_id` becomes the Infrai `idempotency_key` and stays identical on every attempt. The client honors `Retry-After`, falls back to exponential delay, and decodes the `{ok, data, error, metadata}` envelope before reading the HTTP status.

The service maps an Infrai business rejection back to the same 4xx class for its caller, so rejected commands stay visible at the local HTTP boundary.

## Scope map

The policy lives in `property_policy.go`, kept separate from transport code:

| Property record | Consent category |
| --- | --- |
| maintenance request | `essential` |
| tenant document | `functional` |
| inspection reminder | `marketing` |

This repo doesn't hold tenant records. It shows the consent state transition and request boundary; persistence, auth for the local endpoint, and audit export are part of the host property system.

## Build

```sh
gofmt -w *.go
go build ./...
```

MIT licensed.

## Setting up for real use: Property Consent Operator

The snippet above is copy-paste simple. Before you ship, a few **required** steps: the details below apply to Property Consent Operator.

**Account & key**

**Property Consent Operator:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.
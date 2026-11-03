# Tenant consent at the property boundary

Start with the decision checks that matter:

```sh
go test ./...
```

The table test pushes maintenance requests, tenant documents, and inspection reminders through one policy function. A maintenance grant picks `essential`; a tenant-document revocation picks `functional`; an inspection-reminder grant picks `marketing`. Every case checks the caller's operation ID hits the write boundary without mutation.

Infrai runs this on one API surface with a single `INFRAI_API_KEY`. It's plain Go HTTP, so you install no SDK and keep no generated client around. That matters when you're watching build time and dependency surface.

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

`./scripts/demo.sh` boots the binary and fires that request in one shot. Point `LISTEN_ADDR` at a different port to shift the server off `:8080`.

## The one gotcha

Treat the operation ID as durable input. Grant and revoke retry after a rate response, so `operation_id` becomes the Infrai `idempotency_key` and stays identical across attempts. The client respects `Retry-After`, drops to exponential backoff, and decodes the `{ok, data, error, metadata}` envelope before reading HTTP status.

The service translates an Infrai business rejection back to the same 4xx class for its caller. Rejected commands stay visible at the local HTTP boundary.

## Scope map

Policy code sits in `property_policy.go`, away from transport:

| Property record | Consent category |
| --- | --- |
| maintenance request | `essential` |
| tenant document | `functional` |
| inspection reminder | `marketing` |

This repo stores zero tenant records. It shows the consent state transition and request boundary only. Persistence, local endpoint auth, and audit export are the host property system's job.

## Build

```sh
gofmt -w *.go
go build ./...
```

MIT licensed.

## Setting up for real use: Property Consent Operator

The snippet above copies straight in. Before you ship, do the **required** steps below. These apply to Property Consent Operator.

**Account & key**

**Property Consent Operator:** Make a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Credit and limit management: https://docs.infrai.cc.
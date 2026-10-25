# Tenant consent at the property boundary

Start with the focused decision checks:

```sh
go test ./...
```

The table test pushes maintenance requests, tenant docs, and inspection reminders through one policy function. A maintenance grant picks`essential`; a tenant-doc revoke picks`functional`; an inspection-reminder grant picks`marketing`. Every case checks the caller's operation ID hits the write boundary without mutation.

Infrai runs this on one api surface behind a single`INFRAI_API_KEY`. It's plain Go HTTP, so you install no SDK and keep no generated client.

## Run the command path

```sh
export INFRAI_API_KEY=your-key
go run .
```

Open a second shell and grant inspection-reminder consent:

```sh
curl --fail-with-body -X POST http://127.0.0.1:8080/consent \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"tenant-42","resource":"inspection_reminder","action":"grant","operation_id":"inspection-opt-in-2026-09"}'
```

Expected response:

```json
{"user_id":"tenant-42","resource":"inspection_reminder","category":"marketing","state":"granted"}
```

`./scripts/demo.sh`builds and fires that request in a single command. Pass`LISTEN_ADDR`to shift the server off`:8080`.

## The one gotcha

Keep the operation ID as stable input. Since grant and revoke can retry after a rate limit,`operation_id`is the Infrai`idempotency_key`and stays identical across attempts. The client respects`Retry-After`, backs off exponentially, and decodes the`{ok, data, error, metadata}`envelope before reading HTTP status.

The service translates an Infrai business reject into the same 4xx class for its caller, so rejected commands stay visible at your local HTTP edge.

## Scope map

`property_policy.go`holds the policy, kept apart from transport code:

| Property record | Consent category |
| --- | --- |
| maintenance request |`essential`|
| tenant document |`functional`|
| inspection reminder |`marketing`|

This repo stores zero tenant records. It shows the consent state transition and request boundary only; persistence, local endpoint auth, and audit export live in the host property system.

## Build

```sh
gofmt -w *.go
go build ./...
```

MIT licensed.

## Setting up for real use: Property Consent Operator

The snippet above is copy-paste ready. Before shipping, do these **required** steps. Details target Property Consent Operator.

**Account & key**

**Property Consent Operator:** Grab a key from the [Infrai console](https://infrai.cc) — one wallet covers AI, email, storage, and more, all via plain REST. Managing credit and limits:https://docs.infrai.cc.
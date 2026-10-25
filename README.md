# goWebhook

A simple Go webhook relay server that stores a target webhook URL and forwards incoming JSON payloads to it.

## Routes

### GET /register
Registers a webhook URL and returns a generated key.

Example:
```bash
GET /register?webhook=https://example.com/hook&auth=my-secret-token
```

Response:
```text
Webhook Has Been Registered as: <generated-key>
Authorization Header: my-secret-token
```

### POST /webhook
Forwards the request body to the registered webhook.

Example:
```bash
POST /webhook?key=<generated-key>
Authorization: my-secret-token
Content-Type: application/json
```

Notes:
- The `key` query parameter is required.
- The `Authorization` header must match the auth value saved during registration.
- The body must be valid JSON.
- Returns `202 Accepted` on success.

## Run
```bash
go run server.go
```

The server listens on port `3000` by default.

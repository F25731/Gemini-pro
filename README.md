# Banana Image Wrapper

OpenAI-compatible image wrapper for NewAPI -> GetToken model API.

## Endpoints

- `GET /v1/models`
- `POST /v1/images/generations`
- `POST /v1/images/edits`
- `GET /health`
- `GET /admin`

## Models

Image models:

- `banana-pro-1k`, `banana-pro-2k`, `banana-pro-4k`
- `banana2-512`, `banana2-1k`, `banana2-2k`, `banana2-4k`

Only the image models listed above are exposed.

## NewAPI Config

Create an OpenAI-compatible channel:

```text
Base URL: https://your-wrapper-domain/v1
API Key: WRAPPER_API_KEY from .env
Models: banana-pro-1k,banana-pro-2k,banana-pro-4k,banana2-512,banana2-1k,banana2-2k,banana2-4k
```

## Mapping

- `/v1/images/generations` -> text-to-image
- `/v1/images/edits` -> image-to-image
- `/v1/images/generations` with `image_url` or `imageUrls` -> image-to-image
- multipart `image` is uploaded to GetToken first, then sent as `imageUrls`
- upstream task IDs alone are never returned as successful generations
- only final image URLs or `b64_json` outputs count as success

## Failure Response

The wrapper returns OpenAI-style errors so NewAPI can mark the generation as failed and refund pre-consumed quota:

```json
{
  "error": {
    "message": "upstream failure reason",
    "type": "invalid_request_error",
    "code": "upstream_failed"
  }
}
```

## Deploy

```bash
cp .env.example .env
docker compose up -d --build
```

Admin UI:

```text
http://SERVER_IP:HOST_PORT/admin
```

The admin page lets you update the GetToken upstream API key and view latency, success rate, image task stats, workers, and queue status.

## Important Env Vars

```env
WRAPPER_API_KEY=key-for-newapi
ADMIN_TOKEN=admin-ui-token
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin-ui-password
BANANA_API_KEY=gettoken-enterprise-api-key
MAX_WORKERS=2000
MAX_QUEUE=50000
```

Higher concurrency does not bypass upstream account limits. If GetToken returns rate-limit or unavailable-account errors, reduce `MAX_WORKERS` or add upstream capacity.

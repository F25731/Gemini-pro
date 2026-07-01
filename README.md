# Banana / Veo Wrapper

OpenAI-compatible wrapper for NewAPI -> GetToken model API.

## Endpoints

- `GET /v1/models`
- `POST /v1/images/generations`
- `POST /v1/images/edits`
- `POST /v1/videos/generations`
- `GET /health`
- `GET /admin`

## Models

Image models:

- `banana-pro-1k`, `banana-pro-2k`, `banana-pro-4k`
- `banana2-512`, `banana2-1k`, `banana2-2k`, `banana2-4k`

Video models:

- `veo3.1-pro-720p`, `veo3.1-pro-1080p`, `veo3.1-pro-4k`
- `veo3.1-fast-720p`, `veo3.1-fast-1080p`, `veo3.1-fast-4k`

GPT Image 2 is intentionally not exposed.

## NewAPI Config

Create an OpenAI-compatible channel:

```text
Base URL: https://your-wrapper-domain/v1
API Key: WRAPPER_API_KEY from .env
Models: banana-pro-1k,banana-pro-2k,banana-pro-4k,banana2-512,banana2-1k,banana2-2k,banana2-4k,veo3.1-pro-720p,veo3.1-pro-1080p,veo3.1-pro-4k,veo3.1-fast-720p,veo3.1-fast-1080p,veo3.1-fast-4k
```

## Mapping

Image:

- `/v1/images/generations` -> text-to-image
- `/v1/images/edits` -> image-to-image
- `/v1/images/generations` with `image_url` or `imageUrls` -> image-to-image
- multipart `image` is uploaded to GetToken first, then sent as `imageUrls`
- image endpoints submit an upstream task and poll internally, then return OpenAI-style `data[].url`

Video:

- `/v1/videos/generations` without images -> text-to-video
- with `imageUrls` -> image-to-video
- with `firstFrameUrl` and `lastFrameUrl` -> start/end-to-video for `veo3.1-pro-*`
- video endpoints submit an upstream task and poll internally, then return OpenAI-style `data[].url`
- only final image/video URLs count as success; upstream task IDs alone are never returned as successful generations

## Deploy

```bash
cp .env.example .env
docker compose up -d --build
```

Admin UI:

```text
http://SERVER_IP:HOST_PORT/admin
```

The admin page lets you update the GetToken upstream API key and view latency, success rate, image task stats, video task stats, workers, and queue status.

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

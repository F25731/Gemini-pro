# Banana Pro Wrapper

Standalone OpenAI-compatible image wrapper for NewAPI -> GetToken Banana Pro.

## What It Does

It exposes:

- `GET /v1/models`
- `POST /v1/images/generations`
- `POST /v1/images/edits`
- `GET /health`
- `GET /admin`

Models only represent resolution:

- `banana-pro-1k`
- `banana-pro-2k`
- `banana-pro-4k`

Text-to-image and image-to-image are not separate models.

- `/v1/images/generations` -> GetToken `/v1/banana_pro/text-to-image`
- `/v1/images/edits` -> GetToken `/v1/banana_pro/image-to-image`
- `/v1/images/generations` with `image_url` or `imageUrls` is also treated as image-to-image

## NewAPI Config

Create one OpenAI-compatible channel:

```text
Base URL: https://your-wrapper-domain/v1
API Key: WRAPPER_API_KEY from .env
Models: banana-pro-1k,banana-pro-2k,banana-pro-4k
```

Users only choose the resolution model. The wrapper decides text-to-image or image-to-image from the API path/request body.

## Deploy

```bash
cp .env.example .env
docker compose up -d --build
```

Admin UI:

```text
http://SERVER_IP:3000/admin
```

You can set or replace the GetToken upstream API key in the admin UI. It is saved to `./data/config.json` and takes effect immediately.

## Important Env Vars

```env
WRAPPER_API_KEY=key-for-newapi
ADMIN_TOKEN=admin-ui-token
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin-ui-password
BANANA_API_KEY=gettoken-enterprise-api-key
MAX_WORKERS=512
MAX_QUEUE=20000
```

For a 20-core, large-memory server, start with:

```env
MAX_WORKERS=512
MAX_QUEUE=20000
```

If GetToken upstream does not rate-limit you, raise gradually:

```env
MAX_WORKERS=1000
MAX_QUEUE=50000
```

Higher concurrency is not always faster. The real bottleneck is usually upstream task queueing and rate limits.

## Examples

Text-to-image:

```bash
curl -X POST http://127.0.0.1:3000/v1/images/generations \
  -H "Authorization: Bearer change-me-newapi-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "banana-pro-2k",
    "prompt": "brand ad poster, product display in a city night scene",
    "size": "16:9"
  }'
```

Image-to-image:

```bash
curl -X POST http://127.0.0.1:3000/v1/images/edits \
  -H "Authorization: Bearer change-me-newapi-key" \
  -F "model=banana-pro-2k" \
  -F "prompt=keep composition, make it commercial ad quality" \
  -F "size=16:9" \
  -F "image=@input.png"
```

JSON image-to-image:

```json
{
  "model": "banana-pro-2k",
  "prompt": "keep composition, make it commercial ad quality",
  "size": "16:9",
  "imageUrls": ["https://example.com/input.png"]
}
```

## Mapping

| OpenAI field | Banana field |
| --- | --- |
| `model=banana-pro-1k` | `resolution=1k` |
| `model=banana-pro-2k` | `resolution=2k` |
| `model=banana-pro-4k` | `resolution=4k` |
| `prompt` | `prompt` |
| `size=16:9` | `aspectRatio=16:9` |
| `size=1024x1024` | `aspectRatio=1:1` |
| multipart `image` | upload, then `imageUrls` |

Default response:

```json
{
  "created": 1780000000,
  "data": [{ "url": "https://..." }]
}
```

If request uses `response_format=b64_json`, the wrapper downloads the output image and returns base64.

# ZMO API 客户对接文档

本文档用于客户接入 ZMO API 中转站，调用图片生成、图片编辑、视频生成能力。

## 1. 接入信息

Base URL：

```text
https://api.zmoapi.cn/v1
```

认证方式：

```http
Authorization: Bearer 你的 API Key
```

示例：

```http
Authorization: Bearer sk-xxxxxxxx
```

## 2. 支持模型

### 图片模型

```text
banana-pro-1k
banana-pro-2k
banana-pro-4k

banana2-512
banana2-1k
banana2-2k
banana2-4k
```

说明：

- 图片模型不区分文生图和图生图。
- 请求里没有参考图时，自动按文生图处理。
- 请求里有 `imageUrls` 或上传图片时，自动按图生图处理。

### 视频模型

```text
veo31-pro-720p
veo31-pro-1080p
veo31-pro-4k

veo31-fast-720p
veo31-fast-1080p
veo31-fast-4k
```

说明：

- 视频模型不区分文生视频和图生视频。
- 请求里没有参考图时，自动按文生视频处理。
- 请求里有 `imageUrls` 时，自动按图生视频处理。
- `veo31-pro-*` 支持首尾帧视频。
- 暂不支持上传参考视频。

## 3. 查看可用模型

```bash
curl https://api.zmoapi.cn/v1/models \
  -H "Authorization: Bearer sk-xxxxxxxx"
```

返回中会包含当前账号可用的模型列表。

## 4. 图片生成

### 文生图

```bash
curl https://api.zmoapi.cn/v1/images/generations \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "banana-pro-2k",
    "prompt": "一张高级商业广告海报，城市夜景，电影级灯光",
    "size": "16:9",
    "n": 1
  }'
```

返回示例：

```json
{
  "created": 1780000000,
  "data": [
    {
      "url": "https://..."
    }
  ]
}
```

### 图生图，URL 方式

```bash
curl https://api.zmoapi.cn/v1/images/generations \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "banana2-2k",
    "prompt": "保持主体不变，改成黑白高级摄影风格",
    "imageUrls": [
      "https://example.com/input.png"
    ],
    "size": "1:1",
    "n": 1
  }'
```

### 图生图，上传文件方式

```bash
curl https://api.zmoapi.cn/v1/images/edits \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -F "model=banana-pro-2k" \
  -F "prompt=保持构图不变，改成商业海报质感" \
  -F "size=16:9" \
  -F "n=1" \
  -F "image=@input.png"
```

## 5. 视频生成

### 文生视频

```bash
curl https://api.zmoapi.cn/v1/videos/generations \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "veo31-fast-720p",
    "prompt": "黄昏海岸线上的未来城市，镜头缓慢推进",
    "aspectRatio": "16:9",
    "duration": "8"
  }'
```

### 图生视频

```bash
curl https://api.zmoapi.cn/v1/videos/generations \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "veo31-pro-1080p",
    "prompt": "根据参考图生成自然运镜，保持主体一致",
    "imageUrls": [
      "https://example.com/first-frame.png"
    ],
    "aspectRatio": "16:9",
    "duration": "8"
  }'
```

### 首尾帧视频，仅 Veo3.1 Pro

```bash
curl https://api.zmoapi.cn/v1/videos/generations \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "veo31-pro-1080p",
    "prompt": "从首帧自然过渡到尾帧，镜头平滑推进",
    "firstFrameUrl": "https://example.com/start.png",
    "lastFrameUrl": "https://example.com/end.png",
    "aspectRatio": "16:9",
    "duration": "8"
  }'
```

也可以用 `imageUrls` 传 2 张参考图，系统会自动按首尾帧视频处理：

```json
{
  "model": "veo31-pro-1080p",
  "prompt": "从第一张图自然过渡到第二张图",
  "imageUrls": [
    "https://example.com/start.png",
    "https://example.com/end.png"
  ],
  "aspectRatio": "16:9",
  "duration": "8"
}
```

## 6. 参数说明

### 通用参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | string | 是 | 模型名 |
| `prompt` | string | 是 | 提示词 |
| `n` | number | 否 | 生成数量，图片常用 `1` |
| `imageUrls` | string[] | 否 | 参考图 URL 数组 |

### 图片参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `size` | string | 否 | 图片比例，如 `1:1`、`16:9`、`9:16` |

图片参考图限制：

| 模型 | 参考图数量 |
| --- | --- |
| Banana Pro | 最多 10 张 |
| Banana2 | 最多 10 张 |

### 视频参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `aspectRatio` | string | 否 | 视频比例，支持 `16:9`、`9:16` |
| `duration` | string | 否 | 视频时长，目前传 `8` |
| `firstFrameUrl` | string | 否 | 首帧图片 URL，仅 Veo3.1 Pro 首尾帧视频使用 |
| `lastFrameUrl` | string | 否 | 尾帧图片 URL，仅 Veo3.1 Pro 首尾帧视频使用 |

视频参考图限制：

| 模型 | 参考图数量 |
| --- | --- |
| Veo3.1 Fast | 最多 1 张 |
| Veo3.1 Pro 图生视频 | 1 张 |
| Veo3.1 Pro 首尾帧视频 | 2 张 |

## 7. OpenAI SDK 示例

### JavaScript

```js
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "sk-xxxxxxxx",
  baseURL: "https://api.zmoapi.cn/v1",
});

const result = await client.images.generate({
  model: "banana-pro-2k",
  prompt: "一张高级商业广告海报，城市夜景，电影级灯光",
  size: "16:9",
  n: 1,
});

console.log(result.data[0].url);
```

### Python

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-xxxxxxxx",
    base_url="https://api.zmoapi.cn/v1",
)

result = client.images.generate(
    model="banana-pro-2k",
    prompt="一张高级商业广告海报，城市夜景，电影级灯光",
    size="16:9",
    n=1,
)

print(result.data[0].url)
```

## 8. 常见问题

### 图片模型是否要区分文生图和图生图？

不需要。客户只需要选择画质模型，例如 `banana-pro-2k` 或 `banana2-2k`。

- 没有参考图：文生图
- 有参考图：图生图

### 视频模型是否要区分文生视频和图生视频？

不需要。客户只需要选择分辨率模型，例如 `veo31-pro-1080p`。

- 没有参考图：文生视频
- 有 1 张参考图：图生视频
- Veo3.1 Pro 有 2 张参考图：首尾帧视频

### 是否支持参考视频？

暂不支持。当前支持的是参考图片、首帧图、尾帧图。

### 返回的是 URL 还是 base64？

默认返回 URL：

```json
{
  "data": [
    {
      "url": "https://..."
    }
  ]
}
```

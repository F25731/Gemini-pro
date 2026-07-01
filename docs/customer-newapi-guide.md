# ZMO API 客户图片对接文档

本文档用于客户接入 ZMO API 中转站，调用图片生成和图片编辑能力。

## 1. 接入信息

Base URL:

```text
https://api.zmoapi.cn/v1
```

认证方式:

```http
Authorization: Bearer 你的 API Key
```

## 2. 支持模型

```text
banana-pro-1k
banana-pro-2k
banana-pro-4k

banana2-512
banana2-1k
banana2-2k
banana2-4k
```

说明:

- 图片模型不区分文生图和图生图。
- 请求里没有参考图时，自动按文生图处理。
- 请求里有 `image_url`、`imageUrls` 或上传图片时，自动按图生图处理。

## 3. 查看可用模型

```bash
curl https://api.zmoapi.cn/v1/models \
  -H "Authorization: Bearer sk-xxxxxxxx"
```

## 4. 文生图

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

成功返回:

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

## 5. 图生图，URL 方式

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

也可以使用单图字段:

```json
{
  "model": "banana-pro-1k",
  "prompt": "改成商业产品海报",
  "image_url": "https://example.com/input.png"
}
```

## 6. 图生图，上传文件方式

```bash
curl https://api.zmoapi.cn/v1/images/edits \
  -H "Authorization: Bearer sk-xxxxxxxx" \
  -F "model=banana-pro-2k" \
  -F "prompt=保持构图不变，改成商业海报质感" \
  -F "size=16:9" \
  -F "n=1" \
  -F "image=@input.png"
```

## 7. 参数说明

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | string | 是 | 模型名 |
| `prompt` | string | 是 | 提示词 |
| `n` | number | 否 | 生成数量，默认 1，最大 8 |
| `size` | string | 否 | 图片比例，如 `1:1`、`16:9`、`9:16` |
| `image_url` | string | 否 | 单张参考图 URL |
| `imageUrls` | string[] | 否 | 多张参考图 URL |
| `response_format` | string | 否 | 传 `b64_json` 时返回 base64 |

参考图限制:

| 模型 | 参考图数量 | 单图大小 |
| --- | --- | --- |
| Banana Pro | 最多 10 张 | 10 MB |
| Banana2 | 最多 10 张 | 30 MB |

## 8. OpenAI SDK 示例

JavaScript:

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

Python:

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

## 9. 错误返回

失败时返回 OpenAI 兼容错误:

```json
{
  "error": {
    "message": "上游失败原因",
    "type": "invalid_request_error",
    "code": "upstream_failed"
  }
}
```

只有拿到最终图片 URL 或 `b64_json` 才算成功；只拿到上游任务 ID 不算成功。

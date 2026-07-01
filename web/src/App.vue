<template>
  <div class="shell" :class="{ collapsed }">
    <aside class="sidebar">
      <div class="brand">
        <button class="icon-btn" @click="collapsed = !collapsed">{{ collapsed ? ">" : "<" }}</button>
        <div v-if="!collapsed">
          <strong>Banana Wrapper</strong>
          <span>NewAPI 中转管理</span>
        </div>
      </div>
      <nav>
        <button v-for="item in navItems" :key="item.key" :class="{ active: view === item.key }" @click="view = item.key">
          <span>{{ item.icon }}</span>
          <em v-if="!collapsed">{{ item.label }}</em>
        </button>
      </nav>
    </aside>

    <main>
      <header class="topbar">
        <div>
          <p>{{ currentNav?.label || "监控板" }}</p>
          <h1>{{ currentTitle }}</h1>
        </div>
        <form class="login" @submit.prevent="login">
          <input v-model="username" placeholder="后台用户名" />
          <input v-model="password" type="password" placeholder="后台密码" />
          <button>登录</button>
        </form>
      </header>

      <p v-if="error" class="notice error">{{ error }}</p>
      <p v-if="success" class="notice success">{{ success }}</p>

      <section v-if="view === 'dashboard'" class="view">
        <div class="cards">
          <div class="metric"><span>平均延迟</span><strong>{{ formatMs(totalMetrics.avgLatencyMs) }}</strong></div>
          <div class="metric"><span>成功率</span><strong>{{ formatRate(totalMetrics.successRate) }}</strong></div>
          <div class="metric"><span>运行中</span><strong>{{ status?.pool?.active || 0 }}</strong></div>
          <div class="metric"><span>队列中</span><strong>{{ status?.pool?.queued || 0 }}</strong></div>
          <div class="metric"><span>Workers</span><strong>{{ status?.pool?.workers || 0 }}</strong></div>
        </div>
        <div class="panel">
          <h2>分组状态</h2>
          <div class="table">
            <div class="row head"><span>分组</span><span>成功率</span><span>平均延迟</span><span>成功/失败</span></div>
            <div v-if="metricRows.length === 0" class="empty">暂无任务数据，产生请求后会显示延迟和成功率。</div>
            <div v-for="row in metricRows" :key="row.key" class="row">
              <span>{{ row.label }}</span>
              <span>{{ formatRate(row.data.successRate) }}</span>
              <span>{{ formatMs(row.data.avgLatencyMs) }}</span>
              <span>{{ row.data.success }} / {{ row.data.failed }}</span>
            </div>
          </div>
        </div>
      </section>

      <section v-if="view === 'config'" class="view grid">
        <div class="panel">
          <h2>NewAPI 配置</h2>
          <dl>
            <dt>Base URL</dt>
            <dd><code>{{ baseUrl }}</code></dd>
            <dt>Wrapper API Key</dt>
            <dd>
              <span class="secret-state" :class="{ set: config?.wrapperApiKeySet }">
                <i></i>{{ config?.wrapperApiKeySet ? "已配置" : "未配置 WRAPPER_API_KEY" }}
              </span>
            </dd>
            <dt>模型</dt>
            <dd>{{ config?.models?.join(", ") }}</dd>
          </dl>
        </div>
        <div class="panel">
          <h2>上游 GetToken</h2>
          <dl>
            <dt>上游地址</dt>
            <dd><code>{{ config?.bananaBaseUrl }}</code></dd>
            <dt>API Key 状态</dt>
            <dd>
              <span class="secret-state" :class="{ set: config?.bananaApiKeySet }">
                <i></i>{{ config?.bananaApiKeySet ? `已配置 ${config?.bananaApiKeyHint || ""}` : "未配置" }}
              </span>
            </dd>
          </dl>
          <form class="inline-form" @submit.prevent="saveBananaKey">
            <input v-model="bananaApiKey" type="password" placeholder="粘贴上游 API Key" />
            <button>保存</button>
          </form>
        </div>
        <div class="panel wide">
          <h2>后端参数</h2>
          <div class="cards compact">
            <div class="metric"><span>Workers</span><strong>{{ config?.maxWorkers || 0 }}</strong></div>
            <div class="metric"><span>队列容量</span><strong>{{ config?.maxQueue || 0 }}</strong></div>
            <div class="metric"><span>轮询间隔</span><strong>{{ config?.pollIntervalMs || 0 }} ms</strong></div>
            <div class="metric"><span>请求超时</span><strong>{{ config?.requestTimeoutSec || 0 }} 秒</strong></div>
          </div>
        </div>
      </section>

      <section v-if="view === 'capabilities'" class="view">
        <div class="panel">
          <h2>模型能力总览</h2>
          <div class="capability-groups">
            <section class="capability-section">
              <h3>图片模型</h3>
              <div class="cap-table">
                <div class="cap-row head">
                  <span>模型</span><span>支持能力</span><span>参考图</span><span>画质/分辨率</span><span>比例/时长</span><span>说明</span>
                </div>
                <div v-if="imageModels.length === 0" class="empty">暂无模型数据。</div>
                <div v-for="model in imageModels" :key="model.id" class="cap-row">
                  <span><strong>{{ model.id }}</strong><small>{{ model.name }}</small></span>
                  <span>{{ model.capabilities?.join(" / ") || "-" }}</span>
                  <span>{{ refsText(model) }}<small v-if="model.maxFileSizeMb">单图 {{ model.maxFileSizeMb }}MB</small></span>
                  <span>{{ model.qualityTiers?.join(" / ") || model.resolution }}</span>
                  <span>
                    {{ model.aspectRatios?.join(" / ") || "-" }}
                    <small v-if="model.durationOptions?.length">时长 {{ model.durationOptions.join(" / ") }} 秒</small>
                  </span>
                  <span>{{ model.notes?.join("；") || "-" }}</span>
                </div>
              </div>
            </section>
            <section class="capability-section">
              <h3>视频模型</h3>
              <div class="cap-table">
                <div class="cap-row head">
                  <span>模型</span><span>支持能力</span><span>参考图</span><span>画质/分辨率</span><span>比例/时长</span><span>说明</span>
                </div>
                <div v-if="videoModels.length === 0" class="empty">暂无模型数据。</div>
                <div v-for="model in videoModels" :key="model.id" class="cap-row">
                  <span><strong>{{ model.id }}</strong><small>{{ model.name }}</small></span>
                  <span>{{ model.capabilities?.join(" / ") || "-" }}</span>
                  <span>{{ refsText(model) }}<small v-if="model.maxFileSizeMb">单图 {{ model.maxFileSizeMb }}MB</small></span>
                  <span>{{ model.qualityTiers?.join(" / ") || model.resolution }}</span>
                  <span>
                    {{ model.aspectRatios?.join(" / ") || "-" }}
                    <small v-if="model.durationOptions?.length">时长 {{ model.durationOptions.join(" / ") }} 秒</small>
                  </span>
                  <span>{{ model.notes?.join("；") || "-" }}</span>
                </div>
              </div>
            </section>
          </div>
        </div>
      </section>

      <section v-if="view === 'images'" class="view">
        <div class="cards">
          <div class="metric"><span>图片成功率</span><strong>{{ formatRate(imageMetrics.successRate) }}</strong></div>
          <div class="metric"><span>图片平均延迟</span><strong>{{ formatMs(imageMetrics.avgLatencyMs) }}</strong></div>
          <div class="metric"><span>成功</span><strong>{{ imageMetrics.success || 0 }}</strong></div>
          <div class="metric"><span>失败</span><strong>{{ imageMetrics.failed || 0 }}</strong></div>
        </div>
        <div class="panel">
          <h2>图片模型</h2>
          <div v-if="imageModels.length === 0" class="empty">暂无模型数据。</div>
          <div class="model-grid">
            <article v-for="model in imageModels" :key="model.id">
              <strong>{{ model.id }}</strong>
              <span>{{ model.name }}</span>
              <code>{{ model.textEndpoint }}</code>
              <code v-if="model.imageEndpoint">{{ model.imageEndpoint }}</code>
            </article>
          </div>
        </div>
      </section>

      <section v-if="view === 'videos'" class="view">
        <div class="cards">
          <div class="metric"><span>视频成功率</span><strong>{{ formatRate(videoMetrics.successRate) }}</strong></div>
          <div class="metric"><span>视频平均延迟</span><strong>{{ formatMs(videoMetrics.avgLatencyMs) }}</strong></div>
          <div class="metric"><span>成功</span><strong>{{ videoMetrics.success || 0 }}</strong></div>
          <div class="metric"><span>失败</span><strong>{{ videoMetrics.failed || 0 }}</strong></div>
        </div>
        <div class="panel">
          <h2>视频模型</h2>
          <div v-if="videoModels.length === 0" class="empty">暂无模型数据。</div>
          <div class="model-grid">
            <article v-for="model in videoModels" :key="model.id">
              <strong>{{ model.id }}</strong>
              <span>{{ model.name }}</span>
              <code>{{ model.textEndpoint }}</code>
              <code v-if="model.imageEndpoint">{{ model.imageEndpoint }}</code>
              <code v-if="model.startEndEndpoint">{{ model.startEndEndpoint }}</code>
            </article>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

const collapsed = ref(false);
const view = ref("dashboard");
const username = ref("");
const password = ref("");
const bananaApiKey = ref("");
const status = ref(null);
const config = ref(null);
const error = ref("");
const success = ref("");
let timer = null;

const navItems = [
  { key: "dashboard", label: "监控板", icon: "M" },
  { key: "config", label: "API 配置", icon: "A" },
  { key: "capabilities", label: "模型能力", icon: "C" },
  { key: "images", label: "图片监控", icon: "I" },
  { key: "videos", label: "视频监控", icon: "V" },
];

const currentNav = computed(() => navItems.find((item) => item.key === view.value));
const currentTitle = computed(() => ({
  dashboard: "实时运行概览",
  config: "接入与上游配置",
  capabilities: "模型能力说明",
  images: "图片任务监控",
  videos: "视频任务监控",
}[view.value] || "监控板"));
const baseUrl = computed(() => config.value?.publicBaseUrl || `${location.origin}/v1`);
const modelSpecs = computed(() => config.value?.modelSpecs || []);
const imageModels = computed(() => modelSpecs.value.filter((item) => item.media === "image"));
const videoModels = computed(() => modelSpecs.value.filter((item) => item.media === "video"));
const imageMetrics = computed(() => status.value?.metrics?.image || {});
const videoMetrics = computed(() => status.value?.metrics?.video || {});
const totalMetrics = computed(() => {
  const image = imageMetrics.value;
  const video = videoMetrics.value;
  const total = Number(image.total || 0) + Number(video.total || 0);
  const ok = Number(image.success || 0) + Number(video.success || 0);
  const latencyTotal = Number(image.avgLatencyMs || 0) * Number(image.total || 0) + Number(video.avgLatencyMs || 0) * Number(video.total || 0);
  return {
    successRate: total ? ok / total : 0,
    avgLatencyMs: total ? Math.round(latencyTotal / total) : 0,
  };
});
const metricRows = computed(() => {
  const labels = {
    image: "图片总览",
    video: "视频总览",
    "banana-pro": "Banana Pro",
    banana2: "Banana2",
    "veo31-pro": "Veo3.1 Pro",
    "veo31-fast": "Veo3.1 Fast",
  };
  return Object.entries(status.value?.metrics || {}).map(([key, data]) => ({ key, label: labels[key] || key, data }));
});

async function login() {
  try {
    const response = await fetch("/api/admin/login", {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: username.value, password: password.value }),
    });
    await readJson(response);
    error.value = "";
    success.value = "登录成功";
    await refresh();
  } catch (err) {
    success.value = "";
    error.value = err.message || "登录失败";
  }
}

async function saveBananaKey() {
  try {
    const response = await fetch("/api/admin/config", {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ bananaApiKey: bananaApiKey.value }),
    });
    await readJson(response);
    bananaApiKey.value = "";
    error.value = "";
    success.value = "上游 API Key 已保存";
    await refresh();
  } catch (err) {
    success.value = "";
    error.value = err.message || "保存失败";
  }
}

async function refresh() {
  try {
    const [nextStatus, nextConfig] = await Promise.all([
      fetch("/api/admin/status", { credentials: "same-origin" }).then(readJson),
      fetch("/api/admin/config", { credentials: "same-origin" }).then(readJson),
    ]);
    status.value = nextStatus;
    config.value = nextConfig;
    if (error.value === "invalid admin token") error.value = "";
  } catch (err) {
    error.value = err.message || "加载失败";
  }
}

function formatRate(value) {
  return `${Math.round(Number(value || 0) * 1000) / 10}%`;
}

function formatMs(value) {
  const ms = Number(value || 0);
  if (ms >= 1000) return `${Math.round(ms / 100) / 10}s`;
  return `${ms}ms`;
}

function refsText(model) {
  if (!model.maxImageInputs) return "0 张";
  if (model.minImageInputs && model.minImageInputs !== model.maxImageInputs) {
    return `${model.minImageInputs}-${model.maxImageInputs} 张`;
  }
  return `最多 ${model.maxImageInputs} 张`;
}

async function readJson(response) {
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.message || data.error?.message || `HTTP ${response.status}`);
  return data;
}

onMounted(() => {
  refresh();
  timer = setInterval(refresh, 3000);
});

onBeforeUnmount(() => {
  clearInterval(timer);
});
</script>

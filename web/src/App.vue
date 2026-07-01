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
          <MetricCard label="平均延迟" :value="formatMs(totalMetrics.avgLatencyMs)" />
          <MetricCard label="成功率" :value="formatRate(totalMetrics.successRate)" />
          <MetricCard label="运行中" :value="String(status?.pool?.active || 0)" />
          <MetricCard label="队列中" :value="String(status?.pool?.queued || 0)" />
          <MetricCard label="Workers" :value="String(status?.pool?.workers || 0)" />
        </div>
        <div class="panel">
          <h2>分组状态</h2>
          <div class="table">
            <div class="row head"><span>分组</span><span>成功率</span><span>平均延迟</span><span>成功/失败</span></div>
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
            <dd><code>{{ config?.wrapperApiKey || "未配置 WRAPPER_API_KEY" }}</code></dd>
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
            <dd>{{ config?.bananaApiKeySet ? config?.bananaApiKeyHint : "未配置" }}</dd>
          </dl>
          <form class="inline-form" @submit.prevent="saveBananaKey">
            <input v-model="bananaApiKey" type="password" placeholder="粘贴上游 API Key" />
            <button>保存</button>
          </form>
        </div>
        <div class="panel wide">
          <h2>后端参数</h2>
          <div class="cards compact">
            <MetricCard label="Workers" :value="String(config?.maxWorkers || 0)" />
            <MetricCard label="队列容量" :value="String(config?.maxQueue || 0)" />
            <MetricCard label="轮询间隔" :value="`${config?.pollIntervalMs || 0} ms`" />
            <MetricCard label="请求超时" :value="`${config?.requestTimeoutSec || 0} 秒`" />
          </div>
        </div>
      </section>

      <section v-if="view === 'images'" class="view">
        <div class="cards">
          <MetricCard label="图片成功率" :value="formatRate(imageMetrics.successRate)" />
          <MetricCard label="图片平均延迟" :value="formatMs(imageMetrics.avgLatencyMs)" />
          <MetricCard label="成功" :value="String(imageMetrics.success || 0)" />
          <MetricCard label="失败" :value="String(imageMetrics.failed || 0)" />
        </div>
        <ModelPanel title="图片模型" :models="imageModels" />
      </section>

      <section v-if="view === 'videos'" class="view">
        <div class="cards">
          <MetricCard label="视频成功率" :value="formatRate(videoMetrics.successRate)" />
          <MetricCard label="视频平均延迟" :value="formatMs(videoMetrics.avgLatencyMs)" />
          <MetricCard label="成功" :value="String(videoMetrics.success || 0)" />
          <MetricCard label="失败" :value="String(videoMetrics.failed || 0)" />
        </div>
        <ModelPanel title="视频模型" :models="videoModels" />
      </section>
    </main>
  </div>
</template>

<script>
const MetricCard = {
  props: ["label", "value"],
  template: `<div class="metric"><span>{{ label }}</span><strong>{{ value }}</strong></div>`,
};

const ModelPanel = {
  props: ["title", "models"],
  template: `
    <div class="panel">
      <h2>{{ title }}</h2>
      <div class="model-grid">
        <article v-for="model in models" :key="model.id">
          <strong>{{ model.id }}</strong>
          <span>{{ model.name }}</span>
          <code>{{ model.textEndpoint }}</code>
          <code v-if="model.imageEndpoint">{{ model.imageEndpoint }}</code>
          <code v-if="model.startEndEndpoint">{{ model.startEndEndpoint }}</code>
        </article>
      </div>
    </div>
  `,
};

export default {
  components: { MetricCard, ModelPanel },
  data() {
    return {
      collapsed: false,
      view: "dashboard",
      username: "",
      password: "",
      bananaApiKey: "",
      status: null,
      config: null,
      error: "",
      success: "",
      timer: null,
      navItems: [
        { key: "dashboard", label: "监控板", icon: "M" },
        { key: "config", label: "API 配置", icon: "A" },
        { key: "images", label: "图片监控", icon: "I" },
        { key: "videos", label: "视频监控", icon: "V" },
      ],
    };
  },
  computed: {
    currentNav() {
      return this.navItems.find((item) => item.key === this.view);
    },
    currentTitle() {
      const titleMap = {
        dashboard: "实时运行概览",
        config: "接入与上游配置",
        images: "图片任务监控",
        videos: "视频任务监控",
      };
      return titleMap[this.view] || "监控板";
    },
    baseUrl() {
      return this.config?.publicBaseUrl || `${location.origin}/v1`;
    },
    modelSpecs() {
      return this.config?.modelSpecs || [];
    },
    imageModels() {
      return this.modelSpecs.filter((item) => item.media === "image");
    },
    videoModels() {
      return this.modelSpecs.filter((item) => item.media === "video");
    },
    imageMetrics() {
      return this.status?.metrics?.image || {};
    },
    videoMetrics() {
      return this.status?.metrics?.video || {};
    },
    totalMetrics() {
      const image = this.imageMetrics;
      const video = this.videoMetrics;
      const total = Number(image.total || 0) + Number(video.total || 0);
      const success = Number(image.success || 0) + Number(video.success || 0);
      const latencyTotal = Number(image.avgLatencyMs || 0) * Number(image.total || 0) + Number(video.avgLatencyMs || 0) * Number(video.total || 0);
      return {
        successRate: total ? success / total : 0,
        avgLatencyMs: total ? Math.round(latencyTotal / total) : 0,
      };
    },
    metricRows() {
      const labels = {
        image: "图片总览",
        video: "视频总览",
        "banana-pro": "Banana Pro",
        banana2: "Banana2",
        "veo31-pro": "Veo3.1 Pro",
        "veo31-fast": "Veo3.1 Fast",
      };
      return Object.entries(this.status?.metrics || {}).map(([key, data]) => ({ key, label: labels[key] || key, data }));
    },
  },
  mounted() {
    this.refresh();
    this.timer = setInterval(this.refresh, 3000);
  },
  beforeUnmount() {
    clearInterval(this.timer);
  },
  methods: {
    async login() {
      try {
        const response = await fetch("/api/admin/login", {
          method: "POST",
          credentials: "same-origin",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ username: this.username, password: this.password }),
        });
        await readJson(response);
        this.error = "";
        this.success = "登录成功";
        await this.refresh();
      } catch (error) {
        this.success = "";
        this.error = error.message || "登录失败";
      }
    },
    async saveBananaKey() {
      try {
        const response = await fetch("/api/admin/config", {
          method: "POST",
          credentials: "same-origin",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ bananaApiKey: this.bananaApiKey }),
        });
        await readJson(response);
        this.bananaApiKey = "";
        this.error = "";
        this.success = "上游 API Key 已保存";
        await this.refresh();
      } catch (error) {
        this.success = "";
        this.error = error.message || "保存失败";
      }
    },
    async refresh() {
      try {
        const [status, config] = await Promise.all([
          fetch("/api/admin/status", { credentials: "same-origin" }).then(readJson),
          fetch("/api/admin/config", { credentials: "same-origin" }).then(readJson),
        ]);
        this.status = status;
        this.config = config;
        if (this.error === "invalid admin token") this.error = "";
      } catch (error) {
        this.error = error.message || "加载失败";
      }
    },
    formatRate(value) {
      return `${Math.round(Number(value || 0) * 1000) / 10}%`;
    },
    formatMs(value) {
      const ms = Number(value || 0);
      if (ms >= 1000) return `${Math.round(ms / 100) / 10}s`;
      return `${ms}ms`;
    },
  },
};

async function readJson(response) {
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.message || data.error?.message || `HTTP ${response.status}`);
  return data;
}
</script>

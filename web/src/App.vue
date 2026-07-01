<template>
  <div class="shell" :class="{ collapsed }">
    <aside class="sidebar">
      <div class="brand">
        <button class="icon-btn" @click="collapsed = !collapsed">{{ collapsed ? ">" : "<" }}</button>
        <div v-if="!collapsed">
          <strong>Banana Wrapper</strong>
          <span>NewAPI relay admin</span>
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
          <p>{{ currentNav?.label || "Dashboard" }}</p>
          <h1>{{ currentTitle }}</h1>
        </div>
        <form class="login" @submit.prevent="login">
          <input v-model="username" placeholder="Admin user" />
          <input v-model="password" type="password" placeholder="Admin password" />
          <button>Login</button>
        </form>
      </header>

      <p v-if="error" class="notice error">{{ error }}</p>
      <p v-if="success" class="notice success">{{ success }}</p>

      <section v-if="view === 'dashboard'" class="view">
        <div class="cards">
          <div class="metric"><span>Avg latency</span><strong>{{ formatMs(totalMetrics.avgLatencyMs) }}</strong></div>
          <div class="metric"><span>Success rate</span><strong>{{ formatRate(totalMetrics.successRate) }}</strong></div>
          <div class="metric"><span>Active</span><strong>{{ status?.pool?.active || 0 }}</strong></div>
          <div class="metric"><span>Queued</span><strong>{{ status?.pool?.queued || 0 }}</strong></div>
          <div class="metric"><span>Workers</span><strong>{{ status?.pool?.workers || 0 }}</strong></div>
        </div>
        <div class="panel">
          <h2>Group metrics</h2>
          <div class="table">
            <div class="row head"><span>Group</span><span>Success rate</span><span>Avg latency</span><span>OK / Failed</span></div>
            <div v-if="metricRows.length === 0" class="empty">No traffic yet. Metrics will appear after requests.</div>
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
          <h2>NewAPI config</h2>
          <dl>
            <dt>Base URL</dt>
            <dd><code>{{ baseUrl }}</code></dd>
            <dt>Wrapper API Key</dt>
            <dd>
              <div v-if="config?.wrapperApiKey" class="key-line">
                <code>{{ config.wrapperApiKey }}</code>
                <button type="button" @click="copyText(config.wrapperApiKey)">Copy</button>
              </div>
              <span v-else class="secret-state"><i></i>Not configured</span>
            </dd>
            <dt>Models</dt>
            <dd>{{ config?.models?.join(", ") }}</dd>
          </dl>
        </div>
        <div class="panel">
          <h2>Upstream GetToken</h2>
          <dl>
            <dt>Base URL</dt>
            <dd><code>{{ config?.bananaBaseUrl }}</code></dd>
            <dt>API Key status</dt>
            <dd>
              <span class="secret-state" :class="{ set: config?.bananaApiKeySet }">
                <i></i>{{ config?.bananaApiKeySet ? `Configured ${config?.bananaApiKeyHint || ""}` : "Not configured" }}
              </span>
            </dd>
          </dl>
          <form class="inline-form" @submit.prevent="saveBananaKey">
            <input v-model="bananaApiKey" type="password" placeholder="Paste upstream API key" />
            <button>Save</button>
          </form>
        </div>
        <div class="panel wide">
          <h2>Runtime</h2>
          <div class="cards compact">
            <div class="metric"><span>Workers</span><strong>{{ config?.maxWorkers || 0 }}</strong></div>
            <div class="metric"><span>Queue capacity</span><strong>{{ config?.maxQueue || 0 }}</strong></div>
            <div class="metric"><span>Poll interval</span><strong>{{ config?.pollIntervalMs || 0 }} ms</strong></div>
            <div class="metric"><span>Heartbeat</span><strong>{{ config?.heartbeatSec || 0 }} s</strong></div>
            <div class="metric"><span>Timeout</span><strong>{{ config?.requestTimeoutSec || 0 }} s</strong></div>
          </div>
        </div>
      </section>

      <section v-if="view === 'capabilities'" class="view">
        <div class="panel">
          <h2>Model capabilities</h2>
          <div class="capability-groups">
            <CapabilityTable title="Image models" :models="imageModels" />
            <CapabilityTable title="Video models" :models="videoModels" />
          </div>
        </div>
      </section>

      <section v-if="view === 'images'" class="view">
        <div class="cards">
          <div class="metric"><span>Image success rate</span><strong>{{ formatRate(imageMetrics.successRate) }}</strong></div>
          <div class="metric"><span>Image avg latency</span><strong>{{ formatMs(imageMetrics.avgLatencyMs) }}</strong></div>
          <div class="metric"><span>Success</span><strong>{{ imageMetrics.success || 0 }}</strong></div>
          <div class="metric"><span>Failed</span><strong>{{ imageMetrics.failed || 0 }}</strong></div>
        </div>
        <ModelGrid title="Image models" :models="imageModels" />
      </section>

      <section v-if="view === 'videos'" class="view">
        <div class="cards">
          <div class="metric"><span>Video success rate</span><strong>{{ formatRate(videoMetrics.successRate) }}</strong></div>
          <div class="metric"><span>Video avg latency</span><strong>{{ formatMs(videoMetrics.avgLatencyMs) }}</strong></div>
          <div class="metric"><span>Success</span><strong>{{ videoMetrics.success || 0 }}</strong></div>
          <div class="metric"><span>Failed</span><strong>{{ videoMetrics.failed || 0 }}</strong></div>
        </div>
        <ModelGrid title="Video models" :models="videoModels" />
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onBeforeUnmount, onMounted, ref } from "vue";

const CapabilityTable = defineComponent({
  props: { title: String, models: { type: Array, default: () => [] } },
  setup(props) {
    return () => h("section", { class: "capability-section" }, [
      h("h3", props.title),
      h("div", { class: "cap-table" }, [
        h("div", { class: "cap-row head" }, ["Model", "Features", "References", "Quality", "Ratio / duration", "Notes"].map((text) => h("span", text))),
        props.models.length === 0 ? h("div", { class: "empty" }, "No model data.") : null,
        ...props.models.map((model) => h("div", { class: "cap-row", key: model.id }, [
          h("span", [h("strong", model.id), h("small", model.name)]),
          h("span", model.capabilities?.join(" / ") || "-"),
          h("span", [refsText(model), model.maxFileSizeMb ? h("small", `per image ${model.maxFileSizeMb}MB`) : null]),
          h("span", model.qualityTiers?.join(" / ") || model.resolution),
          h("span", [model.aspectRatios?.join(" / ") || "-", model.durationOptions?.length ? h("small", `duration ${model.durationOptions.join(" / ")}s`) : null]),
          h("span", model.notes?.join("; ") || "-"),
        ])),
      ]),
    ]);
  },
});

const ModelGrid = defineComponent({
  props: { title: String, models: { type: Array, default: () => [] } },
  setup(props) {
    return () => h("div", { class: "panel" }, [
      h("h2", props.title),
      props.models.length === 0 ? h("div", { class: "empty" }, "No model data.") : null,
      h("div", { class: "model-grid" }, props.models.map((model) => h("article", { key: model.id }, [
        h("strong", model.id),
        h("span", model.name),
        h("code", model.textEndpoint),
        model.imageEndpoint ? h("code", model.imageEndpoint) : null,
        model.startEndEndpoint ? h("code", model.startEndEndpoint) : null,
      ]))),
    ]);
  },
});

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
  { key: "dashboard", label: "Dashboard", icon: "M" },
  { key: "config", label: "API config", icon: "A" },
  { key: "capabilities", label: "Capabilities", icon: "C" },
  { key: "images", label: "Images", icon: "I" },
  { key: "videos", label: "Videos", icon: "V" },
];

const currentNav = computed(() => navItems.find((item) => item.key === view.value));
const currentTitle = computed(() => ({
  dashboard: "Runtime overview",
  config: "NewAPI and upstream config",
  capabilities: "Model capability matrix",
  images: "Image monitoring",
  videos: "Video monitoring",
}[view.value] || "Dashboard"));
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
  return { successRate: total ? ok / total : 0, avgLatencyMs: total ? Math.round(latencyTotal / total) : 0 };
});
const metricRows = computed(() => {
  const labels = { image: "Images", video: "Videos", "banana-pro": "Banana Pro", banana2: "Banana2", "veo3.1-pro": "Veo3.1 Pro", "veo3.1-fast": "Veo3.1 Fast" };
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
    success.value = "Logged in";
    await refresh();
  } catch (err) {
    success.value = "";
    error.value = err.message || "Login failed";
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
    success.value = "Upstream API key saved";
    await refresh();
  } catch (err) {
    success.value = "";
    error.value = err.message || "Save failed";
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
  } catch (err) {
    error.value = err.message || "Load failed";
  }
}

async function copyText(value) {
  try {
    await navigator.clipboard.writeText(value);
    success.value = "Copied";
  } catch {
    success.value = "Copy failed, select the key manually";
  }
}

function refsText(model) {
  if (!model.maxImageInputs) return "0";
  if (model.minImageInputs && model.minImageInputs !== model.maxImageInputs) return `${model.minImageInputs}-${model.maxImageInputs}`;
  return `up to ${model.maxImageInputs}`;
}

function formatRate(value) {
  return `${Math.round(Number(value || 0) * 1000) / 10}%`;
}

function formatMs(value) {
  const ms = Number(value || 0);
  if (ms >= 1000) return `${Math.round(ms / 100) / 10}s`;
  return `${ms}ms`;
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

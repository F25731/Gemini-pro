<template>
  <main class="page">
    <section class="hero">
      <div>
        <p class="eyebrow">Banana Pro Wrapper</p>
        <h1>OpenAI 兼容中转管理台</h1>
        <p class="sub">一个模型三档清晰度，自动按请求类型转发文生图和图生图。</p>
      </div>
      <div class="token">
        <label>管理登录</label>
        <div>
          <input v-model="username" placeholder="用户名" @keyup.enter="login" />
          <input v-model="password" type="password" placeholder="密码" @keyup.enter="login" />
          <button @click="login">登录</button>
        </div>
        <details>
          <summary>Token 方式</summary>
          <div>
            <input v-model="token" type="password" placeholder="ADMIN_TOKEN" @keyup.enter="saveToken" />
            <button @click="saveToken">保存</button>
          </div>
        </details>
      </div>
    </section>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="success" class="success">{{ success }}</p>

    <section v-if="status && config" class="grid">
      <div class="panel">
        <h2>运行状态</h2>
        <div class="stats">
          <div><strong>{{ status.pool.workers }}</strong><span>Workers</span></div>
          <div><strong>{{ status.pool.active }}</strong><span>运行中</span></div>
          <div><strong>{{ status.pool.queued }}</strong><span>队列中</span></div>
          <div><strong>{{ status.pool.completed }}</strong><span>成功</span></div>
          <div><strong>{{ status.pool.failed }}</strong><span>失败</span></div>
        </div>
      </div>

      <div class="panel">
        <h2>模型</h2>
        <div class="models">
          <code v-for="model in config.models" :key="model">{{ model }}</code>
        </div>
      </div>

      <div class="panel">
        <h2>NewAPI 配置</h2>
        <dl>
          <dt>Base URL</dt>
          <dd><code>{{ baseUrl }}</code></dd>
          <dt>API Key</dt>
          <dd><code>{{ config.wrapperApiKey || "未配置 WRAPPER_API_KEY" }}</code></dd>
          <dt>模型</dt>
          <dd>{{ config.models.join(", ") }}</dd>
          <dt>能力</dt>
          <dd>/images/generations 文生图，/images/edits 图生图</dd>
        </dl>
      </div>

      <div class="panel">
        <h2>上游 GetToken</h2>
        <dl>
          <dt>API Key 状态</dt>
          <dd>{{ config.bananaApiKeySet ? config.bananaApiKeyHint : "未配置" }}</dd>
        </dl>
        <div class="save-key">
          <input v-model="bananaApiKey" type="password" placeholder="粘贴 GetToken 企业级 API Key" @keyup.enter="saveBananaKey" />
          <button @click="saveBananaKey">保存上游 Key</button>
        </div>
      </div>

      <div class="panel">
        <h2>后端配置</h2>
        <dl>
          <dt>GetToken Base</dt>
          <dd><code>{{ config.bananaBaseUrl }}</code></dd>
          <dt>队列容量</dt>
          <dd>{{ config.maxQueue }}</dd>
          <dt>轮询间隔</dt>
          <dd>{{ config.pollIntervalMs }} ms</dd>
          <dt>请求超时</dt>
          <dd>{{ config.requestTimeoutSec }} 秒</dd>
        </dl>
      </div>
    </section>
  </main>
</template>

<script>
export default {
  data() {
    return {
      token: localStorage.getItem("banana-admin-token") || "",
      username: "",
      password: "",
      bananaApiKey: "",
      status: null,
      config: null,
      error: "",
      success: "",
      timer: null,
    };
  },
  computed: {
    authHeaders() {
      return this.token ? { Authorization: `Bearer ${this.token}` } : {};
    },
    baseUrl() {
      return this.config?.publicBaseUrl || `${location.origin}/v1`;
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
    saveToken() {
      localStorage.setItem("banana-admin-token", this.token);
      this.success = "管理 Token 已保存";
      this.refresh();
    },
    async login() {
      try {
        const response = await fetch("/api/admin/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ username: this.username, password: this.password }),
        });
        await readJson(response);
        this.error = "";
        this.success = "登录成功";
        this.refresh();
      } catch (error) {
        this.success = "";
        this.error = error.message || "登录失败";
      }
    },
    async saveBananaKey() {
      try {
        const response = await fetch("/api/admin/config", {
          method: "POST",
          headers: { ...this.authHeaders, "Content-Type": "application/json" },
          body: JSON.stringify({ bananaApiKey: this.bananaApiKey }),
        });
        await readJson(response);
        this.bananaApiKey = "";
        this.error = "";
        this.success = "上游 GetToken API Key 配置成功";
        await this.refresh();
      } catch (error) {
        this.success = "";
        this.error = error.message || "保存失败";
      }
    },
    async refresh() {
      try {
        const [status, config] = await Promise.all([
          fetch("/api/admin/status", { headers: this.authHeaders }).then(readJson),
          fetch("/api/admin/config", { headers: this.authHeaders }).then(readJson),
        ]);
        this.status = status;
        this.config = config;
        this.error = "";
      } catch (error) {
        this.error = error.message || "加载失败";
      }
    },
  },
};

async function readJson(response) {
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.message || data.error?.message || `HTTP ${response.status}`);
  return data;
}
</script>

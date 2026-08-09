<template>
  <div v-if="!authed" class="min-h-screen flex items-center justify-center">
    <div class="w-full max-w-md bg-white dark:bg-[#0b0f19] border border-slate-200 dark:border-slate-800 rounded-2xl p-8 shadow-2xl">
      <div class="text-center mb-8">
        <div class="w-16 h-16 mx-auto bg-indigo-500/20 rounded-full flex items-center justify-center border border-indigo-500/30 mb-4">
          <svg class="w-8 h-8 text-indigo-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="11" width="18" height="11" rx="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" />
          </svg>
        </div>
        <h1 class="text-xl font-bold">Sing-Box 流量监控面板</h1>
        <p class="text-sm text-slate-600 dark:text-slate-400 mt-1">请输入管理员密码</p>
      </div>
      <form @submit.prevent="doLogin" class="space-y-4">
        <input v-model="pwd" type="password" required placeholder="••••••••"
          class="w-full px-4 py-3 bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-indigo-500/50" />
        <p v-if="error" class="text-sm text-red-400 text-center">{{ error }}</p>
        <button :disabled="loading" class="w-full py-3 bg-indigo-600 hover:bg-indigo-500 rounded-xl font-semibold disabled:opacity-50">
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>
    </div>
  </div>

  <div v-else class="min-h-screen">
    <!-- 顶部导航 -->
    <header class="sticky top-0 z-30 bg-white dark:bg-[#131b2e]/80 backdrop-blur-md border-b border-slate-200 dark:border-slate-800/80">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center space-x-3">
            <h1 class="text-lg font-bold">📊 Sing-Box Dashboard</h1>
            <span v-if="ver" class="text-[10px] bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 px-2 py-0.5 rounded-full">v{{ ver }}</span>
          </div>
          <nav class="hidden md:flex items-center space-x-1 bg-slate-50/80 dark:bg-slate-100 dark:bg-slate-900/60 p-1 rounded-xl border border-slate-200 dark:border-slate-800">
            <button v-for="t in tabs" :key="t.id" @click="tab = t.id"
              :class="[tab === t.id ? 'bg-indigo-600 text-white' : 'text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:text-slate-200', 'px-4 py-2 text-sm font-medium rounded-lg transition-all']">
              {{ t.name }}
            </button>
          </nav>
          <div class="flex items-center space-x-2">
            <button @click="toggleTheme" title="切换主题"
              class="inline-flex items-center justify-center w-8 h-8 text-sm bg-slate-200 dark:bg-slate-800 hover:bg-slate-300 dark:hover:bg-slate-700 rounded-lg transition-all">
              {{ isDark ? '☀️' : '🌙' }}
            </button>
            <button @click="doImport" :disabled="loading" title="从 /etc/sing-box/config.json 同步导入"
              class="hidden sm:inline-flex px-3 py-1.5 text-xs text-slate-700 dark:text-slate-300 bg-slate-200 dark:bg-slate-800 hover:bg-slate-300 dark:hover:bg-slate-700 border border-slate-300 dark:border-slate-700 rounded-lg disabled:opacity-50">
              同步母配置
            </button>
            <button @click="doReload" :disabled="loading"
              class="inline-flex px-3.5 py-1.5 text-xs font-semibold text-white bg-gradient-to-r from-indigo-600 to-cyan-600 rounded-lg disabled:opacity-50">
              {{ loading ? '执行中...' : '一键重载配置' }}
            </button>
          </div>
        </div>
      </div>
    </header>

    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
      <Overview v-if="tab === 'overview'" :users="users" :inbounds="inbounds" :stats="stats" @switch-tab="tab = $event" @refresh="refresh" />
      <UsersView v-else-if="tab === 'users'" :users="users" :inbounds="inbounds" @changed="refresh" @refresh="refresh" />
      <NodesView v-else-if="tab === 'nodes'" :inbounds="inbounds" @changed="refresh" @refresh="refresh" />
      <StatsView v-else-if="tab === 'stats'" :stats="stats" @refresh="refresh" />
      <RealtimeView v-else-if="tab === 'realtime'" />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { api, getToken, setToken, clearToken } from './api'
import Overview from './components/Overview.vue'
import UsersView from './components/UsersView.vue'
import NodesView from './components/NodesView.vue'
import StatsView from './components/StatsView.vue'
import RealtimeView from './components/RealtimeView.vue'

const tabs = [
  { id: 'overview', name: '概览大盘' },
  { id: 'users', name: '用户管理' },
  { id: 'nodes', name: '入站节点' },
  { id: 'stats', name: '历史统计' },
  { id: 'realtime', name: '实时监控' },
]

const authed = ref(!!getToken())
const pwd = ref('')
const error = ref('')
const loading = ref(false)
const tab = ref('overview')

// 主题：默认夜间，localStorage 记忆
const isDark = ref(localStorage.getItem('theme') !== 'light')
function applyTheme() {
  document.documentElement.classList.toggle('dark', isDark.value)
}
function toggleTheme() {
  isDark.value = !isDark.value
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  applyTheme()
}

const users = ref([])
const inbounds = ref([])
const stats = ref([])
const ver = ref('')

async function fetchVersion() {
  try {
    const r = await api.version()
    ver.value = r.version || ''
  } catch {}
}

async function doLogin() {
  loading.value = true
  error.value = ''
  try {
    const r = await api.login(pwd.value)
    setToken(r.token)
    authed.value = true
    await fetchVersion()
    await refresh()
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function refresh() {
  try {
    const [u, i, s] = await Promise.all([api.users(), api.inbounds(), api.stats()])
    users.value = u
    inbounds.value = i
    stats.value = s
  } catch (e) {
    // token 失效时由 api.js 派发 auth-failed
  }
}

async function doReload() {
  loading.value = true
  try {
    await api.reload()
    alert('配置已重新生成并重载 sing-box')
  } catch (e) {
    alert('重载失败：' + e.message)
  } finally {
    loading.value = false
  }
}

async function doImport() {
  loading.value = true
  try {
    await api.importCfg()
    alert('已从母配置同步导入')
    await refresh()
  } catch (e) {
    alert('导入失败：' + e.message)
  } finally {
    loading.value = false
  }
}

function onAuthFailed() {
  clearToken()
  authed.value = false
  tab.value = 'overview'
}

onMounted(() => {
  window.addEventListener('auth-failed', onAuthFailed)
  applyTheme()
  fetchVersion()
  if (authed.value) refresh()
})
onBeforeUnmount(() => window.removeEventListener('auth-failed', onAuthFailed))
</script>

// API 封装（对齐后端 /api/v1）
const TOKEN_KEY = 'admin_token'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(t) {
  localStorage.setItem(TOKEN_KEY, t)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

async function req(path, options = {}) {
  const token = getToken()
  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) }
  if (token) headers.Authorization = `Bearer ${token}`

  // 相对路径：支持 nginx 子路径反代（/control/ → 后端）
  const resp = await fetch(`api/v1${path}`, { ...options, headers })
  if (resp.status === 401) {
    clearToken()
    window.dispatchEvent(new Event('auth-failed'))
    throw new Error('认证失败或已过期，请重新登录')
  }
  const data = await resp.json().catch(() => ({}))
  if (!resp.ok) throw new Error(data.error || data.message || `HTTP ${resp.status}`)
  return data
}

export const api = {
  login: (password) => req('/auth/login', { method: 'POST', body: JSON.stringify({ password }) }),
  stats: () => req('/stats'),
  statsUsers: () => req('/stats/users'),
  users: () => req('/users'),
  createUser: (u) => req('/users', { method: 'POST', body: JSON.stringify(u) }),
  updateUser: (id, u) => req(`/users/detail?id=${id}`, { method: 'PUT', body: JSON.stringify(u) }),
  deleteUser: (id) => req(`/users/detail?id=${id}`, { method: 'DELETE' }),
  setCycle: (user_id, cycle_start, cycle_days) =>
    req('/users/cycle', { method: 'PUT', body: JSON.stringify({ user_id, cycle_start, cycle_days }) }),
  inbounds: () => req('/inbounds'),
  createInbound: (n) => req('/inbounds', { method: 'POST', body: JSON.stringify(n) }),
  updateInbound: (id, n) => req(`/inbounds/detail?id=${id}`, { method: 'PUT', body: JSON.stringify(n) }),
  deleteInbound: (id) => req(`/inbounds/detail?id=${id}`, { method: 'DELETE' }),
  reload: () => req('/system/reload', { method: 'POST' }),
  importCfg: () => req('/system/import', { method: 'POST' }),
}

// 字节格式化
export function fmtBytes(b) {
  if (!b || b <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.min(units.length - 1, Math.floor(Math.log(b) / Math.log(1024)))
  return `${(b / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 2)} ${units[i]}`
}

export function fmtTime(t) {
  if (!t) return '永久有效'
  return new Date(t).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit',
  })
}

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 p-4 rounded-2xl">
      <div class="relative flex-1 max-w-md">
        <input v-model="kw" type="text" placeholder="搜索用户名、Email 或 UUID..."
          class="w-full pl-4 pr-4 py-2 text-xs bg-slate-100 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl focus:outline-none focus:border-indigo-500" />
      </div>
      <button @click="openCreate()" class="inline-flex px-4 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 rounded-xl">
        + 添加新代理用户
      </button>
      <button @click="emit('refresh')" title="刷新数据"
        class="inline-flex px-3 py-2 text-xs text-slate-600 dark:text-slate-300 bg-slate-200 dark:bg-slate-800 hover:bg-slate-300 dark:hover:bg-slate-700 rounded-xl transition-all">
        🔄 刷新
      </button>
    </div>

    <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-x-auto">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-100/80 dark:bg-slate-100 dark:bg-slate-900/80 text-slate-600 dark:text-slate-400 uppercase font-semibold border-b border-slate-200 dark:border-slate-800">
          <tr>
            <SortableTh label="用户" :active="sortKey === 'email'" :dir="sortDir" @sort="toggleSort('email')" />
            <SortableTh label="周期流量 / 周期窗口" :active="sortKey === 'period_total_bytes'" :dir="sortDir" @sort="toggleSort('period_total_bytes')" />
            <SortableTh label="已用 / 限额" :active="sortKey === 'used_traffic'" :dir="sortDir" @sort="toggleSort('used_traffic')" />
            <SortableTh label="到期时间" :active="sortKey === 'expire_at'" :dir="sortDir" @sort="toggleSort('expire_at')" />
            <SortableTh label="状态" :active="sortKey === 'status'" :dir="sortDir" @sort="toggleSort('status')" />
            <th class="py-3.5 px-4 text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60">
          <tr v-for="u in sortedUsers" :key="u.id" class="hover:bg-slate-100 dark:hover:bg-slate-200 dark:hover:bg-slate-200 dark:bg-slate-800/40 transition-colors">
            <td class="py-3.5 px-4">
              <div class="font-bold text-sm">{{ u.email }}</div>
              <div class="text-[11px] text-slate-500 font-mono">UUID: {{ u.uuid ? u.uuid.slice(0, 8) + '...' : '无' }}</div>
            </td>
            <td class="py-3.5 px-4">
              <div class="font-bold text-indigo-300">{{ fmtBytes(u.period_total_bytes) }}</div>
              <div class="text-[11px] text-slate-500">↑{{ fmtBytes(u.period_up_bytes) }} ↓{{ fmtBytes(u.period_down_bytes) }}</div>
              <button @click="openCycle(u)" class="text-[11px] text-cyan-400 hover:underline mt-0.5">
                周期: {{ u.cycle_days || 30 }}天 {{ fmtDate(u.cycle_start) }} 起 →
              </button>
            </td>
            <td class="py-3.5 px-4">
              <div class="font-mono font-bold">{{ fmtBytes(u.used_traffic) }} / {{ u.traffic_limit > 0 ? fmtBytes(u.traffic_limit) : '无上限' }}</div>
              <div v-if="u.traffic_limit > 0" class="w-full h-1.5 bg-slate-200 dark:bg-slate-800 rounded-full overflow-hidden mt-1">
                <div :class="u.is_over_limit ? 'bg-red-500' : u.used_traffic / u.traffic_limit > 0.8 ? 'bg-amber-400' : 'bg-indigo-500'"
                  :style="{ width: Math.min(100, u.used_traffic / u.traffic_limit * 100) + '%' }" class="h-full rounded-full"></div>
              </div>
            </td>
            <td class="py-3.5 px-4 text-slate-600 dark:text-slate-400">{{ fmtTime(u.expire_at) }}</td>
            <td class="py-3.5 px-4">
              <span :class="[u.enable ? (u.is_over_limit ? 'bg-red-500/10 text-red-400 border-red-500/20' : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20') : 'bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-300 dark:border-slate-700',
                'inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold border']">
                {{ u.enable ? (u.is_over_limit ? '流量超额' : '正常启用') : '已禁用' }}
              </span>
            </td>
            <td class="py-3.5 px-4 text-right space-x-2">
              <button @click="openEdit(u)" class="p-1.5 text-slate-600 dark:text-slate-400 hover:text-indigo-400 hover:bg-slate-200 dark:hover:bg-slate-200 dark:bg-slate-800 rounded-lg" title="编辑">✏️</button>
              <button @click="remove(u)" class="p-1.5 text-slate-600 dark:text-slate-400 hover:text-red-400 hover:bg-slate-200 dark:hover:bg-slate-200 dark:bg-slate-800 rounded-lg" title="删除">🗑️</button>
            </td>
          </tr>
          <tr v-if="filtered.length === 0"><td colspan="6" class="py-12 text-center text-slate-500">没有找到用户</td></tr>
        </tbody>
      </table>
    </div>

    <!-- 新建/编辑弹窗 -->
    <div v-if="modalOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
      <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800">
          <h3 class="text-lg font-bold">{{ editing ? '编辑用户' : '新建代理用户' }}</h3>
          <button @click="modalOpen = false" class="text-slate-600 dark:text-slate-400 hover:text-white">✕</button>
        </div>
        <form @submit.prevent="save" class="mt-4 space-y-4">
          <div>
            <label class="block text-xs font-semibold mb-1">用户标识 (Email / Name) *</label>
            <input v-model="form.email" required type="text" placeholder="user1@example.com 或 张三"
              class="w-full px-3.5 py-2 text-sm bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl focus:outline-none focus:border-indigo-500" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold mb-1">UUID (VLESS / VMess)</label>
              <div class="flex gap-1">
                <input v-model="form.uuid" type="text" placeholder="自动生成" class="w-full px-3.5 py-2 text-sm bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl font-mono" />
                <button type="button" @click="form.uuid = genUUID()" class="text-[11px] text-indigo-400 whitespace-nowrap">生成</button>
              </div>
            </div>
            <div>
              <label class="block text-xs font-semibold mb-1">密码 (Trojan / TUIC)</label>
              <input v-model="form.password" type="text" placeholder="pass123" class="w-full px-3.5 py-2 text-sm bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold mb-1">流量上限 (GB, 0=不限)</label>
              <input v-model.number="form.trafficLimitGB" type="number" min="0" step="1" class="w-full px-3.5 py-2 text-sm bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl" />
            </div>
            <div>
              <label class="block text-xs font-semibold mb-1">到期时间 (留空永久)</label>
              <input v-model="form.expireAtStr" type="datetime-local" class="w-full px-3.5 py-2 text-sm bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold mb-1">周期天数（默认 30）</label>
              <input v-model.number="form.cycleDays" type="number" min="1" step="1" class="w-full px-3.5 py-2 text-sm bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl" />
            </div>
            <div>
              <label class="block text-xs font-semibold mb-1">周期起始时间</label>
              <input v-model="form.cycleStartStr" type="datetime-local" class="w-full px-3.5 py-2 text-sm bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl" />
            </div>
          </div>
          <div>
            <label class="block text-xs font-semibold mb-1.5">绑定入站节点（不勾选默认全部）</label>
            <div class="max-h-36 overflow-y-auto bg-slate-100 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-2 space-y-1.5">
              <div v-for="n in inbounds" :key="n.id" @click="toggleInbound(n.id)"
                class="flex items-center space-x-2.5 text-xs text-slate-800 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-200 dark:hover:bg-slate-200 dark:bg-slate-800/50 p-1.5 rounded-lg cursor-pointer">
                <input type="checkbox" :checked="form.inbound_ids.includes(n.id)" class="rounded border-slate-300 dark:border-slate-700 bg-slate-200 dark:bg-slate-800 text-indigo-500" />
                <span class="font-medium">{{ n.tag }}</span>
                <span class="text-slate-500">({{ n.type }} :{{ n.listen_port }})</span>
              </div>
              <div v-if="inbounds.length === 0" class="text-xs text-slate-500 py-1 text-center">暂无入站节点</div>
            </div>
          </div>
          <label class="flex items-center space-x-2 text-xs">
            <input v-model="form.enable" type="checkbox" class="rounded border-slate-300 dark:border-slate-700 bg-slate-200 dark:bg-slate-800 text-indigo-500" />
            <span>启用此账号</span>
          </label>
          <div class="flex justify-end space-x-3 pt-4 border-t border-slate-200 dark:border-slate-800">
            <button type="button" @click="modalOpen = false" class="px-4 py-2 text-xs text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-200 dark:bg-slate-800 rounded-xl">取消</button>
            <button type="submit" :disabled="submitting" class="px-5 py-2 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 rounded-xl disabled:opacity-50">
              {{ submitting ? '保存中...' : (editing ? '保存修改' : '确认创建') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- 周期设置弹窗 -->
    <div v-if="cycleOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
      <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl max-w-md w-full p-6 shadow-2xl">
        <div class="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800">
          <h3 class="text-lg font-bold">⚙️ 设置计费周期 — {{ cycleUser?.email }}</h3>
          <button @click="cycleOpen = false" class="text-slate-600 dark:text-slate-400 hover:text-white">✕</button>
        </div>
        <div class="mt-4 space-y-4">
          <p class="text-xs text-slate-600 dark:text-slate-400">周期窗口自动滚动：当前周期内流量 = 大盘显示的该用户流量。修改后立即生效。</p>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold mb-1">周期天数 *</label>
              <input v-model.number="cycleForm.days" type="number" min="1" required class="w-full px-3.5 py-2 text-sm bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl" />
            </div>
            <div>
              <label class="block text-xs font-semibold mb-1">起始时间 *</label>
              <input v-model="cycleForm.start" type="datetime-local" required class="w-full px-3.5 py-2 text-sm bg-slate-100 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-xl" />
            </div>
          </div>
          <div class="flex justify-end space-x-3 pt-4 border-t border-slate-200 dark:border-slate-800">
            <button @click="cycleOpen = false" class="px-4 py-2 text-xs text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-200 dark:bg-slate-800 rounded-xl">取消</button>
            <button @click="saveCycle" :disabled="submitting" class="px-5 py-2 text-xs font-semibold bg-cyan-600 hover:bg-cyan-500 rounded-xl disabled:opacity-50">
              {{ submitting ? '保存中...' : '保存周期' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive } from 'vue'
import { api, fmtBytes, fmtTime } from '../api'
import SortableTh from './SortableTh.vue'

const props = defineProps({
  users: { type: Array, default: () => [] },
  inbounds: { type: Array, default: () => [] },
})
const emit = defineEmits(['changed', 'refresh'])

const kw = ref('')
const filtered = computed(() => {
  const q = kw.value.trim().toLowerCase()
  if (!q) return props.users
  return props.users.filter(u =>
    (u.email && u.email.toLowerCase().includes(q)) || (u.uuid && u.uuid.toLowerCase().includes(q)))
})

// 表头排序
const sortKey = ref('period_total_bytes')
const sortDir = ref('desc')

function toggleSort(k) {
  if (sortKey.value === k) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = k
    sortDir.value = 'asc'
  }
}

function sortValue(u, k) {
  if (k === 'expire_at') return u.expire_at ? new Date(u.expire_at).getTime() : 0
  if (k === 'status') {
    // 启用且未超额 > 启用已超额 > 禁用
    if (u.enable && !u.is_over_limit) return 2
    if (u.enable) return 1
    return 0
  }
  return u[k]
}

const sortedUsers = computed(() => {
  const arr = [...filtered.value]
  const k = sortKey.value
  const dir = sortDir.value === 'asc' ? 1 : -1
  arr.sort((a, b) => {
    const va = sortValue(a, k)
    const vb = sortValue(b, k)
    if (typeof va === 'string' && typeof vb === 'string') return va.localeCompare(vb, 'zh') * dir
    return (va - vb) * dir
  })
  return arr
})

const modalOpen = ref(false)
const cycleOpen = ref(false)
const editing = ref(false)
const submitting = ref(false)
const current = ref(null)
const cycleUser = ref(null)
const cycleForm = reactive({ days: 30, start: '' })

const form = reactive({
  email: '', uuid: '', password: '', flow: 'xtls-rprx-vision',
  trafficLimitGB: 0, expireAtStr: '', enable: true, inbound_ids: [],
  cycleDays: 30, cycleStartStr: '',
})

function genUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
    const r = Math.random() * 16 | 0
    return (c === 'x' ? r : (r & 3 | 8)).toString(16)
  })
}

function openCreate() {
  editing.value = false
  current.value = null
  Object.assign(form, {
    email: '', uuid: genUUID(), password: '', flow: 'xtls-rprx-vision',
    trafficLimitGB: 0, expireAtStr: '', enable: true, inbound_ids: [],
    cycleDays: 30, cycleStartStr: '',
  })
  modalOpen.value = true
}

function openEdit(u) {
  editing.value = true
  current.value = u
  Object.assign(form, {
    email: u.email, uuid: u.uuid || '', password: u.password || '',
    flow: u.flow || '', trafficLimitGB: u.traffic_limit ? u.traffic_limit / (1024 ** 3) : 0,
    expireAtStr: u.expire_at ? new Date(u.expire_at).toISOString().slice(0, 16) : '',
    enable: u.enable !== false, inbound_ids: [...(u.inbound_ids || [])],
    cycleDays: u.cycle_days || 30,
    cycleStartStr: u.cycle_start ? new Date(u.cycle_start).toISOString().slice(0, 16) : '',
  })
  modalOpen.value = true
}

function openCycle(u) {
  cycleUser.value = u
  cycleForm.days = u.cycle_days || 30
  cycleForm.start = u.cycle_start ? new Date(u.cycle_start).toISOString().slice(0, 16) : new Date().toISOString().slice(0, 16)
  cycleOpen.value = true
}

function toggleInbound(id) {
  const i = form.inbound_ids.indexOf(id)
  i > -1 ? form.inbound_ids.splice(i, 1) : form.inbound_ids.push(id)
}

async function save() {
  submitting.value = true
  try {
    const payload = {
      email: form.email,
      uuid: form.uuid,
      password: form.password,
      flow: form.flow,
      enable: form.enable,
      traffic_limit: Math.floor(form.trafficLimitGB * 1024 ** 3),
      expire_at: form.expireAtStr ? new Date(form.expireAtStr).toISOString() : null,
      inbound_ids: form.inbound_ids,
      cycle_start: form.cycleStartStr ? new Date(form.cycleStartStr).toISOString() : null,
      cycle_days: form.cycleDays || 30,
    }
    if (editing.value) await api.updateUser(current.value.id, payload)
    else await api.createUser(payload)
    modalOpen.value = false
    emit('changed')
  } catch (e) {
    alert('保存失败：' + e.message)
  } finally {
    submitting.value = false
  }
}

async function saveCycle() {
  if (!cycleUser.value) return
  submitting.value = true
  try {
    await api.setCycle(cycleUser.value.id, new Date(cycleForm.start).toISOString(), cycleForm.days)
    cycleOpen.value = false
    emit('changed')
  } catch (e) {
    alert('保存失败：' + e.message)
  } finally {
    submitting.value = false
  }
}

async function remove(u) {
  if (!confirm(`确认删除用户 ${u.email}？`)) return
  try {
    await api.deleteUser(u.id)
    emit('changed')
  } catch (e) {
    alert('删除失败：' + e.message)
  }
}

function fmtDate(t) {
  return t ? new Date(t).toLocaleDateString('zh-CN') : '—'
}
</script>

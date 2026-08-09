<template>
  <div class="space-y-6">
    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl p-5">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase">系统上行流量</span>
          <span class="p-2.5 bg-indigo-500/10 rounded-xl text-indigo-400">↑</span>
        </div>
        <div class="mt-3 text-2xl font-bold">{{ fmtBytes(totalUp) }}</div>
      </div>
      <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl p-5">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase">系统下行流量</span>
          <span class="p-2.5 bg-cyan-500/10 rounded-xl text-cyan-400">↓</span>
        </div>
        <div class="mt-3 text-2xl font-bold">{{ fmtBytes(totalDown) }}</div>
      </div>
      <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl p-5">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase">代理用户数</span>
          <span class="p-2.5 bg-emerald-500/10 rounded-xl text-emerald-400">👤</span>
        </div>
        <div class="mt-3 text-2xl font-bold">{{ users.length }} <span class="text-xs text-slate-600 dark:text-slate-400">({{ activeCount }} 活跃)</span></div>
      </div>
      <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl p-5">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase">入站节点数</span>
          <span class="p-2.5 bg-amber-500/10 rounded-xl text-amber-400">🖧</span>
        </div>
        <div class="mt-3 text-2xl font-bold">{{ inbounds.length }} <span class="text-xs text-slate-600 dark:text-slate-400">节点服务</span></div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 周期流量 Top 榜（核心新功能） -->
      <div class="lg:col-span-2 bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl p-6">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h3 class="text-base font-bold">🔥 周期流量 Top 榜</h3>
            <p class="text-xs text-slate-600 dark:text-slate-400 mt-0.5">按当前计费周期内已用流量降序（周期 = 起始时间 + {{ defaultCycleDays }} 天）</p>
          </div>
          <button @click="$emit('switch-tab', 'users')" class="text-xs text-indigo-400 hover:text-indigo-300">查看全部用户 →</button>
        </div>
        <div class="space-y-3">
          <div v-for="(u, i) in topPeriodUsers" :key="u.id"
            class="p-3.5 bg-slate-50/80 dark:bg-slate-100 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-xl flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <span :class="rankCls(i)" class="w-7 h-7 rounded-lg border font-bold text-xs flex items-center justify-center">{{ i + 1 }}</span>
              <div>
                <div class="text-sm font-semibold flex items-center gap-2">
                  {{ u.email }}
                  <span v-if="u.is_over_limit" class="text-[10px] bg-red-500/10 text-red-400 border border-red-500/20 px-1.5 py-0.5 rounded">已超额</span>
                </div>
                <div class="text-xs text-slate-600 dark:text-slate-400 mt-0.5">周期：{{ fmtTime(u.cycle_start) }} ~ {{ fmtTime(u.cycle_end) }}</div>
              </div>
            </div>
            <div class="text-right">
              <div class="text-sm font-bold font-mono">{{ fmtBytes(u.period_total_bytes) }}</div>
              <div class="text-[11px] text-slate-600 dark:text-slate-400">↑{{ fmtBytes(u.period_up_bytes) }} ↓{{ fmtBytes(u.period_down_bytes) }}</div>
            </div>
          </div>
          <div v-if="topPeriodUsers.length === 0" class="py-8 text-center text-xs text-slate-500">暂无周期流量数据</div>
        </div>
      </div>

      <!-- 入站节点状态 -->
      <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-base font-bold">🖧 入站节点状态</h3>
          <div class="flex items-center space-x-2">
            <button @click="$emit('refresh')" title="刷新数据"
              class="px-3 py-1.5 text-xs text-slate-600 dark:text-slate-300 bg-slate-200 dark:bg-slate-800 hover:bg-slate-300 dark:hover:bg-slate-700 rounded-lg transition-all">
              🔄 刷新
            </button>
            <button @click="$emit('switch-tab', 'nodes')" class="text-xs text-cyan-400 hover:text-cyan-300">管理节点 →</button>
          </div>
        </div>
        <div class="space-y-3">
          <div v-for="n in inbounds" :key="n.id"
            class="p-3 bg-slate-50/80 dark:bg-slate-100 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-xl flex items-center justify-between">
            <div>
              <div class="text-sm font-semibold flex items-center gap-2">
                {{ n.tag }}
                <span :class="[n.type === 'vless' ? 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20' : 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20', 'text-[10px] uppercase border px-1.5 py-0.5 rounded']">{{ n.type }}</span>
              </div>
              <div class="text-xs text-slate-600 dark:text-slate-400 mt-0.5">端口 :{{ n.listen_port }} | {{ n.server_name || '内置路由' }}</div>
            </div>
            <span :class="n.enable ? 'bg-emerald-400 shadow-emerald-400/50' : 'bg-slate-600'" class="w-2.5 h-2.5 rounded-full shadow-md"></span>
          </div>
          <div v-if="inbounds.length === 0" class="py-8 text-center text-xs text-slate-500">暂无入站节点</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { fmtBytes, fmtTime } from '../api'

const props = defineProps({
  users: { type: Array, default: () => [] },
  inbounds: { type: Array, default: () => [] },
  stats: { type: Array, default: () => [] },
})
defineEmits(['switch-tab', 'refresh'])

const defaultCycleDays = 30

const totalUp = computed(() => props.stats.filter(s => s.category === 'user').reduce((a, s) => a + s.uplink_bytes, 0))
const totalDown = computed(() => props.stats.filter(s => s.category === 'user').reduce((a, s) => a + s.downlink_bytes, 0))
const activeCount = computed(() => props.users.filter(u => u.enable && !u.is_over_limit).length)
const topPeriodUsers = computed(() =>
  [...props.users].sort((a, b) => (b.period_total_bytes || 0) - (a.period_total_bytes || 0)).slice(0, 5)
)

function rankCls(i) {
  if (i === 0) return 'bg-amber-500/20 text-amber-400 border-amber-500/30'
  if (i === 1) return 'bg-slate-400/20 text-slate-700 dark:text-slate-300 border-slate-400/30'
  if (i === 2) return 'bg-amber-700/20 text-amber-600 border-amber-700/30'
  return 'bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-300 dark:border-slate-700'
}
</script>

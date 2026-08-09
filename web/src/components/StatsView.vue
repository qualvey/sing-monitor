<template>
  <div class="space-y-6">
    <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 p-4 rounded-2xl flex items-center justify-between">
      <div>
        <h2 class="text-base font-bold">📈 全局流量累计明细</h2>
        <p class="text-xs text-slate-600 dark:text-slate-400 mt-0.5">所有用户与节点的累计流量统计</p>
      </div>
      <button @click="emit('refresh')" class="inline-flex items-center px-3 py-1.5 text-xs text-slate-700 dark:text-slate-300 bg-slate-200 dark:bg-slate-800 hover:bg-slate-300 dark:hover:bg-slate-700 rounded-lg">🔄 刷新数据</button>
    </div>

    <div class="bg-white dark:bg-[#131b2e] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-x-auto">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-100/80 dark:bg-slate-100 dark:bg-slate-900/80 text-slate-600 dark:text-slate-400 uppercase font-semibold border-b border-slate-200 dark:border-slate-800">
          <tr>
            <SortableTh label="类型" :active="sortKey === 'category'" :dir="sortDir" @sort="toggleSort('category')" />
            <SortableTh label="目标" :active="sortKey === 'target_name'" :dir="sortDir" @sort="toggleSort('target_name')" />
            <SortableTh label="下行" :active="sortKey === 'downlink_bytes'" :dir="sortDir" @sort="toggleSort('downlink_bytes')" />
            <SortableTh label="上行" :active="sortKey === 'uplink_bytes'" :dir="sortDir" @sort="toggleSort('uplink_bytes')" />
            <SortableTh label="总量" :active="sortKey === 'total_bytes'" :dir="sortDir" @sort="toggleSort('total_bytes')" />
            <SortableTh label="最后更新" align="right" :active="sortKey === 'updated_at'" :dir="sortDir" @sort="toggleSort('updated_at')" />
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60">
          <tr v-for="s in sortedStats" :key="s.id" class="hover:bg-slate-100 dark:hover:bg-slate-200 dark:hover:bg-slate-200 dark:bg-slate-800/40 transition-colors">
            <td class="py-3.5 px-4">
              <span :class="[s.category === 'user' ? 'bg-indigo-500/10 text-indigo-400' : 'bg-cyan-500/10 text-cyan-400', 'inline-flex px-2 py-0.5 rounded text-[11px] font-semibold border uppercase']">{{ s.category }}</span>
            </td>
            <td class="py-3.5 px-4 font-bold text-sm">{{ s.target_name }}</td>
            <td class="py-3.5 px-4 font-mono text-cyan-400">{{ fmtBytes(s.downlink_bytes) }}</td>
            <td class="py-3.5 px-4 font-mono text-indigo-400">{{ fmtBytes(s.uplink_bytes) }}</td>
            <td class="py-3.5 px-4 font-mono font-bold text-white">{{ fmtBytes(s.total_bytes) }}</td>
            <td class="py-3.5 px-4 text-right text-slate-600 dark:text-slate-400 font-mono">{{ fmtTime(s.updated_at) }}</td>
          </tr>
          <tr v-if="stats.length === 0"><td colspan="6" class="py-12 text-center text-slate-500">暂无流量统计数据</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { fmtBytes, fmtTime } from '../api'
import SortableTh from './SortableTh.vue'

const props = defineProps({
  stats: { type: Array, default: () => [] },
})
const emit = defineEmits(['refresh'])

// 表头排序（默认总量降序）
const sortKey = ref('total_bytes')
const sortDir = ref('desc')

function toggleSort(k) {
  if (sortKey.value === k) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = k
    sortDir.value = 'asc'
  }
}

const sortedStats = computed(() => {
  const arr = [...props.stats]
  const k = sortKey.value
  const dir = sortDir.value === 'asc' ? 1 : -1
  arr.sort((a, b) => {
    let va = a[k]
    let vb = b[k]
    if (k === 'updated_at') {
      va = va ? new Date(va).getTime() : 0
      vb = vb ? new Date(vb).getTime() : 0
    }
    if (typeof va === 'string' && typeof vb === 'string') return va.localeCompare(vb, 'zh') * dir
    return (va - vb) * dir
  })
  return arr
})
</script>

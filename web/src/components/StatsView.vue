<template>
  <div class="space-y-6">
    <div class="bg-[#131b2e] border border-slate-800 p-4 rounded-2xl flex items-center justify-between">
      <div>
        <h2 class="text-base font-bold">📈 全局流量累计明细</h2>
        <p class="text-xs text-slate-400 mt-0.5">所有用户与节点的累计流量统计</p>
      </div>
      <button @click="emit('refresh')" class="inline-flex items-center px-3 py-1.5 text-xs text-slate-300 bg-slate-800 hover:bg-slate-700 rounded-lg">🔄 刷新数据</button>
    </div>

    <div class="bg-[#131b2e] border border-slate-800 rounded-2xl overflow-x-auto">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 text-slate-400 uppercase font-semibold border-b border-slate-800">
          <tr>
            <th class="py-3.5 px-4">类型</th>
            <th class="py-3.5 px-4">目标</th>
            <th class="py-3.5 px-4">下行</th>
            <th class="py-3.5 px-4">上行</th>
            <th class="py-3.5 px-4">总量</th>
            <th class="py-3.5 px-4 text-right">最后更新</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60">
          <tr v-for="s in stats" :key="s.id" class="hover:bg-slate-800/40 transition-colors">
            <td class="py-3.5 px-4">
              <span :class="[s.category === 'user' ? 'bg-indigo-500/10 text-indigo-400' : 'bg-cyan-500/10 text-cyan-400', 'inline-flex px-2 py-0.5 rounded text-[11px] font-semibold border uppercase']">{{ s.category }}</span>
            </td>
            <td class="py-3.5 px-4 font-bold text-sm">{{ s.target_name }}</td>
            <td class="py-3.5 px-4 font-mono text-cyan-400">{{ fmtBytes(s.downlink_bytes) }}</td>
            <td class="py-3.5 px-4 font-mono text-indigo-400">{{ fmtBytes(s.uplink_bytes) }}</td>
            <td class="py-3.5 px-4 font-mono font-bold text-white">{{ fmtBytes(s.total_bytes) }}</td>
            <td class="py-3.5 px-4 text-right text-slate-400 font-mono">{{ fmtTime(s.updated_at) }}</td>
          </tr>
          <tr v-if="stats.length === 0"><td colspan="6" class="py-12 text-center text-slate-500">暂无流量统计数据</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { fmtBytes, fmtTime } from '../api'

defineProps({
  stats: { type: Array, default: () => [] },
})
const emit = defineEmits(['refresh'])
</script>

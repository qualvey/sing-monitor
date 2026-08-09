<template>
  <div class="space-y-6">
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <!-- 实时速率图表 -->
      <div class="lg:col-span-2 bg-[#131b2e] border border-slate-800 rounded-2xl p-4">
        <h3 class="text-base font-bold mb-2">⚡ 实时速率 (bytes/s)</h3>
        <div ref="chartEl" class="h-64"></div>
      </div>
      <div class="bg-[#131b2e] border border-slate-800 rounded-2xl p-4 space-y-3">
        <div>
          <div class="text-xs text-slate-400 uppercase">上行速率</div>
          <div class="text-2xl font-bold text-indigo-300">{{ fmtBytes(globalUp) }}/s</div>
        </div>
        <div>
          <div class="text-xs text-slate-400 uppercase">下行速率</div>
          <div class="text-2xl font-bold text-cyan-300">{{ fmtBytes(globalDown) }}/s</div>
        </div>
        <div>
          <div class="text-xs text-slate-400 uppercase">连接状态</div>
          <div class="flex items-center gap-2 mt-1">
            <span :class="connected ? 'bg-emerald-400' : 'bg-red-500'" class="w-2.5 h-2.5 rounded-full animate-pulse"></span>
            <span class="text-sm">{{ connected ? '已连接 (实时推送)' : '连接断开，重连中...' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 用户实时状态表 -->
    <div class="bg-[#131b2e] border border-slate-800 rounded-2xl overflow-x-auto">
      <table class="w-full text-left text-xs">
        <thead class="bg-slate-900/80 text-slate-400 uppercase font-semibold border-b border-slate-800">
          <tr>
            <th class="py-3.5 px-4">用户</th>
            <th class="py-3.5 px-4">上行速率</th>
            <th class="py-3.5 px-4">下行速率</th>
            <th class="py-3.5 px-4">状态</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-800/60">
          <tr v-for="u in users" :key="u.name" class="hover:bg-slate-800/40">
            <td class="py-3 px-4 font-semibold">{{ u.name }}</td>
            <td class="py-3 px-4 font-mono text-indigo-400">{{ fmtBytes(u.uplink) }}/s</td>
            <td class="py-3 px-4 font-mono text-cyan-400">{{ fmtBytes(u.downlink) }}/s</td>
            <td class="py-3 px-4">
              <span :class="u.online ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-slate-800 text-slate-400 border-slate-700',
                'inline-flex px-2 py-0.5 rounded-full text-[11px] font-semibold border'">
                {{ u.online ? '在线' : '离线' }}
              </span>
            </td>
          </tr>
          <tr v-if="users.length === 0"><td colspan="4" class="py-12 text-center text-slate-500">等待实时数据...</td></tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'
import { getToken, fmtBytes } from '../api'

const chartEl = ref(null)
const connected = ref(false)
const globalUp = ref(0)
const globalDown = ref(0)
const users = ref([])

let ws = null
let chart = null
const upSeries = []
const downSeries = []

function connect() {
  // 相对路径 WebSocket：支持 nginx 子路径反代
  const url = new URL('api/v1/ws/rt', window.location.href)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.search = getToken() ? `?token=${getToken()}` : ''
  ws = new WebSocket(url.href)
  ws.onopen = () => { connected.value = true }
  ws.onmessage = (e) => {
    try {
      const d = JSON.parse(e.data)
      globalUp.value = d.global?.uplink || 0
      globalDown.value = d.global?.downlink || 0
      users.value = d.users || []
      pushChart(globalUp.value, globalDown.value)
    } catch {}
  }
  ws.onclose = () => {
    connected.value = false
    setTimeout(connect, 2000)
  }
  ws.onerror = () => ws && ws.close()
}

function pushChart(up, down) {
  // x 轴必须用时间戳（epoch ms），time 轴才能正确渲染
  const t = Date.now()
  upSeries.push([t, up])
  downSeries.push([t, down])
  if (upSeries.length > 60) { upSeries.shift(); downSeries.shift() }
  if (chart) {
    chart.setOption({
      series: [
        { name: '上行', data: upSeries },
        { name: '下行', data: downSeries },
      ],
    })
  }
}

// 速率友好格式化（bytes/s → KB/MB/GB per s）
function fmtRate(v) {
  if (v >= 1e9) return (v / 1e9).toFixed(2) + ' GB/s'
  if (v >= 1e6) return (v / 1e6).toFixed(2) + ' MB/s'
  if (v >= 1e3) return (v / 1e3).toFixed(1) + ' KB/s'
  return v + ' B/s'
}

onMounted(async () => {
  await nextTick()
  chart = echarts.init(chartEl.value)
  chart.setOption({
    tooltip: {
      trigger: 'axis',
      valueFormatter: fmtRate,
    },
    legend: { data: ['上行', '下行'], textStyle: { color: '#94a3b8' } },
    grid: { left: 70, right: 20, top: 30, bottom: 30 },
    xAxis: { type: 'time', axisLabel: { color: '#64748b' } },
    yAxis: {
      type: 'value',
      // 不从 0 开始，放大小流量波动；轴标签格式化单位
      scale: true,
      axisLabel: { color: '#64748b', formatter: fmtRate },
    },
    // 滚轮/拖拽缩放
    dataZoom: [{ type: 'inside' }, { type: 'slider', height: 14, bottom: 5 }],
    series: [
      { name: '上行', type: 'line', smooth: true, showSymbol: false, data: [], lineStyle: { color: '#818cf8' }, itemStyle: { color: '#818cf8' }, areaStyle: { opacity: 0.08 } },
      { name: '下行', type: 'line', smooth: true, showSymbol: false, data: [], lineStyle: { color: '#22d3ee' }, itemStyle: { color: '#22d3ee' }, areaStyle: { opacity: 0.08 } },
    ],
  })
  connect()
})

onBeforeUnmount(() => {
  if (ws) ws.close()
  if (chart) chart.dispose()
})
</script>

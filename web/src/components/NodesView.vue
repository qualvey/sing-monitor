<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between bg-[#131b2e] border border-slate-800 p-4 rounded-2xl">
      <div>
        <h2 class="text-base font-bold">🖧 入站节点管理</h2>
        <p class="text-xs text-slate-400 mt-0.5">支持 VLESS + Reality 与 TUIC v5 入站节点管理</p>
      </div>
      <button @click="openCreate()" class="inline-flex px-4 py-2 text-xs font-semibold bg-cyan-600 hover:bg-cyan-500 rounded-xl">
        + 添加新入站节点
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div v-for="n in inbounds" :key="n.id"
        class="bg-[#131b2e] border border-slate-800 rounded-2xl p-5 shadow-xl">
        <div class="flex items-start justify-between">
          <div class="flex items-center space-x-3">
            <span :class="[n.type === 'vless' ? 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20' : 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20', 'p-2.5 rounded-xl border']">
              {{ n.type === 'vless' ? '🔒' : '🚀' }}
            </span>
            <div>
              <h3 class="text-base font-bold flex items-center gap-2">
                {{ n.tag }}
                <span :class="[n.type === 'vless' ? 'bg-indigo-500/10 text-indigo-400' : 'bg-cyan-500/10 text-cyan-400', 'text-[10px] uppercase border px-2 py-0.5 rounded-full']">
                  {{ n.type === 'vless' ? 'VLESS + Reality' : 'TUIC v5' }}
                </span>
              </h3>
              <p class="text-xs text-slate-400 mt-0.5">监听: {{ n.listen || '::' }} : {{ n.listen_port }}</p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <button @click="openEdit(n)" class="p-1.5 text-slate-400 hover:text-cyan-400 hover:bg-slate-800 rounded-lg" title="编辑">✏️</button>
            <button @click="remove(n)" class="p-1.5 text-slate-400 hover:text-red-400 hover:bg-slate-800 rounded-lg" title="删除">🗑️</button>
          </div>
        </div>

        <div class="mt-4 pt-3 border-t border-slate-800/80 space-y-2 text-xs">
          <template v-if="n.type === 'vless'">
            <div class="flex justify-between"><span class="text-slate-500">SNI:</span><span class="font-mono">{{ n.server_name || '未指定' }}</span></div>
            <div class="flex justify-between"><span class="text-slate-500">Handshake:</span><span class="font-mono">{{ n.handshake_server || '未指定' }}:{{ n.handshake_port || 443 }}</span></div>
            <div class="flex justify-between"><span class="text-slate-500">Short ID:</span><span class="font-mono">{{ n.short_id || '未配置' }}</span></div>
          </template>
          <template v-else>
            <div class="flex justify-between"><span class="text-slate-500">Congestion:</span><span class="font-mono uppercase">{{ n.congestion_control || 'BBR' }}</span></div>
            <div class="flex justify-between"><span class="text-slate-500">Auth Timeout:</span><span class="font-mono">{{ n.auth_timeout || '3s' }}</span></div>
            <div class="flex justify-between"><span class="text-slate-500">ALPN:</span><span class="font-mono">{{ n.alpn || 'h3' }}</span></div>
            <div class="flex justify-between"><span class="text-slate-500">0-RTT:</span><span class="text-emerald-400 font-semibold">{{ n.zero_rtt_handshake ? '已开启' : '未开启' }}</span></div>
          </template>
        </div>

        <div class="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs">
          <span :class="n.enable ? 'text-emerald-400' : 'text-slate-500'">{{ n.enable ? '运行正常' : '已禁用' }}</span>
          <span class="text-emerald-400 font-medium">监听中 :{{ n.listen_port }}</span>
        </div>
      </div>
      <div v-if="inbounds.length === 0" class="col-span-2 py-16 bg-[#131b2e] border border-slate-800 rounded-2xl text-center text-slate-500 text-xs">
        当前无入站节点
      </div>
    </div>

    <!-- 新建/编辑弹窗 -->
    <div v-if="modalOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
      <div class="bg-[#131b2e] border border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between pb-4 border-b border-slate-800">
          <h3 class="text-lg font-bold">{{ editing ? '编辑入站节点' : '新建入站节点' }}</h3>
          <button @click="modalOpen = false" class="text-slate-400 hover:text-white">✕</button>
        </div>
        <form @submit.prevent="save" class="mt-4 space-y-4">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold mb-1">节点 Tag *</label>
              <input v-model="form.tag" required type="text" placeholder="vless-reality-8443" class="w-full px-3.5 py-2 text-sm bg-slate-900 border border-slate-700 rounded-xl" />
            </div>
            <div>
              <label class="block text-xs font-semibold mb-1">协议类型 *</label>
              <select v-model="form.type" :disabled="editing" class="w-full px-3 py-2 text-sm bg-slate-900 border border-slate-700 rounded-xl">
                <option value="vless">VLESS + Reality</option>
                <option value="tuic">TUIC v5</option>
              </select>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold mb-1">监听端口 *</label>
              <input v-model.number="form.listen_port" required type="number" min="1" max="65535" placeholder="8443" class="w-full px-3.5 py-2 text-sm bg-slate-900 border border-slate-700 rounded-xl" />
            </div>
            <div>
              <label class="block text-xs font-semibold mb-1">监听地址</label>
              <input v-model="form.listen" type="text" placeholder="::" class="w-full px-3.5 py-2 text-sm bg-slate-900 border border-slate-700 rounded-xl" />
            </div>
          </div>

          <div v-if="form.type === 'vless'" class="space-y-3 p-3.5 bg-slate-900/80 border border-slate-800 rounded-xl">
            <div class="text-xs font-bold text-indigo-400 uppercase">VLESS + Reality 专属设置</div>
            <div>
              <label class="block text-xs font-semibold mb-1">SNI 伪装域名</label>
              <input v-model="form.server_name" type="text" placeholder="www.amazon.com" class="w-full px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-lg" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-semibold mb-1">PrivateKey</label>
                <div class="flex gap-1">
                  <input v-model="form.private_key" type="text" placeholder="留空自动生成" class="w-full px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-lg font-mono" />
                  <button type="button" @click="form.private_key = genKey()" class="text-[11px] text-cyan-400 whitespace-nowrap">生成</button>
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold mb-1">Short ID</label>
                <input v-model="form.short_id" type="text" placeholder="0123456789abcdef" class="w-full px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-lg font-mono" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-semibold mb-1">Handshake Server</label>
                <input v-model="form.handshake_server" type="text" placeholder="www.amazon.com" class="w-full px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-lg" />
              </div>
              <div>
                <label class="block text-xs font-semibold mb-1">Handshake Port</label>
                <input v-model.number="form.handshake_port" type="number" placeholder="443" class="w-full px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-lg" />
              </div>
            </div>
          </div>

          <div v-else class="space-y-3 p-3.5 bg-slate-900/80 border border-slate-800 rounded-xl">
            <div class="text-xs font-bold text-cyan-400 uppercase">TUIC v5 专属设置</div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-semibold mb-1">拥塞控制</label>
                <select v-model="form.congestion_control" class="w-full px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-lg">
                  <option value="bbr">bbr (推荐)</option><option value="cubic">cubic</option><option value="new_reno">new_reno</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-semibold mb-1">Auth Timeout</label>
                <input v-model="form.auth_timeout" type="text" placeholder="3s" class="w-full px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-lg" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-semibold mb-1">Certificate Provider</label>
                <input v-model="form.certificate_provider" type="text" placeholder="my-acme" class="w-full px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-lg" />
              </div>
              <div>
                <label class="block text-xs font-semibold mb-1">ALPN</label>
                <input v-model="form.alpn" type="text" placeholder="h3" class="w-full px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-lg" />
              </div>
            </div>
            <label class="flex items-center space-x-2 text-xs">
              <input v-model="form.zero_rtt_handshake" type="checkbox" class="rounded border-slate-800 bg-slate-950 text-cyan-500" />
              <span>开启 0-RTT Handshake</span>
            </label>
          </div>

          <label class="flex items-center space-x-2 text-xs">
            <input v-model="form.enable" type="checkbox" class="rounded border-slate-700 bg-slate-800 text-indigo-500" />
            <span>启用此入站节点</span>
          </label>
          <div class="flex justify-end space-x-3 pt-4 border-t border-slate-800">
            <button type="button" @click="modalOpen = false" class="px-4 py-2 text-xs text-slate-300 hover:bg-slate-800 rounded-xl">取消</button>
            <button type="submit" :disabled="submitting" class="px-5 py-2 text-xs font-semibold bg-cyan-600 hover:bg-cyan-500 rounded-xl disabled:opacity-50">
              {{ submitting ? '保存中...' : (editing ? '保存修改' : '确认创建') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { api } from '../api'

const props = defineProps({
  inbounds: { type: Array, default: () => [] },
})
const emit = defineEmits(['changed'])

const modalOpen = ref(false)
const editing = ref(false)
const submitting = ref(false)
const current = ref(null)

const form = reactive({
  tag: '', type: 'vless', listen: '::', listen_port: 8443, enable: true,
  server_name: '', handshake_server: '', handshake_port: 443,
  private_key: '', short_id: '', congestion_control: 'bbr',
  auth_timeout: '3s', zero_rtt_handshake: false, certificate_provider: '', alpn: 'h3',
})

function genKey() {
  const bytes = new Uint8Array(32)
  crypto.getRandomValues(bytes)
  bytes[0] &= 248; bytes[31] &= 127; bytes[31] |= 64
  return btoa(String.fromCharCode(...bytes)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

function openCreate() {
  editing.value = false
  current.value = null
  Object.assign(form, {
    tag: '', type: 'vless', listen: '::', listen_port: 8443, enable: true,
    server_name: 'www.amazon.com', handshake_server: 'www.amazon.com', handshake_port: 443,
    private_key: genKey(), short_id: '0123456789abcdef', congestion_control: 'bbr',
    auth_timeout: '3s', zero_rtt_handshake: false, certificate_provider: '', alpn: 'h3',
  })
  modalOpen.value = true
}

function openEdit(n) {
  editing.value = true
  current.value = n
  Object.assign(form, {
    tag: n.tag, type: n.type, listen: n.listen || '::', listen_port: n.listen_port,
    enable: n.enable !== false, server_name: n.server_name || '', handshake_server: n.handshake_server || '',
    handshake_port: n.handshake_port || 443, private_key: n.private_key || '', short_id: n.short_id || '',
    congestion_control: n.congestion_control || 'bbr', auth_timeout: n.auth_timeout || '3s',
    zero_rtt_handshake: !!n.zero_rtt_handshake, certificate_provider: n.certificate_provider || '', alpn: n.alpn || 'h3',
  })
  modalOpen.value = true
}

async function save() {
  submitting.value = true
  try {
    if (editing.value) await api.updateInbound(current.value.id, { ...form })
    else await api.createInbound({ ...form })
    modalOpen.value = false
    emit('changed')
  } catch (e) {
    alert('保存失败：' + e.message)
  } finally {
    submitting.value = false
  }
}

async function remove(n) {
  if (!confirm(`确认删除节点 ${n.tag}？`)) return
  try {
    await api.deleteInbound(n.id)
    emit('changed')
  } catch (e) {
    alert('删除失败：' + e.message)
  }
}
</script>

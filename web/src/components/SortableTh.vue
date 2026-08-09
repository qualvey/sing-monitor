<template>
  <th
    :class="['py-3.5 px-4 cursor-pointer select-none whitespace-nowrap', active ? 'text-indigo-300' : '', align === 'right' ? 'text-right' : 'text-left']"
    @click="emit('sort')"
  >
    <span :class="['inline-flex items-center gap-1', align === 'right' ? 'justify-end w-full' : '']">
      {{ label }}
      <span v-if="active" class="text-indigo-400 text-[10px]">{{ dir === 'asc' ? '▲' : '▼' }}</span>
    </span>
  </th>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  active: { type: Boolean, default: false },
  dir: { type: String, default: 'asc' },
  align: { type: String, default: 'left' },
})
const emit = defineEmits(['sort'])
</script>

<script setup lang="ts">
defineProps<{
  label: string
  modelValue: string
  options: Array<{ name: string; value: string }>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <!-- 基础选择框只暴露字符串值，选项文案和值由调用方传入。 -->
  <label class="grid gap-2 text-sm text-[var(--page-muted)]">
    <span>{{ label }}</span>
    <select
      :value="modelValue"
      class="h-11 rounded-lg border border-[var(--border-soft)] bg-[var(--surface-input)] px-3 text-[15px] text-[var(--page-text)] outline-none transition focus:border-[var(--focus)] focus:ring-4 focus:ring-[var(--focus-ring)]"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option v-for="option in options" :key="option.value" :value="option.value">
        {{ option.name }}
      </option>
    </select>
  </label>
</template>

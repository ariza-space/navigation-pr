<script setup lang="ts">
defineProps<{
  label: string
  modelValue: string
  required?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <label class="grid gap-2 text-sm text-[var(--page-muted)]">
    <span>{{ label }}</span>
    <textarea
      :value="modelValue"
      :required="required"
      class="min-h-24 resize-y rounded-lg border border-[var(--border-soft)] bg-[var(--surface-input)] px-3 py-3 text-[15px] text-[var(--page-text)] outline-none transition placeholder:text-[var(--page-soft)] focus:border-[var(--focus)] focus:ring-4 focus:ring-[var(--focus-ring)]"
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />
  </label>
</template>

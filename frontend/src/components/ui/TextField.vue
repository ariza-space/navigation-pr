<script setup lang="ts">
withDefaults(defineProps<{
  label: string
  modelValue: string | number
  type?: string
  required?: boolean
  placeholder?: string
  autocomplete?: string
}>(), {
  type: 'text',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()
</script>

<template>
  <label class="grid gap-2 text-sm text-[var(--page-muted)]">
    <span>{{ label }}</span>
    <input
      :value="modelValue"
      :type="type"
      :required="required"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      class="h-11 rounded-lg border border-[var(--border-soft)] bg-[var(--surface-input)] px-3 text-[15px] text-[var(--page-text)] outline-none transition placeholder:text-[var(--page-soft)] focus:border-[var(--focus)] focus:ring-4 focus:ring-[var(--focus-ring)]"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </label>
</template>

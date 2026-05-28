<script setup lang="ts">
defineProps<{
  categories: string[]
  active: string
}>()

const emit = defineEmits<{
  change: [category: string]
}>()
</script>

<template>
  <nav class="mb-6 flex flex-wrap gap-2.5">
    <button
      v-for="category in categories"
      :key="category"
      type="button"
      variant="outline"
      :class="[
        'rounded-full border px-4 py-2.5 text-sm transition',
        category === active
          ? 'border-transparent [background:var(--accent-bg)] font-semibold text-[var(--accent-text)]'
          : 'border-[var(--border-soft)] bg-[var(--surface)] text-[var(--page-muted)] hover:border-[var(--border-hover)] hover:text-[var(--page-text)]',
      ]"
      @click="emit('change', category)"
    >
      {{ category }}
    </button>
  </nav>
</template>

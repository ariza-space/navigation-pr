<script setup lang="ts">
import { computed } from 'vue'

import { cn } from '@/lib/utils'

const props = withDefaults(defineProps<{
  variant?: 'default' | 'ghost' | 'outline' | 'danger'
  size?: 'default' | 'sm' | 'icon'
  type?: 'button' | 'submit' | 'reset'
  class?: string
}>(), {
  variant: 'default',
  size: 'default',
  type: 'button',
})

const classes = computed(() => cn(
  'inline-flex items-center justify-center gap-2 rounded-lg text-sm font-semibold transition focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-[var(--focus-ring)] disabled:pointer-events-none disabled:opacity-50',
  props.size === 'default' && 'h-10 px-4',
  props.size === 'sm' && 'h-8 px-3 text-xs',
  props.size === 'icon' && 'h-9 w-9 p-0',
  props.variant === 'default' && 'border border-transparent bg-[var(--accent-bg)] text-[var(--accent-text)] hover:brightness-105',
  props.variant === 'ghost' && 'border border-transparent bg-transparent text-[var(--page-text)] hover:bg-[var(--surface-hover)]',
  props.variant === 'outline' && 'border border-[var(--border)] bg-[var(--surface)] text-[var(--page-text)] hover:border-[var(--border-hover)] hover:bg-[var(--surface-hover)]',
  props.variant === 'danger' && 'border border-[var(--danger-border)] bg-[var(--danger-bg)] text-[var(--danger-text)] hover:brightness-105',
  props.class,
))
</script>

<template>
  <button :type="type" :class="classes">
    <slot />
  </button>
</template>

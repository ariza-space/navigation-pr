<script setup lang="ts">
import { X } from 'lucide-vue-next'

import UiButton from '@/components/ui/Button.vue'

defineProps<{
  open: boolean
  title: string
  wide?: boolean
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <Teleport to="body">
    <!-- Teleport 到 body，避免父级 overflow 或 z-index 影响弹窗遮罩。 -->
    <div
      v-if="open"
      class="fixed inset-0 z-50 grid place-items-center overflow-y-auto bg-[oklch(12%_.035_260_/_0.72)] p-4 backdrop-blur-xl"
      @mousedown.self="emit('close')"
    >
      <!-- wide 只控制最大宽度，内容布局由调用方自行决定。 -->
      <section
        role="dialog"
        aria-modal="true"
        :aria-label="title"
        :class="[
          'w-full rounded-[24px] border border-[var(--border)] bg-[var(--surface-solid)] p-5 shadow-dialog sm:p-6',
          wide ? 'max-w-3xl' : 'max-w-xl',
        ]"
      >
        <header class="mb-5 flex items-center justify-between gap-4">
          <h2 class="text-xl font-semibold text-[var(--page-text)]">{{ title }}</h2>
          <UiButton variant="outline" size="icon" aria-label="关闭" @click="emit('close')">
            <X class="h-4 w-4" />
          </UiButton>
        </header>
        <slot />
      </section>
    </div>
  </Teleport>
</template>

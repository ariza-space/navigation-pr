<script setup lang="ts">
import UiDialog from '@/components/ui/Dialog.vue'
import { emojiOptions } from '@/lib/options'

defineProps<{
  open: boolean
  modelValue: string
}>()

const emit = defineEmits<{
  close: []
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <UiDialog :open="open" title="选择图标" @close="emit('close')">
    <!-- 点击 emoji 后立即更新 v-model 并关闭弹窗，减少一次额外确认。 -->
    <div class="grid max-h-[390px] grid-cols-5 gap-2 overflow-auto sm:grid-cols-10">
      <button
        v-for="icon in emojiOptions"
        :key="icon"
        type="button"
        :title="icon"
        :class="[
          'grid h-11 place-items-center rounded-lg border border-[var(--border-soft)] bg-[var(--surface-input)] text-xl transition hover:border-[var(--border-hover)] hover:bg-[var(--surface-hover)]',
          icon === modelValue && 'border-[var(--focus)] ring-4 ring-[var(--focus-ring)]',
        ]"
        @click="emit('update:modelValue', icon); emit('close')"
      >
        {{ icon }}
      </button>
    </div>
  </UiDialog>
</template>

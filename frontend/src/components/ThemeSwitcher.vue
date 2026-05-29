<script setup lang="ts">
import { Check, ChevronUp, Palette } from 'lucide-vue-next'
import { ref } from 'vue'

import UiButton from '@/components/ui/Button.vue'
import { useTheme } from '@/composables/useTheme'

const { activeTheme, currentTheme, themeOptions, applyTheme } = useTheme()
// 主题切换器是悬浮控件，局部 open 状态不需要同步到全局 store。
const open = ref(false)
</script>

<template>
  <div class="fixed bottom-3 left-3 z-30 sm:bottom-6 sm:left-6">
    <div
      v-if="open"
      class="mb-2 grid w-52 gap-1 rounded-[18px] border border-[var(--border)] bg-[var(--surface-strong)] p-2 shadow-card backdrop-blur-xl"
    >
      <!-- 第二个参数表示用户主动选择，需要持久化到本地偏好。 -->
      <button
        v-for="theme in themeOptions"
        :key="theme.value"
        type="button"
        :class="[
          'flex h-10 items-center justify-between gap-3 rounded-lg px-3 text-sm text-[var(--page-text)] hover:bg-[var(--surface-hover)]',
          theme.value === activeTheme && 'bg-[var(--accent-bg)] font-semibold text-[var(--accent-text)]',
        ]"
        @click="applyTheme(theme.value, true); open = false"
      >
        <span class="flex items-center gap-2">
          <span class="h-3 w-3 rounded-full shadow-[0_0_18px_currentColor]" :style="{ color: theme.swatch, background: theme.swatch }" />
          {{ theme.name }}
        </span>
        <Check v-if="theme.value === activeTheme" class="h-4 w-4" />
      </button>
    </div>
    <UiButton variant="outline" class="rounded-full bg-[var(--surface-strong)] shadow-card backdrop-blur-xl" @click="open = !open">
      <Palette class="h-4 w-4" />
      <span>主题：{{ currentTheme.name }}</span>
      <ChevronUp class="h-4 w-4" />
    </UiButton>
  </div>
</template>

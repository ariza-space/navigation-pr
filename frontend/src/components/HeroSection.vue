<script setup lang="ts">
import { Search } from 'lucide-vue-next'

import type { AppSettings, Stats } from '@/types/api'

defineProps<{
  settings: AppSettings
  stats: Stats
  query: string
}>()

const emit = defineEmits<{
  'update:query': [value: string]
}>()
</script>

<template>
  <!-- 首页头图区负责搜索入口和概览统计，实际搜索请求在 App.vue 中节流。 -->
  <section class="mb-8 grid items-center gap-6 lg:grid-cols-[1.12fr_.88fr]">
    <div>
      <div class="inline-flex items-center gap-2 rounded-full border border-[var(--border)] bg-[var(--surface)] px-4 py-2 text-sm text-[var(--accent)] backdrop-blur-xl">
        <span class="h-2 w-2 rounded-full bg-[var(--accent-strong)] shadow-[0_0_18px_var(--accent-strong)]" />
        <span>{{ settings.badge }}</span>
      </div>
      <h1 class="my-5 max-w-4xl text-[clamp(38px,6vw,76px)] font-semibold leading-[1.02] tracking-normal text-[var(--page-text)]">
        {{ settings.heroTitle }}
      </h1>
      <p class="max-w-[65ch] text-[17px] leading-8 text-[var(--page-muted)]">{{ settings.subtitle }}</p>
    </div>

    <div class="rounded-[24px] border border-[var(--border)] bg-[var(--surface)] p-5 shadow-card backdrop-blur-xl">
      <label class="mb-4 flex h-14 items-center gap-3 rounded-[18px] border border-[var(--border-soft)] bg-[var(--surface-input)] px-4 text-[var(--page-muted)]">
        <Search class="h-5 w-5 shrink-0" />
        <input
          :value="query"
          class="min-w-0 flex-1 bg-transparent text-[16px] text-[var(--page-text)] outline-none placeholder:text-[var(--page-soft)]"
          placeholder="搜索工具、文档、服务..."
          @input="emit('update:query', ($event.target as HTMLInputElement).value)"
        />
      </label>
      <div class="grid gap-3 sm:grid-cols-3">
        <div class="stat-box">
          <b>{{ stats.siteCount }}</b>
          <small>常用站点</small>
        </div>
        <div class="stat-box">
          <b>{{ stats.categoryCount }}</b>
          <small>分类分组</small>
        </div>
        <div class="stat-box">
          <b>{{ stats.coverage }}</b>
          <small>日常覆盖</small>
        </div>
      </div>
    </div>
  </section>
</template>

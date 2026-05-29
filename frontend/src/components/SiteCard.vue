<script setup lang="ts">
import { Edit3, ExternalLink, Trash2 } from 'lucide-vue-next'

import UiButton from '@/components/ui/Button.vue'
import type { Site, UserSession } from '@/types/api'

defineProps<{
  site: Site
  user: UserSession | null
}>()

const emit = defineEmits<{
  edit: [site: Site]
  delete: [site: Site]
}>()
</script>

<template>
  <!-- 整张卡片是外链，管理按钮用 prevent 避免点击时同时打开站点。 -->
  <a
    :href="site.url"
    target="_blank"
    rel="noopener noreferrer"
    class="site-card group"
    :style="{ '--glow': site.glow || 'rgba(96,165,250,.45)' }"
  >
    <div v-if="user" class="absolute right-3 top-3 z-10 flex gap-2 opacity-100 transition sm:opacity-0 sm:group-hover:opacity-100">
      <UiButton variant="outline" size="icon" title="编辑" @click.prevent="emit('edit', site)">
        <Edit3 class="h-4 w-4" />
      </UiButton>
      <UiButton variant="danger" size="icon" title="删除" @click.prevent="emit('delete', site)">
        <Trash2 class="h-4 w-4" />
      </UiButton>
    </div>
    <div class="relative mb-5 grid h-12 w-12 place-items-center rounded-[16px] border border-[var(--border-soft)] bg-[var(--surface)] text-2xl">
      {{ site.icon || '🔗' }}
    </div>
    <h3 class="relative mb-2 pr-20 text-[19px] font-semibold text-[var(--page-text)]">{{ site.name }}</h3>
    <p class="relative line-clamp-3 text-sm leading-6 text-[var(--page-muted)]">{{ site.description || site.category }}</p>
    <ExternalLink class="absolute bottom-5 right-5 h-5 w-5 text-[var(--accent)] opacity-75" />
  </a>
</template>

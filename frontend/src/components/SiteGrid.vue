<script setup lang="ts">
import SiteCard from '@/components/SiteCard.vue'
import type { Site, UserSession } from '@/types/api'

defineProps<{
  sites: Site[]
  user: UserSession | null
}>()

const emit = defineEmits<{
  edit: [site: Site]
  delete: [site: Site]
}>()
</script>

<template>
  <section class="grid gap-[18px] sm:grid-cols-2 xl:grid-cols-4">
    <div v-if="!sites.length" class="col-span-full grid min-h-[170px] place-items-center rounded-[24px] border border-dashed border-[var(--border)] bg-[var(--surface)] text-[var(--page-soft)]">
      没有找到匹配的站点
    </div>
    <SiteCard
      v-for="site in sites"
      :key="site.id"
      :site="site"
      :user="user"
      @edit="emit('edit', $event)"
      @delete="emit('delete', $event)"
    />
  </section>
</template>

<script setup lang="ts">
import { Pencil, Trash2 } from 'lucide-vue-next'

import UiButton from '@/components/ui/Button.vue'
import UiDialog from '@/components/ui/Dialog.vue'
import type { CategoryStat } from '@/types/api'

defineProps<{
  open: boolean
  categories: CategoryStat[]
  error: string
}>()

const emit = defineEmits<{
  close: []
  rename: [category: CategoryStat]
  delete: [category: CategoryStat]
}>()
</script>

<template>
  <UiDialog :open="open" title="分类管理" @close="emit('close')">
    <!-- 分类行只展示统计和操作，真正的重命名/删除确认由父组件处理。 -->
    <div class="grid max-h-[420px] gap-2 overflow-auto">
      <div v-if="error" class="rounded-lg border border-[var(--danger-border)] bg-[var(--danger-bg)] px-3 py-2 text-sm text-[var(--danger-text)]">
        {{ error }}
      </div>
      <div v-else-if="!categories.length" class="grid min-h-32 place-items-center rounded-[18px] border border-dashed border-[var(--border)] text-[var(--page-soft)]">
        暂无分类
      </div>
      <div
        v-for="category in categories"
        :key="category.name"
        class="flex items-center justify-between gap-3 rounded-[16px] border border-[var(--border-soft)] bg-[var(--surface)] p-3"
      >
        <div class="min-w-0">
          <b class="block truncate text-[var(--page-text)]">{{ category.name }}</b>
          <small class="text-[var(--page-soft)]">{{ category.count }} 个站点</small>
        </div>
        <div class="flex shrink-0 gap-2">
          <UiButton variant="outline" size="sm" @click="emit('rename', category)">
            <Pencil class="h-4 w-4" /> 编辑
          </UiButton>
          <UiButton variant="danger" size="sm" @click="emit('delete', category)">
            <Trash2 class="h-4 w-4" /> 删除
          </UiButton>
        </div>
      </div>
    </div>
  </UiDialog>
</template>

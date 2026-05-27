<script setup lang="ts">
import { ChevronDown } from 'lucide-vue-next'
import { computed, reactive, ref, watch } from 'vue'

import EmojiDialog from '@/components/EmojiDialog.vue'
import UiButton from '@/components/ui/Button.vue'
import UiDialog from '@/components/ui/Dialog.vue'
import TextArea from '@/components/ui/TextArea.vue'
import TextField from '@/components/ui/TextField.vue'
import { emojiOptions, glowOptions } from '@/lib/options'
import type { Site, SiteInput } from '@/types/api'

const props = defineProps<{
  open: boolean
  site: Site | null
  categories: string[]
  error: string
}>()

const emit = defineEmits<{
  close: []
  save: [input: SiteInput, id?: string]
}>()

const form = reactive<SiteInput>({
  name: '',
  category: '',
  url: '',
  icon: emojiOptions[0],
  sort: 0,
  glow: glowOptions[0].value,
  description: '',
})

const categoryMenuOpen = ref(false)
const emojiOpen = ref(false)

const categoryOptions = computed(() => props.categories.filter(category => category !== '全部'))
const filteredCategories = computed(() => {
  const keyword = form.category.trim().toLowerCase()
  return categoryOptions.value.filter(category => !keyword || category.toLowerCase().includes(keyword))
})

watch(() => [props.open, props.site] as const, () => {
  if (!props.open) return
  form.name = props.site?.name || ''
  form.category = props.site?.category || ''
  form.url = props.site?.url || ''
  form.icon = props.site?.icon || emojiOptions[0]
  form.sort = props.site?.sort || 0
  form.glow = props.site?.glow || glowOptions[0].value
  form.description = props.site?.description || ''
  categoryMenuOpen.value = false
}, { immediate: true })

function submit() {
  emit('save', { ...form, sort: Number(form.sort || 0) }, props.site?.id)
}
</script>

<template>
  <UiDialog :open="open" :title="site ? '编辑站点' : '新增站点'" @close="emit('close')">
    <form class="grid gap-4" @submit.prevent="submit">
      <div class="grid gap-4 sm:grid-cols-2">
        <TextField v-model="form.name" label="名称" required />
        <div class="relative grid gap-2 text-sm text-[var(--page-muted)]">
          <span>分类</span>
          <div class="flex overflow-hidden rounded-lg border border-[var(--border-soft)] bg-[var(--surface-input)] focus-within:border-[var(--focus)] focus-within:ring-4 focus-within:ring-[var(--focus-ring)]">
            <input
              v-model="form.category"
              required
              autocomplete="off"
              class="h-11 min-w-0 flex-1 bg-transparent px-3 text-[15px] text-[var(--page-text)] outline-none"
              @focus="categoryMenuOpen = true"
            />
            <button class="grid w-11 place-items-center border-l border-[var(--border-soft)] text-[var(--accent)]" type="button" @click="categoryMenuOpen = !categoryMenuOpen">
              <ChevronDown class="h-4 w-4" />
            </button>
          </div>
          <div
            v-if="categoryMenuOpen"
            class="absolute left-0 right-0 top-full z-20 mt-2 max-h-52 overflow-auto rounded-[16px] border border-[var(--border)] bg-[var(--surface-strong)] p-2 shadow-card"
          >
            <div v-if="!filteredCategories.length" class="px-3 py-2 text-xs leading-5 text-[var(--page-soft)]">
              {{ form.category ? '没有匹配分类，保存后会创建为新分类。' : '暂无可选分类，可以直接输入新分类。' }}
            </div>
            <button
              v-for="category in filteredCategories"
              :key="category"
              type="button"
              class="flex w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm text-[var(--page-text)] hover:bg-[var(--surface-hover)]"
              @mousedown.prevent
              @click="form.category = category; categoryMenuOpen = false"
            >
              <span>{{ category }}</span>
              <small class="text-[var(--page-soft)]">已有</small>
            </button>
          </div>
        </div>
        <TextField v-model="form.url" class="sm:col-span-2" label="地址" type="url" placeholder="https://example.com" required />
        <div class="grid gap-2 text-sm text-[var(--page-muted)]">
          <span>图标</span>
          <UiButton variant="outline" class="h-11 justify-between" @click="emojiOpen = true">
            <span class="flex items-center gap-3">
              <span class="grid h-8 w-8 place-items-center rounded-lg bg-[var(--surface)] text-xl">{{ form.icon }}</span>
              选择 emoji 图标
            </span>
            <span>›</span>
          </UiButton>
        </div>
        <TextField v-model="form.sort" label="排序" type="number" />
        <div class="grid gap-2 sm:col-span-2">
          <span class="text-sm text-[var(--page-muted)]">光效颜色</span>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <button
              v-for="option in glowOptions"
              :key="option.value"
              type="button"
              :class="[
                'flex h-10 items-center justify-center gap-2 rounded-lg border text-sm text-[var(--page-text)] transition hover:border-[var(--border-hover)] hover:bg-[var(--surface-hover)]',
                option.value === form.glow ? 'border-transparent bg-[var(--accent-bg)] font-semibold text-[var(--accent-text)]' : 'border-[var(--border-soft)] bg-[var(--surface-input)]',
              ]"
              @click="form.glow = option.value"
            >
              <span class="h-3.5 w-3.5 rounded-full shadow-[0_0_18px_currentColor]" :style="{ color: option.value, background: option.value }" />
              {{ option.name }}
            </button>
          </div>
        </div>
        <TextArea v-model="form.description" class="sm:col-span-2" label="描述" />
      </div>
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p class="min-h-5 text-sm text-[var(--danger-text)]">{{ error }}</p>
        <div class="flex justify-end gap-2">
          <UiButton variant="outline" @click="emit('close')">取消</UiButton>
          <UiButton type="submit">保存</UiButton>
        </div>
      </div>
    </form>
    <EmojiDialog v-model="form.icon" :open="emojiOpen" @close="emojiOpen = false" />
  </UiDialog>
</template>

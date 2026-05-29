<script setup lang="ts">
import { reactive, watch } from 'vue'

import UiButton from '@/components/ui/Button.vue'
import UiDialog from '@/components/ui/Dialog.vue'
import SelectField from '@/components/ui/SelectField.vue'
import TextArea from '@/components/ui/TextArea.vue'
import TextField from '@/components/ui/TextField.vue'
import { themeOptions } from '@/composables/useTheme'
import type { AppSettings } from '@/types/api'

const props = defineProps<{
  open: boolean
  settings: AppSettings
  error: string
}>()

const emit = defineEmits<{
  close: []
  save: [settings: AppSettings]
}>()

const form = reactive<AppSettings>({ siteTitle: '', badge: '', heroTitle: '', subtitle: '', theme: 'dark' })

// 设置可能在后台加载或保存后变化，打开弹窗时以最新服务端值为准。
watch(() => [props.open, props.settings] as const, () => {
  if (!props.open) return
  Object.assign(form, props.settings)
}, { immediate: true })
</script>

<template>
  <UiDialog :open="open" title="页面设置" @close="emit('close')">
    <!-- 本地表单提交给父组件保存，成功后再由父组件应用默认主题。 -->
    <form class="grid gap-4" @submit.prevent="emit('save', { ...form })">
      <TextField v-model="form.siteTitle" label="Title" required />
      <TextField v-model="form.badge" label="徽章" required />
      <TextField v-model="form.heroTitle" label="主标题" required />
      <TextArea v-model="form.subtitle" label="简介" required />
      <SelectField v-model="form.theme" label="全局默认主题" :options="themeOptions" />
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p class="min-h-5 text-sm text-[var(--danger-text)]">{{ error }}</p>
        <UiButton type="submit">保存</UiButton>
      </div>
    </form>
  </UiDialog>
</template>

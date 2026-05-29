<script setup lang="ts">
import { reactive, watch } from 'vue'

import UiButton from '@/components/ui/Button.vue'
import UiDialog from '@/components/ui/Dialog.vue'
import TextField from '@/components/ui/TextField.vue'
import type { UserSession } from '@/types/api'

const props = defineProps<{
  open: boolean
  user: UserSession | null
  error: string
}>()

const emit = defineEmits<{
  close: []
  save: [input: { username: string; currentPassword: string; newPassword: string }]
}>()

const form = reactive({ username: '', currentPassword: '', newPassword: '' })

// 每次打开都重置敏感字段，避免上一次输入的密码留在弹窗里。
watch(() => props.open, open => {
  if (!open) return
  form.username = props.user?.username || ''
  form.currentPassword = ''
  form.newPassword = ''
})
</script>

<template>
  <UiDialog :open="open" title="修改账号密码" @close="emit('close')">
    <!-- 新密码允许留空，此时后端只更新账号名。 -->
    <form class="grid gap-4" @submit.prevent="emit('save', { ...form })">
      <TextField v-model="form.username" label="账号" required />
      <TextField v-model="form.currentPassword" label="当前密码" type="password" autocomplete="current-password" required />
      <TextField v-model="form.newPassword" label="新密码" type="password" autocomplete="new-password" placeholder="留空则只修改账号" />
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p class="min-h-5 text-sm text-[var(--danger-text)]">{{ error }}</p>
        <UiButton type="submit">保存</UiButton>
      </div>
    </form>
  </UiDialog>
</template>

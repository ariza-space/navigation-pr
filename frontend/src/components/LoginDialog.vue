<script setup lang="ts">
import { reactive } from 'vue'

import UiButton from '@/components/ui/Button.vue'
import TextField from '@/components/ui/TextField.vue'

defineProps<{
  open: boolean
  error: string
}>()

const emit = defineEmits<{
  close: []
  login: [input: { username: string; password: string }]
}>()

const form = reactive({ username: '', password: '' })
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-[60] grid place-items-center bg-[oklch(12%_.035_260_/_0.86)] p-4 backdrop-blur-xl">
      <section class="w-full max-w-md rounded-[24px] border border-[var(--border)] bg-[var(--surface-strong)] p-6 shadow-dialog">
        <h2 class="mb-2 text-2xl font-semibold text-[var(--page-text)]">登录导航站</h2>
        <p class="mb-5 leading-7 text-[var(--page-muted)]">请输入管理员账号密码</p>
        <form class="grid gap-4" @submit.prevent="emit('login', { ...form })">
          <TextField v-model="form.username" label="账号" autocomplete="username" required />
          <TextField v-model="form.password" label="密码" type="password" autocomplete="current-password" required />
          <p class="min-h-5 text-sm text-[var(--danger-text)]">{{ error }}</p>
          <UiButton type="submit" variant="outline">登录</UiButton>
          <UiButton variant="outline" @click="emit('close')">暂不登录，继续浏览</UiButton>
        </form>
      </section>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { LogIn, LogOut, Settings, UserCog } from 'lucide-vue-next'
import { ref } from 'vue'

import UiButton from '@/components/ui/Button.vue'
import type { UserSession } from '@/types/api'

defineProps<{
  user: UserSession | null
}>()

const emit = defineEmits<{
  login: []
  account: []
  settings: []
  logout: []
}>()

// 菜单展开只属于这个入口自身，登录态变化由父组件通过 user 传入。
const open = ref(false)

function clickUser(user: UserSession | null) {
  if (!user) {
    emit('login')
    return
  }
  open.value = !open.value
}

function menu(action: 'account' | 'settings' | 'logout') {
  // 先收起菜单再派发动作，避免弹窗打开后下拉菜单仍留在页面上。
  open.value = false
  if (action === 'account') emit('account')
  if (action === 'settings') emit('settings')
  if (action === 'logout') emit('logout')
}
</script>

<template>
  <div class="relative">
    <UiButton variant="outline" class="rounded-full backdrop-blur-xl" @click="clickUser(user)">
      <LogIn v-if="!user" class="h-4 w-4" />
      <UserCog v-else class="h-4 w-4" />
      <span>{{ user?.username || '登录' }}</span>
    </UiButton>
    <div
      v-if="open && user"
      class="absolute right-0 top-[calc(100%+10px)] z-20 grid w-44 gap-1 rounded-[16px] border border-[var(--border)] bg-[var(--surface-strong)] p-2 shadow-card backdrop-blur-xl"
    >
      <button class="menu-item" type="button" @click="menu('settings')">
        <Settings class="h-4 w-4" /> 设置
      </button>
      <button class="menu-item" type="button" @click="menu('account')">
        <UserCog class="h-4 w-4" /> 修改账号密码
      </button>
      <button class="menu-item" type="button" @click="menu('logout')">
        <LogOut class="h-4 w-4" /> 退出登录
      </button>
    </div>
  </div>
</template>

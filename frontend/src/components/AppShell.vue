<script setup lang="ts">
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'
import UserMenu from '@/components/UserMenu.vue'
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
</script>

<template>
  <main class="relative mx-auto min-h-screen w-[min(1180px,calc(100%-24px))] py-7 sm:w-[min(1180px,calc(100%-40px))] sm:py-14">
    <div class="relative z-10 mb-6 flex justify-end">
      <UserMenu
        :user="user"
        @login="emit('login')"
        @account="emit('account')"
        @settings="emit('settings')"
        @logout="emit('logout')"
      />
    </div>
    <slot />
  </main>
  <ThemeSwitcher />
</template>

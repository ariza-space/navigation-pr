import { ref } from 'vue'

import { getSettings, updateSettings } from '@/lib/api'
import type { AppSettings } from '@/types/api'

export function useSettings() {
  const settings = ref<AppSettings>({
    siteTitle: '导航站',
    badge: 'DEV PORTAL / 个人导航站',
    heroTitle: '常用站点导航',
    subtitle: '聚合了常用网站',
    theme: 'dark',
  })

  async function loadSettings() {
    settings.value = await getSettings()
    document.title = settings.value.siteTitle
    return settings.value
  }

  async function saveSettings(input: AppSettings) {
    settings.value = await updateSettings(input)
    document.title = settings.value.siteTitle
    return settings.value
  }

  return {
    settings,
    loadSettings,
    saveSettings,
  }
}

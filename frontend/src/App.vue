<script setup lang="ts">
import { Plus, Tags } from 'lucide-vue-next'
import { onMounted, ref, watch } from 'vue'

import AccountDialog from '@/components/AccountDialog.vue'
import AppShell from '@/components/AppShell.vue'
import CategoryDialog from '@/components/CategoryDialog.vue'
import CategoryTabs from '@/components/CategoryTabs.vue'
import HeroSection from '@/components/HeroSection.vue'
import LoginDialog from '@/components/LoginDialog.vue'
import NoteWorkspace from '@/components/NoteWorkspace.vue'
import SettingsDialog from '@/components/SettingsDialog.vue'
import SiteDialog from '@/components/SiteDialog.vue'
import SiteGrid from '@/components/SiteGrid.vue'
import UiButton from '@/components/ui/Button.vue'
import { useAuth } from '@/composables/useAuth'
import { useNotes } from '@/composables/useNotes'
import { useSettings } from '@/composables/useSettings'
import { useSites } from '@/composables/useSites'
import { useTheme } from '@/composables/useTheme'
import { debounce } from '@/lib/utils'
import type { AccountInput, AppSettings, CategoryStat, Note, Site, SiteInput } from '@/types/api'

const auth = useAuth()
const sites = useSites()
const notes = useNotes()
const settingsStore = useSettings()
const { applyDefaultTheme } = useTheme()

// 顶层页面只保存当前视图和弹窗开关，业务数据交给各 composable 维护。
const activeModule = ref<'sites' | 'notes'>('sites')
const siteDialogOpen = ref(false)
const editingSite = ref<Site | null>(null)
const siteError = ref('')
const categoryDialogOpen = ref(false)
const categoryError = ref('')
const accountDialogOpen = ref(false)
const accountError = ref('')
const settingsDialogOpen = ref(false)
const settingsError = ref('')
const bootError = ref('')

// 搜索输入变化较频繁，延迟请求可以减少后端和 SQLite 的无效查询。
const debouncedLoadSites = debounce(async () => {
  try {
    await sites.loadSitesOnly()
  } catch (error) {
    bootError.value = error instanceof Error ? error.message : '读取站点数据失败'
  }
}, 250)

watch(sites.query, () => {
  debouncedLoadSites()
})

const debouncedLoadNotes = debounce(async () => {
  try {
    await notes.loadNotes()
  } catch (error) {
    if (auth.handleAuthError(error)) return
    notes.error.value = error instanceof Error ? error.message : '读取笔记失败'
  }
}, 250)

watch(notes.query, () => {
  debouncedLoadNotes()
})

// 登录弹窗完成后，如果用户当前停留在文档模块，需要补拉一次文档列表。
watch(auth.user, async (user) => {
  if (user && activeModule.value === 'notes') {
    await loadNotes()
  }
})

async function bootstrap() {
  // 首屏初始化顺序：会话用于决定管理能力，设置用于主题，站点数据用于首页内容。
  await auth.refreshSession()
  try {
    const settings = await settingsStore.loadSettings()
    applyDefaultTheme(settings.theme)
    await sites.loadAll()
  } catch (error) {
    bootError.value = error instanceof Error ? error.message : '页面加载失败'
  }
}

function openSiteDialog(site: Site | null = null) {
  // 新增和编辑共用一个弹窗，传入 site 时由子组件回填表单。
  if (!auth.requireLogin()) return
  siteError.value = ''
  editingSite.value = site
  siteDialogOpen.value = true
}

async function saveSite(input: SiteInput, id?: string) {
  // 所有写操作都在这里统一处理登录过期，避免子组件直接理解认证细节。
  siteError.value = ''
  try {
    await sites.saveSite(input, id)
    siteDialogOpen.value = false
  } catch (error) {
    if (auth.handleAuthError(error)) return
    siteError.value = error instanceof Error ? error.message : '保存站点失败'
  }
}

async function removeSite(site: Site) {
  if (!auth.requireLogin()) return
  if (!window.confirm(`确定删除「${site.name}」吗？`)) return
  try {
    await sites.removeSite(site)
  } catch (error) {
    if (!auth.handleAuthError(error)) window.alert(error instanceof Error ? error.message : '删除站点失败')
  }
}

async function openCategoryDialog() {
  if (!auth.requireLogin()) return
  categoryError.value = ''
  categoryDialogOpen.value = true
  try {
    await sites.loadCategoryStats()
  } catch (error) {
    if (auth.handleAuthError(error)) return
    categoryError.value = error instanceof Error ? error.message : '读取分类失败'
  }
}

async function removeCategory(category: CategoryStat) {
  // 分类删除只清空站点的分类字段，不删除站点本身，所以确认文案要讲清影响范围。
  if (!auth.requireLogin()) return
  const message = `确定删除「${category.name}」分类吗？该分类下的 ${category.count} 个站点会保留，但分类会被清空。`
  if (!window.confirm(message)) return
  try {
    await sites.removeCategory(category.name)
  } catch (error) {
    if (!auth.handleAuthError(error)) window.alert(error instanceof Error ? error.message : '删除分类失败')
  }
}

async function renameCategory(category: CategoryStat) {
  if (!auth.requireLogin()) return
  const nextName = window.prompt('请输入新的分类名称', category.name)
  if (nextName === null) return
  const normalizedName = nextName.trim()
  if (!normalizedName || normalizedName === category.name) return
  try {
    await sites.renameCategory(category.name, normalizedName)
  } catch (error) {
    if (!auth.handleAuthError(error)) window.alert(error instanceof Error ? error.message : '重命名分类失败')
  }
}

async function saveAccount(input: AccountInput) {
  accountError.value = ''
  try {
    await auth.saveAccount(input)
    accountDialogOpen.value = false
  } catch (error) {
    if (auth.handleAuthError(error)) return
    accountError.value = error instanceof Error ? error.message : '保存账号失败'
  }
}

async function saveSettings(input: AppSettings) {
  settingsError.value = ''
  try {
    const settings = await settingsStore.saveSettings(input)
    applyDefaultTheme(settings.theme)
    settingsDialogOpen.value = false
  } catch (error) {
    if (auth.handleAuthError(error)) return
    settingsError.value = error instanceof Error ? error.message : '保存设置失败'
  }
}

async function changeCategory(category: string) {
  sites.category.value = category
  await sites.loadSitesOnly()
}

async function switchModule(module: 'sites' | 'notes') {
  // 切换到文档模块时立即加载数据，站点模块则依赖已有缓存和分类筛选。
  activeModule.value = module
  if (module === 'notes') {
    await loadNotes()
  }
}

async function loadNotes() {
  try {
    await notes.loadNotes()
  } catch (error) {
    if (auth.handleAuthError(error)) return
    notes.error.value = error instanceof Error ? error.message : '读取笔记失败'
  }
}

async function selectNote(note: Note) {
  try {
    await notes.selectNote(note)
  } catch (error) {
    if (auth.handleAuthError(error)) return
    notes.error.value = error instanceof Error ? error.message : '读取笔记失败'
  }
}

async function saveNote() {
  if (!auth.requireLogin()) return
  try {
    await notes.saveDraft()
  } catch (error) {
    if (auth.handleAuthError(error)) return
    notes.error.value = error instanceof Error ? error.message : '保存笔记失败'
  }
}

async function deleteNote() {
  // 文档删除是软删除，真实 Markdown 文件保留给用户后续恢复或手工处理。
  if (!auth.requireLogin()) return
  if (!notes.selected.value) return
  if (!window.confirm(`确定删除「${notes.selected.value.title}」吗？文件会保留，笔记会被软删除。`)) return
  try {
    await notes.removeSelected()
  } catch (error) {
    if (auth.handleAuthError(error)) return
    notes.error.value = error instanceof Error ? error.message : '删除笔记失败'
  }
}

async function syncNotes() {
  if (!auth.requireLogin()) return
  try {
    await notes.syncIndex()
  } catch (error) {
    if (auth.handleAuthError(error)) return
    notes.error.value = error instanceof Error ? error.message : '同步笔记失败'
  }
}

function updateNoteQuery(query: string) {
  notes.query.value = query
}

onMounted(() => {
  bootstrap()
})
</script>

<template>
  <AppShell
    :user="auth.user.value"
    :active-module="activeModule"
    @login="auth.loginOpen.value = true"
    @account="accountDialogOpen = true"
    @settings="settingsDialogOpen = true"
    @logout="auth.logout"
    @module="switchModule"
  >
    <template v-if="activeModule === 'sites'">
      <HeroSection v-model:query="sites.query.value" :settings="settingsStore.settings.value" :stats="sites.stats.value" />
      <CategoryTabs :categories="sites.categories.value" :active="sites.category.value" @change="changeCategory" />

      <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h2 class="text-2xl font-semibold text-[var(--page-text)]">常用入口</h2>
          <p class="mt-1 text-sm text-[var(--page-soft)]">点击卡片快速访问</p>
        </div>
        <div v-if="auth.user.value" class="flex flex-wrap gap-2">
          <UiButton variant="outline" @click="openCategoryDialog">
            <Tags class="h-4 w-4" /> 分类管理
          </UiButton>
          <UiButton @click="openSiteDialog()">
            <Plus class="h-4 w-4" /> 新增站点
          </UiButton>
        </div>
      </div>

      <div v-if="bootError" class="mb-4 rounded-[16px] border border-[var(--danger-border)] bg-[var(--danger-bg)] px-4 py-3 text-sm text-[var(--danger-text)]">
        {{ bootError }}
      </div>
      <SiteGrid :sites="sites.sites.value" :user="auth.user.value" @edit="openSiteDialog" @delete="removeSite" />
    </template>

    <NoteWorkspace
      v-else
      :notes="notes.notes.value"
      :selected="notes.selected.value"
      :draft="notes.draft.value"
      :user="auth.user.value"
      :loading="notes.loading.value"
      :saving="notes.saving.value"
      :syncing="notes.syncing.value"
      :error="notes.error.value"
      :query="notes.query.value"
      @new="notes.resetDraft"
      @select="selectNote"
      @sync="syncNotes"
      @save="saveNote"
      @delete="deleteNote"
      @search="updateNoteQuery"
      @update:draft="notes.draft.value = $event"
    />
  </AppShell>

  <LoginDialog
    :open="auth.loginOpen.value"
    :error="auth.loginError.value"
    @close="auth.setAnonymous"
    @login="auth.login"
  />
  <SiteDialog
    :open="siteDialogOpen"
    :site="editingSite"
    :categories="sites.categories.value"
    :error="siteError"
    @close="siteDialogOpen = false"
    @save="saveSite"
  />
  <CategoryDialog
    :open="categoryDialogOpen"
    :categories="sites.categoryStats.value"
    :error="categoryError"
    @close="categoryDialogOpen = false"
    @rename="renameCategory"
    @delete="removeCategory"
  />
  <AccountDialog
    :open="accountDialogOpen"
    :user="auth.user.value"
    :error="accountError"
    @close="accountDialogOpen = false"
    @save="saveAccount"
  />
  <SettingsDialog
    :open="settingsDialogOpen"
    :settings="settingsStore.settings.value"
    :error="settingsError"
    @close="settingsDialogOpen = false"
    @save="saveSettings"
  />
</template>

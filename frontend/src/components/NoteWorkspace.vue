<script setup lang="ts">
import { Eye, FileText, Globe2, Loader2, Lock, Pencil, Pin, Plus, RefreshCw, Save, Search, Trash2 } from 'lucide-vue-next'
import { MdEditor, MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { computed, ref, watch } from 'vue'

import UiButton from '@/components/ui/Button.vue'
import { debounce } from '@/lib/utils'
import type { Note, NoteContent, NoteInput, UserSession } from '@/types/api'

const props = defineProps<{
  notes: Note[]
  selected: NoteContent | null
  draft: NoteInput
  user: UserSession | null
  loading: boolean
  saving: boolean
  syncing: boolean
  error: string
  query: string
}>()

const emit = defineEmits<{
  new: []
  select: [note: Note]
  sync: []
  save: []
  delete: []
  search: [query: string]
  'update:draft': [draft: NoteInput]
}>()

const tagText = ref('')
const mode = ref<'preview' | 'edit'>('edit')

// 列表搜索走父级 composable，子组件只做输入节流和事件派发。
const debouncedSearch = debounce((value: string) => {
  emit('search', value)
}, 250)

const activeID = computed(() => props.selected?.id || '')
const canEdit = computed(() => Boolean(props.user))
const isPreview = computed(() => Boolean(props.selected) && mode.value === 'preview')
const displayPreview = computed(() => isPreview.value || !canEdit.value)
// md-editor-v3 需要 v-model 字符串，这里把 props.draft 的不可变更新包装成双向模型。
const contentModel = computed({
  get: () => props.draft.content,
  set: (value: string) => patchDraft({ content: value }),
})

watch(() => props.query, (value) => {
  debouncedSearch(value)
})

// 标签在 API 中是数组，表单里用逗号文本编辑，选中文档切换时要重新同步。
watch(() => props.draft.tags, (tags) => {
  tagText.value = tags.join(', ')
}, { immediate: true })

// draft 由父组件持有，任何局部修改都通过整体拷贝向上更新，避免直接改 props。
function patchDraft(patch: Partial<NoteInput>) {
  emit('update:draft', { ...props.draft, ...patch })
}

function createNote() {
  mode.value = 'edit'
  emit('new')
}

function openNote(note: Note) {
  mode.value = 'preview'
  emit('select', note)
}

// 输入框允许用户自由输入，保存前再归一化成去空白、无空标签的数组。
function updateTags(value: string) {
  tagText.value = value
  const tags = value.split(',').map((tag) => tag.trim()).filter(Boolean)
  patchDraft({ tags })
}

function formatDate(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('zh-SG', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function visibilityText(visibility: Note['visibility']) {
  return visibility === 'public' ? '公开' : '隐私'
}
</script>

<template>
  <section class="notes-workspace">
    <!-- 左侧只承担检索、同步和选择文档，正文编辑状态留给右侧表单。 -->
    <div class="notes-panel notes-sidebar">
      <div class="flex items-center justify-between gap-3">
        <div>
          <h2 class="text-2xl font-semibold text-[var(--page-text)]">文档</h2>
          <p class="mt-1 text-sm text-[var(--page-soft)]">基于 Markdown</p>
        </div>
        <div class="flex shrink-0 gap-2">
          <UiButton v-if="canEdit" title="同步 Markdown 文件索引" type="button" variant="outline" :disabled="syncing" @click="emit('sync')">
            <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': syncing }" /> 索引同步
          </UiButton>
          <UiButton v-if="canEdit" title="新建文档" type="button" variant="outline" @click="createNote">
            <Plus class="h-4 w-4" /> 新建
          </UiButton>
        </div>
      </div>

      <label class="mt-5 flex h-11 items-center gap-2 rounded-lg border border-[var(--border-soft)] bg-[var(--surface-input)] px-3 text-[var(--page-muted)]">
        <Search class="h-4 w-4 shrink-0" />
        <input
          :value="query"
          class="min-w-0 flex-1 bg-transparent text-sm text-[var(--page-text)] outline-none placeholder:text-[var(--page-soft)]"
          placeholder="搜索标题或摘要"
          @input="emit('search', ($event.target as HTMLInputElement).value)"
        />
      </label>

      <div class="mt-5 grid gap-2">
        <button
          v-for="note in notes"
          :key="note.id"
          type="button"
          class="note-row"
          :class="{ 'note-row-active': note.id === activeID }"
          @click="openNote(note)"
        >
          <span class="flex min-w-0 items-center gap-2">
            <Pin v-if="note.pinned" class="h-3.5 w-3.5 shrink-0 text-[var(--accent)]" />
            <FileText v-else class="h-3.5 w-3.5 shrink-0 text-[var(--page-soft)]" />
            <span class="truncate font-semibold">{{ note.title }}</span>
          </span>
          <span class="line-clamp-2 text-left text-xs leading-5 text-[var(--page-soft)]">{{ note.summary || '暂无摘要' }}</span>
          <span class="flex items-center justify-between gap-3 text-[11px] text-[var(--page-soft)]">
            <span class="flex min-w-0 items-center gap-2">
              <span v-if="user" class="inline-flex shrink-0 items-center gap-1 rounded-full border border-[var(--border-soft)] px-2 py-0.5">
                <Globe2 v-if="note.visibility === 'public'" class="h-3 w-3" />
                <Lock v-else class="h-3 w-3" />
                {{ visibilityText(note.visibility) }}
              </span>
              <span class="truncate">{{ note.tags.join(' / ') || '未标记' }}</span>
            </span>
            <span class="shrink-0">{{ formatDate(note.updatedAt) }}</span>
          </span>
        </button>

        <div v-if="!notes.length && !loading" class="rounded-lg border border-dashed border-[var(--border)] px-4 py-8 text-center text-sm text-[var(--page-soft)]">
          暂无文档
        </div>
        <div v-if="loading" class="flex items-center justify-center gap-2 py-8 text-sm text-[var(--page-muted)]">
          <Loader2 class="h-4 w-4 animate-spin" /> 读取中
        </div>
      </div>
    </div>

    <!-- 编辑区用 form 包住标题、标签和 Markdown 内容，提交统一触发保存。 -->
    <form class="notes-panel notes-editor" @submit.prevent="emit('save')">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="min-w-0">
          <p class="text-xs font-semibold uppercase tracking-[.16em] text-[var(--page-soft)]">
            {{ selected ? (displayPreview ? '预览文档' : '正在编辑') : (canEdit ? '新建文档' : '公开文档') }}
          </p>
          <h3 class="mt-1 truncate text-xl font-semibold text-[var(--page-text)]">
            {{ draft.title || '未命名文档' }}
          </h3>
        </div>
        <div class="flex flex-wrap gap-2">
          <UiButton v-if="canEdit && selected && isPreview" type="button" variant="outline" @click="mode = 'edit'">
            <Pencil class="h-4 w-4" /> 编辑
          </UiButton>
          <UiButton v-if="canEdit && selected && !isPreview" type="button" variant="outline" @click="mode = 'preview'">
            <Eye class="h-4 w-4" /> 预览
          </UiButton>
          <UiButton v-if="canEdit && !displayPreview" type="button" variant="outline" @click="patchDraft({ pinned: !draft.pinned })">
            <Pin class="h-4 w-4" /> {{ draft.pinned ? '取消置顶' : '置顶' }}
          </UiButton>
          <UiButton v-if="canEdit && !displayPreview" type="submit" :disabled="saving" variant="success">
            <Save class="h-4 w-4" /> {{ saving ? '保存中' : '保存' }}
          </UiButton>
          <UiButton v-if="canEdit && selected" type="button" variant="danger" @click="emit('delete')">
            <Trash2 class="h-4 w-4" /> 删除
          </UiButton>
        </div>
      </div>

      <div v-if="error" class="rounded-lg border border-[var(--danger-border)] bg-[var(--danger-bg)] px-4 py-3 text-sm text-[var(--danger-text)]">
        {{ error }}
      </div>

      <div v-if="displayPreview" class="note-preview-meta">
        <span v-if="draft.pinned">置顶</span>
        <span class="inline-flex items-center gap-1">
          <Globe2 v-if="draft.visibility === 'public'" class="h-3 w-3" />
          <Lock v-else class="h-3 w-3" />
          {{ visibilityText(draft.visibility) }}
        </span>
        <span>{{ draft.tags.join(' / ') || '未标记' }}</span>
      </div>

      <div v-if="canEdit && !displayPreview" class="grid gap-4">
        <label class="grid gap-2 text-sm text-[var(--page-muted)]">
          <span>标题</span>
          <input
            :value="draft.title"
            class="h-11 rounded-lg border border-[var(--border-soft)] bg-[var(--surface-input)] px-3 text-[15px] text-[var(--page-text)] outline-none transition placeholder:text-[var(--page-soft)] focus:border-[var(--focus)] focus:ring-4 focus:ring-[var(--focus-ring)]"
            placeholder="文档标题"
            @input="patchDraft({ title: ($event.target as HTMLInputElement).value })"
          />
        </label>
      </div>

      <div v-if="canEdit && !displayPreview" class="grid gap-4 md:grid-cols-[minmax(0,1fr)_220px]">
        <label class="grid gap-2 text-sm text-[var(--page-muted)]">
          <span>标签</span>
          <input
            :value="tagText"
            class="h-11 rounded-lg border border-[var(--border-soft)] bg-[var(--surface-input)] px-3 text-[15px] text-[var(--page-text)] outline-none transition placeholder:text-[var(--page-soft)] focus:border-[var(--focus)] focus:ring-4 focus:ring-[var(--focus-ring)]"
            placeholder="idea, work"
            @input="updateTags(($event.target as HTMLInputElement).value)"
          />
        </label>
        <label class="grid gap-2 text-sm text-[var(--page-muted)]">
          <span>权限范围</span>
          <select
            :value="draft.visibility"
            class="h-11 rounded-lg border border-[var(--border-soft)] bg-[var(--surface-input)] px-3 text-[15px] text-[var(--page-text)] outline-none transition focus:border-[var(--focus)] focus:ring-4 focus:ring-[var(--focus-ring)]"
            @change="patchDraft({ visibility: ($event.target as HTMLSelectElement).value as NoteInput['visibility'] })"
          >
            <option value="private">隐私</option>
            <option value="public">公开</option>
          </select>
        </label>
      </div>

      <!-- 预览和编辑复用同一份 draft，切换模式不会丢失未保存内容。 -->
      <div class="note-markdown-shell" :class="{ 'note-markdown-preview-shell': displayPreview }">
        <MdPreview
          v-if="displayPreview"
          :editor-id="`note-preview-${selected?.id || 'draft'}`"
          :model-value="draft.content || (canEdit ? '暂无内容' : '请选择左侧公开文档')"
          preview-theme="github"
          code-theme="github"
        />
        <MdEditor
          v-else
          v-model="contentModel"
          language="zh-CN"
          preview-theme="github"
          code-theme="github"
          placeholder="# 标题&#10;&#10;写下正文..."
          no-upload-img
        />
      </div>
    </form>
  </section>
</template>

<style scoped>
/* 工作区在桌面端保持左右分栏，便于一边浏览索引一边编辑正文。 */
.notes-workspace {
  display: grid;
  grid-template-columns: minmax(280px, 360px) minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.notes-panel {
  border: 1px solid var(--border);
  border-radius: 18px;
  background: var(--card-bg);
  box-shadow: 0 18px 60px var(--shadow);
}

.notes-sidebar {
  padding: 20px;
}

.notes-editor {
  display: flex;
  min-height: 680px;
  flex-direction: column;
  gap: 18px;
  padding: 22px;
}

.note-preview-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  color: var(--page-soft);
  font-size: 12px;
}

.note-preview-meta span {
  border: 1px solid var(--border-soft);
  border-radius: 999px;
  background: var(--surface);
  padding: 4px 10px;
}

.note-markdown-shell {
  min-height: 480px;
  flex: 1;
  overflow: hidden;
  border: 1px solid var(--border-soft);
  border-radius: 12px;
  background: var(--surface-input);
}

/* md-editor-v3 的内部节点需要 :deep 才能接入当前主题变量。 */
.note-markdown-shell :deep(.md-editor) {
  height: 100%;
  min-height: 480px;
  border-radius: 12px;
  background: var(--surface-input);
  color: var(--page-text);
}

.note-markdown-shell :deep(.md-editor-preview-wrapper) {
  min-height: 480px;
  background: var(--surface-input);
  padding: 20px;
}

.note-markdown-preview-shell {
  padding: 8px;
}

.note-markdown-preview-shell :deep(.md-editor-preview-wrapper) {
  padding: 28px 32px;
}

.note-markdown-shell :deep(.md-editor-preview) {
  color: var(--page-text);
}

.note-markdown-preview-shell :deep(.md-editor-preview) {
  max-width: 78ch;
  margin: 20px;
}

.note-row {
  display: grid;
  gap: 8px;
  width: 100%;
  min-height: 104px;
  border: 1px solid var(--border-soft);
  border-radius: 12px;
  background: var(--surface);
  padding: 12px;
  color: var(--page-text);
  transition: border-color .2s ease, background .2s ease, transform .2s ease;
}

.note-row:hover,
.note-row-active {
  border-color: var(--border-hover);
  background: var(--surface-hover);
}

.note-row:hover {
  transform: translateY(-1px);
}

@media (max-width: 860px) {
  .notes-workspace {
    grid-template-columns: 1fr;
  }

  .notes-editor {
    min-height: auto;
  }
}
</style>

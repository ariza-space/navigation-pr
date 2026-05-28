import { ref } from 'vue'

import {
  createNote,
  deleteNote as deleteNoteRequest,
  getNote,
  listNotes,
  syncNoteIndex,
  updateNote,
} from '@/lib/api'
import type { Note, NoteContent, NoteInput } from '@/types/api'

const emptyDraft: NoteInput = {
  title: '',
  content: '',
  tags: [],
  status: 'active',
  pinned: false,
}

export function useNotes() {
  const notes = ref<Note[]>([])
  const selected = ref<NoteContent | null>(null)
  const draft = ref<NoteInput>({ ...emptyDraft })
  const query = ref('')
  const loading = ref(false)
  const saving = ref(false)
  const syncing = ref(false)
  const error = ref('')

  function resetDraft() {
    selected.value = null
    draft.value = { ...emptyDraft }
  }

  async function loadNotes() {
    loading.value = true
    error.value = ''
    try {
      notes.value = await listNotes({ q: query.value })
    } finally {
      loading.value = false
    }
  }

  async function selectNote(note: Note) {
    loading.value = true
    error.value = ''
    try {
      const detail = await getNote(note.id)
      selected.value = detail
      draft.value = {
        title: detail.title,
        content: detail.content,
        tags: [...detail.tags],
        status: detail.status === 'deleted' ? 'active' : detail.status,
        pinned: detail.pinned,
      }
    } finally {
      loading.value = false
    }
  }

  async function saveDraft() {
    saving.value = true
    error.value = ''
    try {
      const saved = selected.value
        ? await updateNote(selected.value.id, draft.value)
        : await createNote(draft.value)
      selected.value = saved
      draft.value = {
        title: saved.title,
        content: saved.content,
        tags: [...saved.tags],
        status: saved.status,
        pinned: saved.pinned,
      }
      await loadNotes()
      const next = notes.value.find((note) => note.id === saved.id)
      if (next) selected.value = saved
      return saved
    } finally {
      saving.value = false
    }
  }

  async function removeSelected() {
    if (!selected.value) return
    const id = selected.value.id
    await deleteNoteRequest(id)
    resetDraft()
    await loadNotes()
  }

  async function syncIndex() {
    syncing.value = true
    error.value = ''
    try {
      const result = await syncNoteIndex()
      await loadNotes()
      if (selected.value) {
        const stillExists = notes.value.some((note) => note.id === selected.value?.id)
        if (!stillExists) resetDraft()
      }
      return result
    } finally {
      syncing.value = false
    }
  }

  return {
    notes,
    selected,
    draft,
    query,
    loading,
    saving,
    syncing,
    error,
    loadNotes,
    selectNote,
    saveDraft,
    removeSelected,
    resetDraft,
    syncIndex,
  }
}

package service

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"navigation/internal/domain"
)

const (
	maxNoteTitleLength   = 120
	maxNoteContentLength = 1024 * 1024
	maxNoteTags          = 20
	maxNoteTagLength     = 32
	maxNoteSummaryLength = 120
)

// NoteMetaStore 定义笔记元数据持久化能力。
type NoteMetaStore interface {
	ListNotes(status, query string, includePrivate bool) ([]domain.Note, error)
	GetNote(id string) (domain.Note, error)
	CreateNote(domain.Note) error
	UpdateNote(domain.Note) error
	SoftDeleteNote(id, deletedAt string) error
	RebuildNoteIndex([]domain.Note) error
}

// NoteContentStore 定义笔记 Markdown 文件读写能力。
type NoteContentStore interface {
	NewRelativePath(id string, now time.Time) string
	Write(relativePath, content string) error
	Read(relativePath string) (string, error)
	MoveToTrash(relativePath, id string) (string, error)
	ListMarkdownFiles() ([]domain.NoteFile, error)
}

// NoteService 封装笔记业务规则。
type NoteService struct {
	mu      sync.Mutex
	meta    NoteMetaStore
	content NoteContentStore
}

// NewNoteService 创建笔记服务。
func NewNoteService(meta NoteMetaStore, content NoteContentStore) *NoteService {
	return &NoteService{meta: meta, content: content}
}

// ListNotes 返回笔记列表，不读取 Markdown 全文。
func (s *NoteService) ListNotes(status, query string, includePrivate bool) ([]domain.Note, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		status = domain.NoteStatusActive
	}
	if !validNoteListStatus(status) {
		return nil, ValidationError{Message: "笔记状态不正确"}
	}
	if !includePrivate && status != domain.NoteStatusActive {
		return []domain.Note{}, nil
	}
	notes, err := s.meta.ListNotes(status, strings.TrimSpace(query), includePrivate)
	if err != nil {
		return nil, StoreError{Op: StoreOpRead, Err: err}
	}
	return notes, nil
}

// GetNote 返回笔记元数据和 Markdown 正文。
func (s *NoteService) GetNote(id string, includePrivate bool) (domain.NoteContent, error) {
	note, err := s.meta.GetNote(strings.TrimSpace(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.NoteContent{}, ErrNotFound
		}
		return domain.NoteContent{}, StoreError{Op: StoreOpRead, Err: err}
	}
	if note.Status == domain.NoteStatusDeleted || (!includePrivate && (note.Status != domain.NoteStatusActive || note.Visibility != domain.NoteVisibilityPublic)) {
		return domain.NoteContent{}, ErrNotFound
	}
	content, err := s.content.Read(note.FilePath)
	if err != nil {
		return domain.NoteContent{}, StoreError{Op: StoreOpRead, Err: err}
	}
	return domain.NoteContent{Note: note, Content: content}, nil
}

// CreateNote 创建笔记文件和元数据。
func (s *NoteService) CreateNote(input domain.NoteContent) (domain.NoteContent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nowTime := time.Now()
	now := nowTime.Format(time.RFC3339)
	input.ID = newNoteID()
	input.CreatedAt = now
	input.UpdatedAt = now
	input.DeletedAt = ""
	input.FilePath = s.content.NewRelativePath(input.ID, nowTime)
	if input.Status == "" {
		input.Status = domain.NoteStatusActive
	}
	if input.Visibility == "" {
		input.Visibility = domain.NoteVisibilityPrivate
	}
	normalizeNoteContent(&input)
	if err := validateNoteContent(input); err != nil {
		return domain.NoteContent{}, err
	}

	if err := s.content.Write(input.FilePath, input.Content); err != nil {
		return domain.NoteContent{}, StoreError{Op: StoreOpSave, Err: err}
	}
	if err := s.meta.CreateNote(input.Note); err != nil {
		return domain.NoteContent{}, StoreError{Op: StoreOpSave, Err: err}
	}
	return input, nil
}

// UpdateNote 更新笔记文件和元数据。
func (s *NoteService) UpdateNote(id string, input domain.NoteContent) (domain.NoteContent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.meta.GetNote(strings.TrimSpace(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.NoteContent{}, ErrNotFound
		}
		return domain.NoteContent{}, StoreError{Op: StoreOpRead, Err: err}
	}
	if existing.Status == domain.NoteStatusDeleted {
		return domain.NoteContent{}, ErrNotFound
	}

	input.ID = existing.ID
	input.FilePath = existing.FilePath
	input.CreatedAt = existing.CreatedAt
	input.DeletedAt = existing.DeletedAt
	input.UpdatedAt = time.Now().Format(time.RFC3339)
	if input.Status == "" {
		input.Status = existing.Status
	}
	if input.Visibility == "" {
		input.Visibility = existing.Visibility
	}
	if input.Visibility == "" {
		input.Visibility = domain.NoteVisibilityPrivate
	}
	normalizeNoteContent(&input)
	if err := validateNoteContent(input); err != nil {
		return domain.NoteContent{}, err
	}

	if err := s.content.Write(input.FilePath, input.Content); err != nil {
		return domain.NoteContent{}, StoreError{Op: StoreOpSave, Err: err}
	}
	if err := s.meta.UpdateNote(input.Note); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.NoteContent{}, ErrNotFound
		}
		return domain.NoteContent{}, StoreError{Op: StoreOpSave, Err: err}
	}
	return input, nil
}

// DeleteNote 将笔记软删除，不物理删除 Markdown 文件。
func (s *NoteService) DeleteNote(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id = strings.TrimSpace(id)
	if _, err := s.meta.GetNote(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return StoreError{Op: StoreOpRead, Err: err}
	}
	if err := s.meta.SoftDeleteNote(id, time.Now().Format(time.RFC3339)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return StoreError{Op: StoreOpSave, Err: err}
	}
	return nil
}

// SyncNoteIndex 扫描 Markdown 实体文件，并以文件为准重建数据库索引。
func (s *NoteService) SyncNoteIndex() (domain.NoteSyncResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	files, err := s.content.ListMarkdownFiles()
	if err != nil {
		return domain.NoteSyncResult{}, StoreError{Op: StoreOpRead, Err: err}
	}
	result := domain.NoteSyncResult{Scanned: len(files)}
	notes := make([]domain.Note, 0, len(files))
	usedIDs := map[string]bool{}
	for _, file := range files {
		if len([]byte(file.Content)) > maxNoteContentLength {
			result.Skipped++
			continue
		}
		note := domain.NoteContent{
			Note: domain.Note{
				ID:         noteIDFromPath(file.FilePath, usedIDs),
				FilePath:   file.FilePath,
				Status:     domain.NoteStatusActive,
				Visibility: domain.NoteVisibilityPrivate,
				CreatedAt:  file.UpdatedAt,
				UpdatedAt:  file.UpdatedAt,
				Tags:       []string{},
			},
			Content: file.Content,
		}
		normalizeNoteContent(&note)
		if note.Title == "" {
			note.Title = titleFromPath(file.FilePath)
		}
		if err := validateNoteContent(note); err != nil {
			result.Skipped++
			continue
		}
		notes = append(notes, note.Note)
		result.Indexed++
	}
	if err := s.meta.RebuildNoteIndex(notes); err != nil {
		return domain.NoteSyncResult{}, StoreError{Op: StoreOpSave, Err: err}
	}
	return result, nil
}

func normalizeNoteContent(input *domain.NoteContent) {
	input.Title = strings.TrimSpace(input.Title)
	input.Status = strings.TrimSpace(input.Status)
	input.Visibility = strings.TrimSpace(input.Visibility)
	input.Tags = normalizeTags(input.Tags)
	if input.Title == "" {
		input.Title = titleFromMarkdown(input.Content)
	}
	input.Summary = makeNoteSummary(input.Content)
}

func validateNoteContent(input domain.NoteContent) error {
	if input.Title == "" {
		return ValidationError{Message: "笔记标题不能为空"}
	}
	if utf8.RuneCountInString(input.Title) > maxNoteTitleLength {
		return ValidationError{Message: "笔记标题不能超过 120 个字符"}
	}
	if len([]byte(input.Content)) > maxNoteContentLength {
		return ValidationError{Message: "笔记正文不能超过 1MB"}
	}
	if !validNoteWriteStatus(input.Status) {
		return ValidationError{Message: "笔记状态不正确"}
	}
	if !validNoteVisibility(input.Visibility) {
		return ValidationError{Message: "笔记权限范围不正确"}
	}
	return nil
}

func validNoteWriteStatus(status string) bool {
	return status == domain.NoteStatusActive || status == domain.NoteStatusArchived
}

func validNoteListStatus(status string) bool {
	return validNoteWriteStatus(status) || status == domain.NoteStatusDeleted
}

func validNoteVisibility(visibility string) bool {
	return visibility == domain.NoteVisibilityPrivate || visibility == domain.NoteVisibilityPublic
}

func normalizeTags(tags []string) []string {
	normalized := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || seen[tag] {
			continue
		}
		if utf8.RuneCountInString(tag) > maxNoteTagLength {
			tag = string([]rune(tag)[:maxNoteTagLength])
		}
		seen[tag] = true
		normalized = append(normalized, tag)
		if len(normalized) == maxNoteTags {
			break
		}
	}
	return normalized
}

func titleFromMarkdown(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

var markdownTokenPattern = regexp.MustCompile(`[#>*_` + "`" + `\[\]()]`)

func makeNoteSummary(content string) string {
	text := markdownTokenPattern.ReplaceAllString(content, " ")
	text = strings.Join(strings.Fields(text), " ")
	runes := []rune(text)
	if len(runes) > maxNoteSummaryLength {
		return string(runes[:maxNoteSummaryLength])
	}
	return text
}

func newNoteID() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err == nil {
		return "note_" + hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("note_%d", time.Now().UnixNano())
}

var noteIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func noteIDFromPath(relativePath string, used map[string]bool) string {
	base := strings.TrimSuffix(filepath.Base(relativePath), filepath.Ext(relativePath))
	id := base
	if !noteIDPattern.MatchString(id) || id == "" {
		id = "note_" + shortPathHash(relativePath)
	}
	if !used[id] {
		used[id] = true
		return id
	}
	id = "note_" + shortPathHash(relativePath)
	for used[id] {
		id = "note_" + shortPathHash(relativePath+fmt.Sprintf("_%d", len(used)))
	}
	used[id] = true
	return id
}

func shortPathHash(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])[:12]
}

func titleFromPath(relativePath string) string {
	title := strings.TrimSuffix(filepath.Base(relativePath), filepath.Ext(relativePath))
	title = strings.TrimSpace(strings.ReplaceAll(title, "_", " "))
	if title == "" {
		return "未命名笔记"
	}
	runes := []rune(title)
	if len(runes) > maxNoteTitleLength {
		return string(runes[:maxNoteTitleLength])
	}
	return title
}

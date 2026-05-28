package storage

import (
	"database/sql"
	"encoding/json"
	"strings"

	"navigation/internal/domain"
)

// ListNotes 按状态和关键字读取笔记元数据。
func (s *SQLiteSiteStore) ListNotes(status, query string) ([]domain.Note, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		status = domain.NoteStatusActive
	}
	query = strings.TrimSpace(query)

	sqlQuery := `
		SELECT id, title, file_path, summary, tags, status, pinned, created_at, updated_at, deleted_at
		FROM notes
		WHERE status = ?
	`
	args := []any{status}
	if query != "" {
		sqlQuery += " AND (title LIKE ? OR summary LIKE ?)"
		like := "%" + query + "%"
		args = append(args, like, like)
	}
	sqlQuery += " ORDER BY pinned DESC, updated_at DESC"

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []domain.Note{}
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, rows.Err()
}

// GetNote 读取单条笔记元数据。
func (s *SQLiteSiteStore) GetNote(id string) (domain.Note, error) {
	row := s.db.QueryRow(`
		SELECT id, title, file_path, summary, tags, status, pinned, created_at, updated_at, deleted_at
		FROM notes
		WHERE id = ?
	`, id)
	return scanNote(row)
}

// CreateNote 新增笔记元数据。
func (s *SQLiteSiteStore) CreateNote(note domain.Note) error {
	tags, err := json.Marshal(note.Tags)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO notes (
			id, title, file_path, summary, tags, status, pinned, created_at, updated_at, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, note.ID, note.Title, note.FilePath, note.Summary, string(tags), note.Status, boolToInt(note.Pinned), note.CreatedAt, note.UpdatedAt, note.DeletedAt)
	return err
}

// UpdateNote 更新笔记元数据。
func (s *SQLiteSiteStore) UpdateNote(note domain.Note) error {
	tags, err := json.Marshal(note.Tags)
	if err != nil {
		return err
	}
	result, err := s.db.Exec(`
		UPDATE notes
		SET title = ?, file_path = ?, summary = ?, tags = ?, status = ?, pinned = ?, created_at = ?, updated_at = ?, deleted_at = ?
		WHERE id = ?
	`, note.Title, note.FilePath, note.Summary, string(tags), note.Status, boolToInt(note.Pinned), note.CreatedAt, note.UpdatedAt, note.DeletedAt, note.ID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SoftDeleteNote 将笔记标记为已删除。
func (s *SQLiteSiteStore) SoftDeleteNote(id, deletedAt string) error {
	result, err := s.db.Exec(`
		UPDATE notes
		SET status = ?, deleted_at = ?, updated_at = ?
		WHERE id = ?
	`, domain.NoteStatusDeleted, deletedAt, deletedAt, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RebuildNoteIndex 用扫描到的 Markdown 文件重建 notes 元数据索引。
func (s *SQLiteSiteStore) RebuildNoteIndex(notes []domain.Note) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	existing := map[string]domain.Note{}
	rows, err := tx.Query(`
		SELECT id, title, file_path, summary, tags, status, pinned, created_at, updated_at, deleted_at
		FROM notes
	`)
	if err != nil {
		return err
	}
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			rows.Close()
			return err
		}
		existing[note.ID] = note
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM notes`); err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO notes (
			id, title, file_path, summary, tags, status, pinned, created_at, updated_at, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, note := range notes {
		if old, ok := existing[note.ID]; ok {
			note.Pinned = old.Pinned
			if old.CreatedAt != "" {
				note.CreatedAt = old.CreatedAt
			}
		}
		tags, err := json.Marshal(note.Tags)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(note.ID, note.Title, note.FilePath, note.Summary, string(tags), note.Status, boolToInt(note.Pinned), note.CreatedAt, note.UpdatedAt, note.DeletedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type noteScanner interface {
	Scan(dest ...any) error
}

func scanNote(scanner noteScanner) (domain.Note, error) {
	var note domain.Note
	var tags string
	var pinned int
	if err := scanner.Scan(
		&note.ID,
		&note.Title,
		&note.FilePath,
		&note.Summary,
		&tags,
		&note.Status,
		&pinned,
		&note.CreatedAt,
		&note.UpdatedAt,
		&note.DeletedAt,
	); err != nil {
		return domain.Note{}, err
	}
	if strings.TrimSpace(tags) == "" {
		tags = "[]"
	}
	if err := json.Unmarshal([]byte(tags), &note.Tags); err != nil {
		return domain.Note{}, err
	}
	note.Pinned = pinned != 0
	return note, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

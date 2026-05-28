package domain

const (
	// NoteStatusActive 表示可见的普通笔记。
	NoteStatusActive = "active"
	// NoteStatusArchived 表示已归档但未删除的笔记。
	NoteStatusArchived = "archived"
	// NoteStatusDeleted 表示已软删除的笔记。
	NoteStatusDeleted = "deleted"
)

// Note 保存笔记列表和索引所需的元数据。
type Note struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	FilePath  string   `json:"filePath"`
	Summary   string   `json:"summary"`
	Tags      []string `json:"tags"`
	Status    string   `json:"status"`
	Pinned    bool     `json:"pinned"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	DeletedAt string   `json:"deletedAt,omitempty"`
}

// NoteContent 在笔记元数据基础上附带 Markdown 正文。
type NoteContent struct {
	Note
	Content string `json:"content"`
}

// NoteFile 表示从 Markdown 实体文件扫描出的正文。
type NoteFile struct {
	FilePath  string `json:"filePath"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updatedAt"`
}

// NoteSyncResult 表示一次笔记文件索引重建的结果。
type NoteSyncResult struct {
	Scanned int `json:"scanned"`
	Indexed int `json:"indexed"`
	Skipped int `json:"skipped"`
}

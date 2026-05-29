package httptransport

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"navigation/internal/domain"
	"navigation/internal/service"
)

type testStore struct {
	sites    []domain.Site
	notes    []domain.Note
	contents map[string]string
	user     domain.User
	settings domain.AppSettings
}

func (s *testStore) ListSites() ([]domain.Site, error) {
	sites := make([]domain.Site, len(s.sites))
	copy(sites, s.sites)
	return sites, nil
}

func (s *testStore) SaveSites(sites []domain.Site) error {
	s.sites = make([]domain.Site, len(sites))
	copy(s.sites, sites)
	return nil
}

func (s *testStore) GetUser() (domain.User, error) {
	if s.user.Username == "" {
		return domain.User{}, sql.ErrNoRows
	}
	return s.user, nil
}

func (s *testStore) SaveUser(user domain.User) error {
	s.user = user
	return nil
}

func (s *testStore) GetSettings() (domain.AppSettings, error) {
	return s.settings, nil
}

func (s *testStore) SaveSettings(settings domain.AppSettings) error {
	s.settings = settings
	return nil
}

func (s *testStore) ListNotes(status, query string, includePrivate bool) ([]domain.Note, error) {
	notes := []domain.Note{}
	for _, note := range s.notes {
		if status != "" && note.Status != status {
			continue
		}
		if !includePrivate && note.Visibility != domain.NoteVisibilityPublic {
			continue
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (s *testStore) GetNote(id string) (domain.Note, error) {
	for _, note := range s.notes {
		if note.ID == id {
			return note, nil
		}
	}
	return domain.Note{}, sql.ErrNoRows
}

func (s *testStore) CreateNote(note domain.Note) error {
	s.notes = append(s.notes, note)
	return nil
}

func (s *testStore) UpdateNote(note domain.Note) error {
	for i := range s.notes {
		if s.notes[i].ID == note.ID {
			s.notes[i] = note
			return nil
		}
	}
	return sql.ErrNoRows
}

func (s *testStore) SoftDeleteNote(id, deletedAt string) error {
	for i := range s.notes {
		if s.notes[i].ID == id {
			s.notes[i].Status = domain.NoteStatusDeleted
			s.notes[i].DeletedAt = deletedAt
			s.notes[i].UpdatedAt = deletedAt
			return nil
		}
	}
	return sql.ErrNoRows
}

func (s *testStore) RebuildNoteIndex(notes []domain.Note) error {
	s.notes = append([]domain.Note(nil), notes...)
	return nil
}

func (s *testStore) NewRelativePath(id string, now time.Time) string {
	return "notes/2026/05/" + id + ".md"
}

func (s *testStore) Write(relativePath, content string) error {
	if s.contents == nil {
		s.contents = map[string]string{}
	}
	s.contents[relativePath] = content
	return nil
}

func (s *testStore) Read(relativePath string) (string, error) {
	content, ok := s.contents[relativePath]
	if !ok {
		return "", sql.ErrNoRows
	}
	return content, nil
}

func (s *testStore) MoveToTrash(relativePath, id string) (string, error) {
	return relativePath, nil
}

func (s *testStore) ListMarkdownFiles() ([]domain.NoteFile, error) {
	files := []domain.NoteFile{}
	for path, content := range s.contents {
		files = append(files, domain.NoteFile{
			FilePath:  path,
			Content:   content,
			UpdatedAt: "2026-05-28T00:00:00Z",
		})
	}
	return files, nil
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	store := &testStore{
		sites: []domain.Site{{ID: "site-1", Name: "Go", URL: "https://go.dev", Category: "Dev", Sort: 1}},
	}
	auth, err := service.NewAuthService(store)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	static := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("index")}}
	return NewHandler(service.NewSiteService(store), auth, service.NewNoteService(store, store), static).Routes()
}

func TestAnonymousReadEndpoints(t *testing.T) {
	handler := newTestHandler(t)
	paths := []string{"/api/sites", "/api/categories", "/api/stats", "/api/settings", "/api/category-stats"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
		})
	}
}

func TestAnonymousCanReadOnlyPublicNotes(t *testing.T) {
	store := &testStore{
		sites: []domain.Site{{ID: "site-1", Name: "Go", URL: "https://go.dev", Category: "Dev", Sort: 1}},
		notes: []domain.Note{
			{
				ID:         "public",
				Title:      "公开文档",
				FilePath:   "notes/2026/05/public.md",
				Status:     domain.NoteStatusActive,
				Visibility: domain.NoteVisibilityPublic,
				CreatedAt:  "2026-05-28T00:00:00Z",
				UpdatedAt:  "2026-05-28T00:00:00Z",
			},
			{
				ID:         "private",
				Title:      "隐私文档",
				FilePath:   "notes/2026/05/private.md",
				Status:     domain.NoteStatusActive,
				Visibility: domain.NoteVisibilityPrivate,
				CreatedAt:  "2026-05-28T00:00:00Z",
				UpdatedAt:  "2026-05-28T00:00:00Z",
			},
		},
		contents: map[string]string{
			"notes/2026/05/public.md":  "公开正文",
			"notes/2026/05/private.md": "隐私正文",
		},
	}
	auth, err := service.NewAuthService(store)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	static := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("index")}}
	handler := NewHandler(service.NewSiteService(store), auth, service.NewNoteService(store, store), static).Routes()

	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, httptest.NewRequest(http.MethodGet, "/api/notes", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRec.Code, http.StatusOK)
	}
	var notes []domain.Note
	if err := json.NewDecoder(listRec.Body).Decode(&notes); err != nil {
		t.Fatalf("decode notes: %v", err)
	}
	if len(notes) != 1 || notes[0].ID != "public" {
		t.Fatalf("notes = %#v, want only public note", notes)
	}

	publicRec := httptest.NewRecorder()
	handler.ServeHTTP(publicRec, httptest.NewRequest(http.MethodGet, "/api/notes/public", nil))
	if publicRec.Code != http.StatusOK {
		t.Fatalf("public status = %d, want %d", publicRec.Code, http.StatusOK)
	}

	privateRec := httptest.NewRecorder()
	handler.ServeHTTP(privateRec, httptest.NewRequest(http.MethodGet, "/api/notes/private", nil))
	if privateRec.Code != http.StatusNotFound {
		t.Fatalf("private status = %d, want %d", privateRec.Code, http.StatusNotFound)
	}
}

func TestAnonymousWriteEndpointsRequireLogin(t *testing.T) {
	handler := newTestHandler(t)
	requests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPost, path: "/api/sites", body: `{}`},
		{method: http.MethodPut, path: "/api/sites/site-1", body: `{}`},
		{method: http.MethodDelete, path: "/api/sites/site-1"},
		{method: http.MethodPut, path: "/api/categories/Dev", body: `{"name":"Docs"}`},
		{method: http.MethodDelete, path: "/api/categories/Dev"},
		{method: http.MethodPut, path: "/api/settings", body: `{}`},
		{method: http.MethodPost, path: "/api/notes", body: `{}`},
		{method: http.MethodPost, path: "/api/notes/sync"},
		{method: http.MethodPut, path: "/api/notes/note-1", body: `{}`},
		{method: http.MethodDelete, path: "/api/notes/note-1"},
	}

	for _, request := range requests {
		t.Run(request.method+" "+request.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(request.method, request.path, bytes.NewBufferString(request.body)))
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAuthenticatedNoteSync(t *testing.T) {
	store := &testStore{
		sites:    []domain.Site{{ID: "site-1", Name: "Go", URL: "https://go.dev", Category: "Dev", Sort: 1}},
		contents: map[string]string{"notes/2026/05/note_1.md": "# 同步笔记\n\n正文"},
	}
	auth, err := service.NewAuthService(store)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	credential, ok := auth.InitialCredential()
	if !ok {
		t.Fatal("InitialCredential() missing")
	}
	token, _, err := auth.Login(credential.Username, credential.Password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	static := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("index")}}
	handler := NewHandler(service.NewSiteService(store), auth, service.NewNoteService(store, store), static).Routes()

	req := httptest.NewRequest(http.MethodPost, "/api/notes/sync", nil)
	req.AddCookie(&http.Cookie{Name: "navigation_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("sync status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var result domain.NoteSyncResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode sync result: %v", err)
	}
	if result.Indexed != 1 || len(store.notes) != 1 || store.notes[0].Title != "同步笔记" {
		t.Fatalf("result = %#v notes = %#v, want one synced note", result, store.notes)
	}
}

func TestAuthenticatedNoteLifecycle(t *testing.T) {
	store := &testStore{
		sites:    []domain.Site{{ID: "site-1", Name: "Go", URL: "https://go.dev", Category: "Dev", Sort: 1}},
		contents: map[string]string{},
	}
	auth, err := service.NewAuthService(store)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	credential, ok := auth.InitialCredential()
	if !ok {
		t.Fatal("InitialCredential() missing")
	}
	token, _, err := auth.Login(credential.Username, credential.Password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	static := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("index")}}
	handler := NewHandler(service.NewSiteService(store), auth, service.NewNoteService(store, store), static).Routes()

	createReq := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewBufferString(`{"title":"新笔记","content":"# 新笔记\n\n正文","tags":[" idea ","idea"],"pinned":true}`))
	createReq.AddCookie(&http.Cookie{Name: "navigation_session", Value: token})
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d, body = %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}

	var created domain.NoteContent
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created note: %v", err)
	}
	if created.ID == "" || created.FilePath == "" {
		t.Fatalf("created note missing id or path: %#v", created)
	}
	if len(created.Tags) != 1 || created.Tags[0] != "idea" {
		t.Fatalf("tags = %#v, want normalized idea", created.Tags)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/notes/"+created.ID, nil)
	getReq.AddCookie(&http.Cookie{Name: "navigation_session", Value: token})
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getRec.Code, http.StatusOK)
	}
}

func TestCreateInvalidNoteReturnsBadRequest(t *testing.T) {
	store := &testStore{
		sites:    []domain.Site{{ID: "site-1", Name: "Go", URL: "https://go.dev", Category: "Dev", Sort: 1}},
		contents: map[string]string{},
	}
	auth, err := service.NewAuthService(store)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	credential, ok := auth.InitialCredential()
	if !ok {
		t.Fatal("InitialCredential() missing")
	}
	token, _, err := auth.Login(credential.Username, credential.Password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	static := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("index")}}
	handler := NewHandler(service.NewSiteService(store), auth, service.NewNoteService(store, store), static).Routes()

	req := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewBufferString(`{"content":"正文"}`))
	req.AddCookie(&http.Cookie{Name: "navigation_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLoginRateLimit(t *testing.T) {
	handler := newTestHandler(t)
	for i := 0; i < loginFailureLimit; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"wrong"}`))
		req.RemoteAddr = "192.0.2.10:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("attempt %d status = %d, want %d", i+1, rec.Code, http.StatusBadRequest)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"wrong"}`))
	req.RemoteAddr = "192.0.2.10:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestWriteOriginMustMatchRequestOrigin(t *testing.T) {
	store := &testStore{
		sites:    []domain.Site{{ID: "site-1", Name: "Go", URL: "https://go.dev", Category: "Dev", Sort: 1}},
		contents: map[string]string{},
	}
	auth, err := service.NewAuthService(store)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	credential, ok := auth.InitialCredential()
	if !ok {
		t.Fatal("InitialCredential() missing")
	}
	token, _, err := auth.Login(credential.Username, credential.Password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	static := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("index")}}
	handler := NewHandler(service.NewSiteService(store), auth, service.NewNoteService(store, store), static).Routes()

	req := httptest.NewRequest(http.MethodPost, "http://navigation.local/api/notes", bytes.NewBufferString(`{"title":"x","content":"x"}`))
	req.Header.Set("Origin", "http://evil.example")
	req.AddCookie(&http.Cookie{Name: "navigation_session", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestSecureCookieOption(t *testing.T) {
	store := &testStore{
		sites: []domain.Site{{ID: "site-1", Name: "Go", URL: "https://go.dev", Category: "Dev", Sort: 1}},
	}
	auth, err := service.NewAuthService(store)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	credential, ok := auth.InitialCredential()
	if !ok {
		t.Fatal("InitialCredential() missing")
	}
	static := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("index")}}
	handler := NewHandler(service.NewSiteService(store), auth, service.NewNoteService(store, store), static, WithSecureCookies(true)).Routes()

	body := `{"username":"` + credential.Username + `","password":"` + credential.Password + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login did not set a cookie")
	}
	if !cookies[0].Secure {
		t.Fatal("session cookie Secure flag is false")
	}
}

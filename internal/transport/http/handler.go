package httptransport

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"navigation/internal/domain"
	"navigation/internal/service"
)

const (
	loginFailureLimit  = 5
	loginFailureWindow = time.Minute
)

// Handler 负责 HTTP 路由、请求解析和响应编码。
type Handler struct {
	service       *service.SiteService
	auth          *service.AuthService
	notes         *service.NoteService
	static        fs.FS
	secureCookies bool
	loginLimiter  *loginLimiter
}

// HandlerOption 配置 HTTP 处理器行为。
type HandlerOption func(*Handler)

// WithSecureCookies 控制会话 Cookie 是否仅允许 HTTPS 传输。
func WithSecureCookies(secure bool) HandlerOption {
	return func(h *Handler) {
		h.secureCookies = secure
	}
}

// NewHandler 创建 HTTP 处理器。
func NewHandler(service *service.SiteService, auth *service.AuthService, notes *service.NoteService, static fs.FS, options ...HandlerOption) *Handler {
	if distFS, err := fs.Sub(static, "web/dist"); err == nil {
		static = distFS
	}
	handler := &Handler{
		service:      service,
		auth:         auth,
		notes:        notes,
		static:       static,
		loginLimiter: newLoginLimiter(loginFailureLimit, loginFailureWindow),
	}
	for _, option := range options {
		option(handler)
	}
	return handler
}

// Routes 注册页面和 API 路由。
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", h.handleLogin)
	mux.HandleFunc("/api/session", h.handleSession)
	mux.HandleFunc("/api/logout", h.requireAuth(h.handleLogout))
	mux.HandleFunc("/api/account", h.requireAuth(h.handleAccount))
	mux.HandleFunc("/api/settings", h.handleSettings)
	mux.HandleFunc("/api/sites", h.handleSites)
	mux.HandleFunc("/api/sites/", h.requireAuth(h.handleSiteByID))
	mux.HandleFunc("/api/notes", h.requireAuth(h.handleNotes))
	mux.HandleFunc("/api/notes/sync", h.requireAuth(h.handleNoteSync))
	mux.HandleFunc("/api/notes/", h.requireAuth(h.handleNoteByID))
	mux.HandleFunc("/api/categories", h.handleCategories)
	mux.HandleFunc("/api/categories/", h.requireAuth(h.handleCategoryByName))
	mux.HandleFunc("/api/category-stats", h.handleCategoryStats)
	mux.HandleFunc("/api/stats", h.handleStats)
	mux.Handle("/assets/", http.FileServerFS(h.static))
	mux.HandleFunc("/", h.serveIndex)
	return mux
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.ensureAuth(w, r) {
			return
		}
		next(w, r)
	}
}

func (h *Handler) ensureAuth(w http.ResponseWriter, r *http.Request) bool {
	if !h.ensureWriteOrigin(w, r) {
		return false
	}
	cookie, err := r.Cookie("navigation_session")
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "请先登录")
		return false
	}
	if _, ok := h.auth.UserBySession(cookie.Value); !ok {
		h.clearSessionCookie(w)
		writeError(w, http.StatusUnauthorized, "登录已失效")
		return false
	}
	return true
}

func (h *Handler) ensureWriteOrigin(w http.ResponseWriter, r *http.Request) bool {
	if !isWriteMethod(r.Method) {
		return true
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		origin = strings.TrimSpace(r.Header.Get("Referer"))
	}
	if origin == "" {
		return true
	}
	originURL, err := url.Parse(origin)
	if err != nil || originURL.Scheme == "" || originURL.Host == "" {
		writeError(w, http.StatusForbidden, "请求来源不正确")
		return false
	}
	if !sameOrigin(r, originURL) {
		writeError(w, http.StatusForbidden, "请求来源不正确")
		return false
	}
	return true
}

func (h *Handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}
	http.ServeFileFS(w, r, h.static, "index.html")
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}
	clientKey := loginClientKey(r)
	if !h.loginLimiter.Allow(clientKey) {
		writeError(w, http.StatusTooManyRequests, "登录失败次数过多，请稍后再试")
		return
	}
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "请求数据格式不正确")
		return
	}
	token, user, err := h.auth.Login(input.Username, input.Password)
	if err != nil {
		h.loginLimiter.RecordFailure(clientKey)
		h.writeServiceError(w, err, "账号不存在")
		return
	}
	h.loginLimiter.RecordSuccess(clientKey)
	h.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]string{"username": user.Username})
}

func (h *Handler) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}
	cookie, err := r.Cookie("navigation_session")
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "请先登录")
		return
	}
	user, ok := h.auth.UserBySession(cookie.Value)
	if !ok {
		h.clearSessionCookie(w)
		writeError(w, http.StatusUnauthorized, "登录已失效")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"username": user.Username})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}
	if cookie, err := r.Cookie("navigation_session"); err == nil {
		h.auth.Logout(cookie.Value)
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}
	var input struct {
		Username        string `json:"username"`
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "请求数据格式不正确")
		return
	}
	user, passwordChanged, err := h.auth.UpdateAccount(input.Username, input.CurrentPassword, input.NewPassword)
	if err != nil {
		h.writeServiceError(w, err, "账号不存在")
		return
	}
	if passwordChanged {
		token, _, err := h.auth.Login(user.Username, input.NewPassword)
		if err != nil {
			h.clearSessionCookie(w)
			h.writeServiceError(w, err, "账号不存在")
			return
		}
		h.setSessionCookie(w, token)
	}
	writeJSON(w, http.StatusOK, map[string]string{"username": user.Username})
}

func (h *Handler) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := h.auth.Settings()
		if err != nil {
			h.writeServiceError(w, err, "没有找到设置")
			return
		}
		writeJSON(w, http.StatusOK, settings)
	case http.MethodPut:
		if !h.ensureAuth(w, r) {
			return
		}
		var input domain.AppSettings
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "请求数据格式不正确")
			return
		}
		settings, err := h.auth.UpdateSettings(input)
		if err != nil {
			h.writeServiceError(w, err, "没有找到设置")
			return
		}
		writeJSON(w, http.StatusOK, settings)
	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func (h *Handler) handleSites(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listSites(w, r)
	case http.MethodPost:
		if !h.ensureAuth(w, r) {
			return
		}
		h.createSite(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func (h *Handler) handleSiteByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/sites/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少站点 ID")
		return
	}

	switch r.Method {
	case http.MethodPut:
		h.updateSite(w, r, id)
	case http.MethodDelete:
		h.deleteSite(w, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func (h *Handler) listSites(w http.ResponseWriter, r *http.Request) {
	sites, err := h.service.ListSites(r.URL.Query().Get("category"), r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}
	writeJSON(w, http.StatusOK, sites)
}

func (h *Handler) createSite(w http.ResponseWriter, r *http.Request) {
	var input domain.Site
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "请求数据格式不正确")
		return
	}

	site, err := h.service.CreateSite(input)
	if err != nil {
		h.writeServiceError(w, err, "没有找到这个站点")
		return
	}
	writeJSON(w, http.StatusCreated, site)
}

func (h *Handler) updateSite(w http.ResponseWriter, r *http.Request, id string) {
	var input domain.Site
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "请求数据格式不正确")
		return
	}

	site, err := h.service.UpdateSite(id, input)
	if err != nil {
		h.writeServiceError(w, err, "没有找到这个站点")
		return
	}
	writeJSON(w, http.StatusOK, site)
}

func (h *Handler) deleteSite(w http.ResponseWriter, id string) {
	if err := h.service.DeleteSite(id); err != nil {
		h.writeServiceError(w, err, "没有找到这个站点")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleNotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		notes, err := h.notes.ListNotes(r.URL.Query().Get("status"), r.URL.Query().Get("q"))
		if err != nil {
			h.writeNoteServiceError(w, err, "没有找到这篇笔记")
			return
		}
		writeJSON(w, http.StatusOK, notes)
	case http.MethodPost:
		var input domain.NoteContent
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "请求数据格式不正确")
			return
		}
		note, err := h.notes.CreateNote(input)
		if err != nil {
			h.writeNoteServiceError(w, err, "没有找到这篇笔记")
			return
		}
		writeJSON(w, http.StatusCreated, note)
	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func (h *Handler) handleNoteByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/notes/"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少笔记 ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		note, err := h.notes.GetNote(id)
		if err != nil {
			h.writeNoteServiceError(w, err, "没有找到这篇笔记")
			return
		}
		writeJSON(w, http.StatusOK, note)
	case http.MethodPut:
		var input domain.NoteContent
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "请求数据格式不正确")
			return
		}
		note, err := h.notes.UpdateNote(id, input)
		if err != nil {
			h.writeNoteServiceError(w, err, "没有找到这篇笔记")
			return
		}
		writeJSON(w, http.StatusOK, note)
	case http.MethodDelete:
		if err := h.notes.DeleteNote(id); err != nil {
			h.writeNoteServiceError(w, err, "没有找到这篇笔记")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func (h *Handler) handleNoteSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}
	result, err := h.notes.SyncNoteIndex()
	if err != nil {
		h.writeNoteServiceError(w, err, "同步笔记失败")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) handleCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	categories, err := h.service.ListCategories()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (h *Handler) handleCategoryByName(w http.ResponseWriter, r *http.Request) {
	name, err := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/api/categories/"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "分类名称不正确")
		return
	}
	name = strings.TrimSpace(name)
	if name == "" || name == "全部" {
		writeError(w, http.StatusBadRequest, "分类名称不正确")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		updated, err := h.service.DeleteCategory(name)
		if err != nil {
			h.writeServiceError(w, err, "没有找到这个分类")
			return
		}
		writeJSON(w, http.StatusOK, map[string]int{"uncategorizedSites": updated})
	case http.MethodPut:
		var input struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "请求数据格式不正确")
			return
		}
		updated, err := h.service.RenameCategory(name, input.Name)
		if err != nil {
			h.writeServiceError(w, err, "没有找到这个分类")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"name": strings.TrimSpace(input.Name), "renamedSites": updated})
	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func (h *Handler) handleCategoryStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	categories, err := h.service.CategoryStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	stats, err := h.service.Stats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error, notFoundMessage string) {
	var validationErr service.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, notFoundMessage)
		return
	}

	var storeErr service.StoreError
	if errors.As(err, &storeErr) {
		if storeErr.Op == service.StoreOpRead {
			writeError(w, http.StatusInternalServerError, "读取站点数据失败")
			return
		}
		writeError(w, http.StatusInternalServerError, "保存站点数据失败")
		return
	}
	writeError(w, http.StatusInternalServerError, "保存站点数据失败")
}

func (h *Handler) writeNoteServiceError(w http.ResponseWriter, err error, notFoundMessage string) {
	var validationErr service.ValidationError
	if errors.As(err, &validationErr) {
		writeError(w, http.StatusBadRequest, validationErr.Error())
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, notFoundMessage)
		return
	}

	var storeErr service.StoreError
	if errors.As(err, &storeErr) {
		if storeErr.Op == service.StoreOpRead {
			writeError(w, http.StatusInternalServerError, "读取笔记数据失败")
			return
		}
		writeError(w, http.StatusInternalServerError, "保存笔记数据失败")
		return
	}
	writeError(w, http.StatusInternalServerError, "保存笔记数据失败")
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "navigation_session",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "navigation_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func isWriteMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

func sameOrigin(r *http.Request, origin *url.URL) bool {
	expectedScheme := requestScheme(r)
	return strings.EqualFold(origin.Scheme, expectedScheme) && strings.EqualFold(origin.Host, r.Host)
}

func requestScheme(r *http.Request) string {
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		return strings.ToLower(strings.TrimSpace(strings.Split(proto, ",")[0]))
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func loginClientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}

type loginLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	failures map[string][]time.Time
}

func newLoginLimiter(limit int, window time.Duration) *loginLimiter {
	return &loginLimiter{limit: limit, window: window, failures: map[string][]time.Time{}}
}

func (l *loginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pruneLocked(key, time.Now())
	return len(l.failures[key]) < l.limit
}

func (l *loginLimiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.pruneLocked(key, now)
	l.failures[key] = append(l.failures[key], now)
}

func (l *loginLimiter) RecordSuccess(key string) {
	l.mu.Lock()
	delete(l.failures, key)
	l.mu.Unlock()
}

func (l *loginLimiter) pruneLocked(key string, now time.Time) {
	cutoff := now.Add(-l.window)
	failures := l.failures[key]
	keepFrom := 0
	for keepFrom < len(failures) && failures[keepFrom].Before(cutoff) {
		keepFrom++
	}
	if keepFrom >= len(failures) {
		delete(l.failures, key)
		return
	}
	l.failures[key] = failures[keepFrom:]
}

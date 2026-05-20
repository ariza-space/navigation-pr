package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	dataFile    = "data/sites.json"
	defaultPort = 8080
	defaultGlow = "rgba(96,165,250,.45)"
)

type Site struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Category    string `json:"category"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
	Glow        string `json:"glow"`
	Sort        int    `json:"sort"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type Stats struct {
	SiteCount     int    `json:"siteCount"`
	CategoryCount int    `json:"categoryCount"`
	Coverage      string `json:"coverage"`
}

type CategoryStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type app struct {
	mu       sync.Mutex
	dataPath string
}

func main() {
	port := flag.Int("port", defaultPort, "HTTP server port")
	flag.Parse()
	if *port < 1 || *port > 65535 {
		log.Fatalf("端口必须在 1 到 65535 之间: %d", *port)
	}

	a := &app{dataPath: dataFile}
	if err := a.ensureDataFile(); err != nil {
		log.Fatalf("初始化数据文件失败: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/sites", a.handleSites)
	mux.HandleFunc("/api/sites/", a.handleSiteByID)
	mux.HandleFunc("/api/categories", a.handleCategories)
	mux.HandleFunc("/api/categories/", a.handleCategoryByName)
	mux.HandleFunc("/api/category-stats", a.handleCategoryStats)
	mux.HandleFunc("/api/stats", a.handleStats)
	mux.HandleFunc("/", serveIndex)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("导航站已启动: http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func (a *app) handleSites(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.listSites(w, r)
	case http.MethodPost:
		a.createSite(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func (a *app) handleSiteByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/sites/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少站点 ID")
		return
	}

	switch r.Method {
	case http.MethodPut:
		a.updateSite(w, r, id)
	case http.MethodDelete:
		a.deleteSite(w, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
	}
}

func (a *app) listSites(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	sites, err := a.loadSites()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}

	filtered := make([]Site, 0, len(sites))
	for _, site := range sites {
		if category != "" && category != "全部" && site.Category != category {
			continue
		}
		if query != "" {
			haystack := strings.ToLower(site.Name + " " + site.Description + " " + site.Category)
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		filtered = append(filtered, site)
	}

	sortSites(filtered)
	writeJSON(w, http.StatusOK, filtered)
}

func (a *app) createSite(w http.ResponseWriter, r *http.Request) {
	var input Site
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "请求数据格式不正确")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	sites, err := a.loadSitesLocked()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}

	now := time.Now().Format(time.RFC3339)
	input.ID = newID()
	input.CreatedAt = now
	input.UpdatedAt = now
	normalizeSite(&input, nextSort(sites))

	if err := validateSite(input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sites = append(sites, input)
	sortSites(sites)
	if err := a.saveSitesLocked(sites); err != nil {
		writeError(w, http.StatusInternalServerError, "保存站点数据失败")
		return
	}

	writeJSON(w, http.StatusCreated, input)
}

func (a *app) updateSite(w http.ResponseWriter, r *http.Request, id string) {
	var input Site
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "请求数据格式不正确")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	sites, err := a.loadSitesLocked()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}

	for i, site := range sites {
		if site.ID != id {
			continue
		}
		input.ID = site.ID
		input.CreatedAt = site.CreatedAt
		input.UpdatedAt = time.Now().Format(time.RFC3339)
		normalizeSite(&input, site.Sort)
		if err := validateSite(input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		sites[i] = input
		sortSites(sites)
		if err := a.saveSitesLocked(sites); err != nil {
			writeError(w, http.StatusInternalServerError, "保存站点数据失败")
			return
		}
		writeJSON(w, http.StatusOK, input)
		return
	}

	writeError(w, http.StatusNotFound, "没有找到这个站点")
}

func (a *app) deleteSite(w http.ResponseWriter, id string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	sites, err := a.loadSitesLocked()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}

	for i, site := range sites {
		if site.ID != id {
			continue
		}
		sites = append(sites[:i], sites[i+1:]...)
		if err := a.saveSitesLocked(sites); err != nil {
			writeError(w, http.StatusInternalServerError, "保存站点数据失败")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeError(w, http.StatusNotFound, "没有找到这个站点")
}

func (a *app) handleCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	sites, err := a.loadSites()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}

	seen := map[string]bool{}
	categories := []string{"全部"}
	for _, site := range sites {
		if site.Category == "" || seen[site.Category] {
			continue
		}
		seen[site.Category] = true
		categories = append(categories, site.Category)
	}
	writeJSON(w, http.StatusOK, categories)
}

func (a *app) handleCategoryByName(w http.ResponseWriter, r *http.Request) {
	name, err := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/api/categories/"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "分类名称不正确")
		return
	}
	name = strings.TrimSpace(name)
	if name == "" || name == "全部" {
		writeError(w, http.StatusBadRequest, "不能删除这个分类")
		return
	}

	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	a.deleteCategory(w, name)
}

func (a *app) deleteCategory(w http.ResponseWriter, name string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	sites, err := a.loadSitesLocked()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}

	updated := 0
	now := time.Now().Format(time.RFC3339)
	for i := range sites {
		if sites[i].Category == name {
			sites[i].Category = ""
			sites[i].UpdatedAt = now
			updated++
		}
	}

	if updated == 0 {
		writeError(w, http.StatusNotFound, "没有找到这个分类")
		return
	}

	if err := a.saveSitesLocked(sites); err != nil {
		writeError(w, http.StatusInternalServerError, "保存站点数据失败")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"uncategorizedSites": updated})
}

func (a *app) handleCategoryStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	sites, err := a.loadSites()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}

	counts := map[string]int{}
	for _, site := range sites {
		if site.Category != "" {
			counts[site.Category]++
		}
	}

	categories := make([]CategoryStat, 0, len(counts))
	for name, count := range counts {
		categories = append(categories, CategoryStat{Name: name, Count: count})
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	writeJSON(w, http.StatusOK, categories)
}

func (a *app) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
		return
	}

	sites, err := a.loadSites()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取站点数据失败")
		return
	}

	categories := map[string]bool{}
	for _, site := range sites {
		if site.Category != "" {
			categories[site.Category] = true
		}
	}

	writeJSON(w, http.StatusOK, Stats{
		SiteCount:     len(sites),
		CategoryCount: len(categories),
		Coverage:      "99%",
	})
}

func (a *app) ensureDataFile() error {
	if _, err := os.Stat(a.dataPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(a.dataPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(a.dataPath, []byte("[]\n"), 0644)
}

func (a *app) loadSites() ([]Site, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.loadSitesLocked()
}

func (a *app) loadSitesLocked() ([]Site, error) {
	data, err := os.ReadFile(a.dataPath)
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return []Site{}, nil
	}

	var sites []Site
	if err := json.Unmarshal(data, &sites); err != nil {
		return nil, err
	}
	return sites, nil
}

func (a *app) saveSitesLocked(sites []Site) error {
	if err := os.MkdirAll(filepath.Dir(a.dataPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(sites, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmpPath := a.dataPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, a.dataPath)
}

func normalizeSite(site *Site, fallbackSort int) {
	site.Name = strings.TrimSpace(site.Name)
	site.URL = strings.TrimSpace(site.URL)
	site.Category = strings.TrimSpace(site.Category)
	site.Icon = strings.TrimSpace(site.Icon)
	site.Description = strings.TrimSpace(site.Description)
	site.Glow = strings.TrimSpace(site.Glow)

	if site.Icon == "" {
		site.Icon = "🔗"
	}
	if site.Glow == "" {
		site.Glow = defaultGlow
	}
	if site.Sort <= 0 {
		site.Sort = fallbackSort
	}
}

func validateSite(site Site) error {
	if site.Name == "" {
		return errors.New("站点名称不能为空")
	}
	if site.URL == "" {
		return errors.New("站点地址不能为空")
	}
	parsed, err := url.ParseRequestURI(site.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("站点地址格式不正确")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("站点地址必须以 http:// 或 https:// 开头")
	}
	if site.Category == "" {
		return errors.New("站点分类不能为空")
	}
	return nil
}

func nextSort(sites []Site) int {
	maxSort := 0
	for _, site := range sites {
		if site.Sort > maxSort {
			maxSort = site.Sort
		}
	}
	return maxSort + 1
}

func sortSites(sites []Site) {
	sort.SliceStable(sites, func(i, j int) bool {
		if sites[i].Sort == sites[j].Sort {
			return sites[i].Name < sites[j].Name
		}
		return sites[i].Sort < sites[j].Sort
	})
}

func newID() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err == nil {
		return "site_" + hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("site_%d", time.Now().UnixNano())
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("写入响应失败: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

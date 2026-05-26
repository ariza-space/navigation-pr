package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"navigation/internal/domain"
)

const defaultGlow = "rgba(96,165,250,.45)"

var ErrNotFound = errors.New("not found")

const (
	StoreOpRead = "read"
	StoreOpSave = "save"
)

type StoreError struct {
	Op  string
	Err error
}

func (e StoreError) Error() string {
	return e.Err.Error()
}

func (e StoreError) Unwrap() error {
	return e.Err
}

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type SiteStore interface {
	ListSites() ([]domain.Site, error)
	SaveSites([]domain.Site) error
}

type SiteService struct {
	mu    sync.Mutex
	store SiteStore
}

func NewSiteService(store SiteStore) *SiteService {
	return &SiteService{store: store}
}

func (s *SiteService) ListSites(category, query string) ([]domain.Site, error) {
	category = strings.TrimSpace(category)
	query = strings.ToLower(strings.TrimSpace(query))

	sites, err := s.store.ListSites()
	if err != nil {
		return nil, err
	}

	filtered := make([]domain.Site, 0, len(sites))
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
	return filtered, nil
}

func (s *SiteService) CreateSite(input domain.Site) (domain.Site, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sites, err := s.store.ListSites()
	if err != nil {
		return domain.Site{}, StoreError{Op: StoreOpRead, Err: err}
	}

	now := time.Now().Format(time.RFC3339)
	input.ID = newID()
	input.CreatedAt = now
	input.UpdatedAt = now
	normalizeSite(&input, nextSort(sites))

	if err := validateSite(input); err != nil {
		return domain.Site{}, err
	}

	sites = append(sites, input)
	sortSites(sites)
	if err := s.store.SaveSites(sites); err != nil {
		return domain.Site{}, StoreError{Op: StoreOpSave, Err: err}
	}

	return input, nil
}

func (s *SiteService) UpdateSite(id string, input domain.Site) (domain.Site, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sites, err := s.store.ListSites()
	if err != nil {
		return domain.Site{}, StoreError{Op: StoreOpRead, Err: err}
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
			return domain.Site{}, err
		}
		sites[i] = input
		sortSites(sites)
		if err := s.store.SaveSites(sites); err != nil {
			return domain.Site{}, StoreError{Op: StoreOpSave, Err: err}
		}
		return input, nil
	}

	return domain.Site{}, ErrNotFound
}

func (s *SiteService) DeleteSite(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sites, err := s.store.ListSites()
	if err != nil {
		return StoreError{Op: StoreOpRead, Err: err}
	}

	for i, site := range sites {
		if site.ID != id {
			continue
		}
		sites = append(sites[:i], sites[i+1:]...)
		if err := s.store.SaveSites(sites); err != nil {
			return StoreError{Op: StoreOpSave, Err: err}
		}
		return nil
	}

	return ErrNotFound
}

func (s *SiteService) ListCategories() ([]string, error) {
	sites, err := s.store.ListSites()
	if err != nil {
		return nil, err
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
	return categories, nil
}

func (s *SiteService) DeleteCategory(name string) (int, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "全部" {
		return 0, ValidationError{Message: "不能删除这个分类"}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sites, err := s.store.ListSites()
	if err != nil {
		return 0, StoreError{Op: StoreOpRead, Err: err}
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
		return 0, ErrNotFound
	}
	if err := s.store.SaveSites(sites); err != nil {
		return 0, StoreError{Op: StoreOpSave, Err: err}
	}
	return updated, nil
}

func (s *SiteService) CategoryStats() ([]domain.CategoryStat, error) {
	sites, err := s.store.ListSites()
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}
	for _, site := range sites {
		if site.Category != "" {
			counts[site.Category]++
		}
	}

	categories := make([]domain.CategoryStat, 0, len(counts))
	for name, count := range counts {
		categories = append(categories, domain.CategoryStat{Name: name, Count: count})
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return categories, nil
}

func (s *SiteService) Stats() (domain.Stats, error) {
	sites, err := s.store.ListSites()
	if err != nil {
		return domain.Stats{}, err
	}

	categories := map[string]bool{}
	for _, site := range sites {
		if site.Category != "" {
			categories[site.Category] = true
		}
	}

	return domain.Stats{
		SiteCount:     len(sites),
		CategoryCount: len(categories),
		Coverage:      "99%",
	}, nil
}

func normalizeSite(site *domain.Site, fallbackSort int) {
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

func validateSite(site domain.Site) error {
	if site.Name == "" {
		return ValidationError{Message: "站点名称不能为空"}
	}
	if site.URL == "" {
		return ValidationError{Message: "站点地址不能为空"}
	}
	parsed, err := url.ParseRequestURI(site.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ValidationError{Message: "站点地址格式不正确"}
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ValidationError{Message: "站点地址必须以 http:// 或 https:// 开头"}
	}
	if site.Category == "" {
		return ValidationError{Message: "站点分类不能为空"}
	}
	return nil
}

func nextSort(sites []domain.Site) int {
	maxSort := 0
	for _, site := range sites {
		if site.Sort > maxSort {
			maxSort = site.Sort
		}
	}
	return maxSort + 1
}

func sortSites(sites []domain.Site) {
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

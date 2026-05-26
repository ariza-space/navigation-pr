package domain

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

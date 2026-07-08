package pagination

import "strconv"

const (
	// DefaultPage is used when the page parameter is absent or invalid.
	DefaultPage = 1
	// DefaultPageSize is used when the pageSize parameter is absent or invalid.
	DefaultPageSize = 20
	// MaxPageSize protects APIs from unbounded page sizes.
	MaxPageSize = 100
)

// Page describes normalized pagination input.
type Page struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

// Parse normalizes string page parameters into a bounded Page.
func Parse(pageValue string, pageSizeValue string) Page {
	page := parsePositiveInt(pageValue, DefaultPage)
	pageSize := parsePositiveInt(pageSizeValue, DefaultPageSize)
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Page{PageNo: page, PageSize: pageSize}
}

// LimitOffset returns SQL-style limit and offset values.
func (p Page) LimitOffset() (int, int) {
	return p.PageSize, (p.PageNo - 1) * p.PageSize
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

package pagination

import "strconv"

const (
	// DefaultPage 是页码缺失或无效时使用的默认值。
	DefaultPage = 1
	// DefaultPageSize 是 pageSize 缺失或无效时使用的默认值。
	DefaultPageSize = 20
	// MaxPageSize 用于限制分页大小，避免无界查询。
	MaxPageSize = 100
)

// Page 描述标准化后的分页参数。
type Page struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

// Parse 将字符串分页参数标准化为有边界的 Page。
func Parse(pageValue string, pageSizeValue string) Page {
	page := parsePositiveInt(pageValue, DefaultPage)
	pageSize := parsePositiveInt(pageSizeValue, DefaultPageSize)
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Page{PageNo: page, PageSize: pageSize}
}

// LimitOffset 返回 SQL 风格的 limit 和 offset。
func (p Page) LimitOffset() (int, int) {
	return p.PageSize, (p.PageNo - 1) * p.PageSize
}

// TotalPages 根据总记录数和页大小计算总页数。
func TotalPages(totalRecords int64, pageSize int64) int64 {
	if totalRecords <= 0 || pageSize <= 0 {
		return 0
	}
	return (totalRecords + pageSize - 1) / pageSize
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

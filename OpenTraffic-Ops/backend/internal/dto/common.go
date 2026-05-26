package dto

// PageQuery 分页查询
type PageQuery struct {
	PageNum  int    `form:"pageNum" json:"pageNum" binding:"min=1"`
	PageSize int    `form:"pageSize" json:"pageSize" binding:"min=1"`
	OrderBy  string `form:"orderByColumn" json:"orderByColumn"`
	IsAsc    string `form:"isAsc" json:"isAsc"`
}

// GetOffset 获取偏移量
func (p *PageQuery) GetOffset() int {
	if p.PageNum <= 0 {
		p.PageNum = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return (p.PageNum - 1) * p.PageSize
}

// GetLimit 获取每页数量
func (p *PageQuery) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return p.PageSize
}

// GetOrder 获取排序
func (p *PageQuery) GetOrder(defaultOrder string) string {
	if p.OrderBy == "" {
		return defaultOrder
	}
	asc := "asc"
	if p.IsAsc == "desc" {
		asc = "desc"
	}
	return p.OrderBy + " " + asc
}

package dto

// PaginationQuery represents standard list request query parameters
type PaginationQuery struct {
	Page  int `form:"page,default=1" binding:"omitempty,min=1"`
	Limit int `form:"limit,default=20" binding:"omitempty,min=1,max=100"`
}

// GetOffset calculates SQL query OFFSET
func (p *PaginationQuery) GetOffset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.GetLimit()
}

// GetLimit returns limit with default fallback
func (p *PaginationQuery) GetLimit() int {
	if p.Limit <= 0 {
		return 20
	}
	if p.Limit > 100 {
		return 100
	}
	return p.Limit
}

// PaginatedResponse wraps generic items with pagination metadata
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

package httpx

import "math"

type PaginationDTO struct {
	Total     int64 `json:"total" example:"120"`
	Current   int   `json:"current" example:"3"`
	TotalPage int   `json:"total_page" example:"6"`
	PageSize  int   `json:"page_size" example:"20"`
}

func NewPaginationDTO(pageSize, pageNum int, total int64) *PaginationDTO {
	return &PaginationDTO{
		Total:     total,
		PageSize:  pageSize,
		TotalPage: int(math.Ceil(float64(total) / float64(pageSize))),
		Current:   pageNum,
	}
}

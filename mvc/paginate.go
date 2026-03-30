package mvc

import (
	"gorm.io/gorm"
)

type Page struct {
	PageNum int `json:"pageNum"`
	Size    int `json:"size"`
	Sort    any `json:"sort"`
}

func Paginate(page *Page) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		pageNum := page.PageNum
		size := page.Size

		if pageNum == 0 {
			pageNum = 1
		}

		if size <= 0 {
			size = 10
		}

		// 如果Page结构体中包含排序字段，则设置排序条件
		if page.Sort != nil {
			db = db.Order(page.Sort)
		}

		offset := (pageNum - 1) * size
		return db.Offset(offset).Limit(size)
	}
}

func PaginateEs(page Page) (int, int) {
	pageNum := page.PageNum
	size := page.Size

	if pageNum == 0 {
		pageNum = 1
	}

	offset := (pageNum - 1) * size

	return offset, size
}

func (page *Page) Paginate() (int, int) {
	pageNum := page.PageNum
	size := page.Size

	if pageNum == 0 {
		pageNum = 1
	}

	offset := (pageNum - 1) * size

	return offset, size
}

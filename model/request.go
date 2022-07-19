package model

// PageInfo Paging common input parameter structure
type PageInfo struct {
	Page         int      `json:"page" form:"page"`           // 页码
	PageSize     int      `json:"page_size" form:"page_size"` // 每页大小
	UserId       int      `json:"user_id" form:"user_id"`
	Flag         int      `json:"flag" form:"flag"`
	CreateUserId int      `db:"create_user_id" json:"create_user_id"` // 创建人用户id（tbl_user_id）
	Factors      []string `db:"factors" json:"factors"`
}

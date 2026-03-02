// Package dto VO 定义，Service 层跨模块传递，必须脱敏（规范 5.2）
package dto

// UserVO 用户视图对象，已脱敏
type UserVO struct {
	ID        int64   `json:"id"`
	Nickname  string  `json:"nickname"`
	Avatar    string  `json:"avatar,omitempty"`
	Spark     float64 `json:"spark"`
	Status    int     `json:"status"`
	CreatedAt int64   `json:"created_at,omitempty"`
}

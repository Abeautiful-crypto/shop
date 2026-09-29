package dto

// CreateCommand 包含创建 shop 记录的入参。
type CreateCommand struct {
	Name        string
	Description string
}

// UpdateCommand 包含更新 shop 记录的入参。
type UpdateCommand struct {
	ID          string
	Name        string
	Description string
}

// ListCommand 包含查询 shop 记录列表的分页入参。
type ListCommand struct {
	Page     int32
	PageSize int32
}

package metadata_fields

type RegisterFieldsRequest struct {
	Fields []RegisterFieldItem `json:"fields" validate:"required,min=1,max=60,dive"`
}

type RegisterFieldItem struct {
	Name string `json:"name" validate:"required,min=1,max=64,notreservedmeta"`
	Type string `json:"type" validate:"required,oneof=str num bool date arr"`
}


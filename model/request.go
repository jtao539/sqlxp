package model

// PageInfo Paging common input parameter structure
type PageInfo struct {
	Page     int `json:"page" form:"page"`           // 页码
	PageSize int `json:"page_size" form:"page_size"` // 每页大小
	SelectParams
}

type SelectParams struct {
	factorsMap map[string][]interface{}
	fieldsMap  map[string]string
}

// AddFactors 增加查询条件 str 为sql语句, params 为sql语句参数
func (s *SelectParams) AddFactors(str string, params ...interface{}) {
	if s.factorsMap == nil {
		s.factorsMap = make(map[string][]interface{})
	}
	s.factorsMap[str] = params
}

// AddFields 增加查询字段 alias 字段别名, str 为该字段的查询sql语句
func (s *SelectParams) AddFields(alias string, str string) {
	if s.fieldsMap == nil {
		s.fieldsMap = make(map[string]string)
	}
	s.fieldsMap[alias] = str
}

func (s *SelectParams) GetFieldsMap() map[string]string {
	return s.fieldsMap
}

func (s *SelectParams) GetFactorsMap() map[string][]interface{} {
	return s.factorsMap
}

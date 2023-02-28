package model

type SelectParams struct {
	factorsMap map[string][]interface{}
	fieldsMap  map[string]string
	ordersMap  map[string]bool
	groupsMap  map[string]string
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

func (s *SelectParams) OrderBy(field string, desc bool) {
	if s.ordersMap == nil {
		s.ordersMap = make(map[string]bool)
	}
	s.ordersMap[field] = desc
}

func (s *SelectParams) GroupBy(field string) {
	if s.groupsMap == nil {
		s.groupsMap = make(map[string]string)
	}
	s.groupsMap[field] = ""
}

func (s *SelectParams) GetFieldsMap() map[string]string {
	return s.fieldsMap
}

func (s *SelectParams) GetFactorsMap() map[string][]interface{} {
	return s.factorsMap
}

package model

type SelectParams struct {
	factorsMap  map[string][]interface{}
	fieldsMap   map[string]string
	ordersMap   map[string]bool
	ordersArray []string // 保证顺序
	groupsArray []string // 保证顺序
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
	if s.ordersArray == nil {
		s.ordersArray = make([]string, 0)
	}
	s.ordersMap[field] = desc
	for i := 0; i < len(s.ordersArray); i++ {
		if s.ordersArray[i] == field {
			return
		}
	}
	s.ordersArray = append(s.ordersArray, field)
}

func (s *SelectParams) GroupBy(field string) {
	if s.groupsArray == nil {
		s.groupsArray = make([]string, 0)
	}
	for i := 0; i < len(s.groupsArray); i++ {
		if s.groupsArray[i] == field {
			return
		}
	}
	s.groupsArray = append(s.groupsArray, field)
}

func (s *SelectParams) GetFieldsMap() map[string]string {
	return s.fieldsMap
}

func (s *SelectParams) GetFactorsMap() map[string][]interface{} {
	return s.factorsMap
}

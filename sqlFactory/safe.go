package sqlFactory

import (
	"fmt"
	"github.com/jtao539/sqlxp/util"
	"reflect"
	"strconv"
	"strings"
)

func SafeInsert(o interface{}, tbl string) string {
	sql := "INSERT INTO "
	sql += tbl + "("
	t := reflect.TypeOf(o)
	for i := 0; i < t.NumField(); i++ {
		tagName := t.Field(i).Tag.Get(Tag)
		sql += tagName + ","
	}
	if strings.Contains(sql, ",") {
		sql = sql[:strings.LastIndex(sql, ",")]
	}
	sql += ") VALUES ("
	for i := 0; i < t.NumField(); i++ {
		tagName := t.Field(i).Tag.Get(Tag)
		sql += ":" + tagName + ","
	}
	if strings.Contains(sql, ",") {
		sql = sql[:strings.LastIndex(sql, ",")]
	}
	sql += ")"
	return sql
}

// SafeUpdate 安全的更新语句生成，o 为DTO, a 为entity tbl 为表名称， 通过对比o和a获取跳过的字段
// 返回值包含带占位符的sql和参数数组
func SafeUpdate(o interface{}, a interface{}, tbl string) (string, []interface{}) {
	return safeLocalUpdate(o, a, tbl, func(tagName string) bool {
		return contain(tagName, "id")
	})
}

// SafeUpdateP 安全的更新语句生成，o 为DTO, a 为entity tbl 为表名称， 通过对比o和a获取跳过的字段，args 为需要手动跳过的字段
// 返回值包含带占位符的sql和参数数组
func SafeUpdateP(o interface{}, a interface{}, tbl string, args ...string) (string, []interface{}) {
	return safeLocalUpdate(o, a, tbl, func(tagName string) bool {
		return containArray(tagName, args)
	}, func(tagName string) bool {
		return contain(tagName, "id")
	})
}

// SafeLocalUpdate 安全的更新语句生成，o 为DTO, a 为entity tbl 为表名称， 通过对比o和a获取跳过的字段， fs为所有满足条件需要跳过的函数
// 返回值为带占位符的SQL以及对应的参数数组
func safeLocalUpdate(o interface{}, a interface{}, tbl string, fs ...func(tagName string) bool) (sqlStr string, params []interface{}) {
	var paramsResult []interface{}
	sql := "UPDATE "
	sql += tbl + " SET "
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	for i := 0; i < t.NumField(); i++ {
		tagName := t.Field(i).Tag.Get(Tag)
		if noContain(tagName, a) {
			continue
		}
		flag := false
		for i := 0; i < len(fs); i++ {
			if fs[i](tagName) {
				flag = true
			}
		}
		if flag {
			continue
		}
		sql += tagName + "="
		switch v.Field(i).Kind() {
		case reflect.Int:
			if v.Field(i).Int() == 0 {
				sql += "NULL" + ", "
			} else {
				sql += "?, "
				paramsResult = append(paramsResult, strconv.FormatInt(v.Field(i).Int(), 10))
			}
		case reflect.String:
			if v.Field(i).String() == "" {
				sql += "NULL" + ", "
			} else {
				sql += "?, "
				paramsResult = append(paramsResult, v.Field(i).String())
			}
		case reflect.Float64:
			if v.Field(i).Float() == 0 {
				sql += "NULL" + ", "
			} else {
				sql += "?, "
				value := strconv.FormatFloat(v.Field(i).Float(), 'g', 15, 64)
				paramsResult = append(paramsResult, value)
			}
		case reflect.Array, reflect.Slice:
			v := v.Field(i)
			sql += "?, "
			var result string
			for i := 0; i < v.Len(); i++ {
				result += v.Index(i).String() + ","
			}
			if strings.Contains(result, ",") {
				result = result[:strings.LastIndex(result, ",")]
			}
			paramsResult = append(paramsResult, result)
		}
	}
	if strings.Contains(sql, ",") {
		sql = sql[:strings.LastIndex(sql, ",")]
	}
	sql += " where id=?"
	paramsResult = append(paramsResult, strconv.FormatInt(v.FieldByName("Id").Int(), 10))
	return sql, paramsResult
}

// SafeSelect 安全条件查询语句生成(采用参数化查询，未直接拼接SQL语句), o 为DTO, a 为entity tbl 为表名称, tags为手动跳过的查找字段
// 返回值为带占位符的SQL以及对应的参数数组
func SafeSelect(o interface{}, tbl string, tags ...string) (sqlStr string, params []interface{}, countSql string) {
	var paramsResult []interface{}
	ov := reflect.ValueOf(o)
	ot := reflect.TypeOf(o)
	var DTO reflect.Value
	for i := 0; i < ot.NumField(); i++ {
		if ot.Field(i).Name != "PageInfo" {
			DTO = ov.Field(i)
		}
	}
	PageInfo := ov.FieldByName("PageInfo")
	dt := DTO.Type()
	sql := "SELECT *"
	sql += " FROM " + tbl + " WHERE "
	for i := 0; i < dt.NumField(); i++ {
		tagName := dt.Field(i).Tag.Get(Tag)
		if containArray(tagName, tags) {
			continue
		}
		switch DTO.Field(i).Kind() {
		case reflect.Int:
			if DTO.Field(i).Int() != 0 {
				sql += tagName + "=? " + " AND "
				paramsResult = append(paramsResult, strconv.FormatInt(DTO.Field(i).Int(), 10))
			}
		case reflect.String:
			if DTO.Field(i).String() != "" {
				sql += tagName + " like ?" + " AND "
				paramsResult = append(paramsResult, "%"+DTO.Field(i).String()+"%")
			}
		case reflect.Float64:
			if DTO.Field(i).Float() != 0 {
				value := strconv.FormatFloat(DTO.Field(i).Float(), 'g', 15, 64)
				sql += tagName + "=? AND "
				paramsResult = append(paramsResult, value)
			}
		}
	}
	if strings.Contains(sql, "AND") {
		sql = sql[:strings.LastIndex(sql, "AND")]
	} else {
		sql += "1=1"
	}
	countStr := strings.ReplaceAll(sql, "*", " COUNT(1) as total ")
	page := PageInfo.FieldByName("Page").Int()
	pageSize := PageInfo.FieldByName("PageSize").Int()
	if page > 0 && pageSize > 0 {
		sql += " limit " + strconv.FormatInt((page-1)*pageSize, 10) + " , " + strconv.FormatInt(pageSize, 10)
	}
	return sql, paramsResult, countStr
}

func SafeSelectOrder(o interface{}, tbl string, desc bool, orderField string, tags ...string) (sqlStr string, params []interface{}, countSql string) {
	var paramsResult []interface{}
	ov := reflect.ValueOf(o)
	ot := reflect.TypeOf(o)
	var DTO reflect.Value
	for i := 0; i < ot.NumField(); i++ {
		if ot.Field(i).Name != "PageInfo" {
			DTO = ov.Field(i)
		}
	}
	PageInfo := ov.FieldByName("PageInfo")
	dt := DTO.Type()
	sql := "SELECT *"
	sql += " FROM " + tbl + " WHERE "
	for i := 0; i < dt.NumField(); i++ {
		tagName := dt.Field(i).Tag.Get(Tag)
		if containArray(tagName, tags) {
			continue
		}
		switch DTO.Field(i).Kind() {
		case reflect.Int:
			if DTO.Field(i).Int() != 0 {
				sql += tagName + "=? " + " AND "
				paramsResult = append(paramsResult, strconv.FormatInt(DTO.Field(i).Int(), 10))
			}
		case reflect.String:
			if DTO.Field(i).String() != "" {
				sql += tagName + " like ?" + " AND "
				paramsResult = append(paramsResult, "%"+DTO.Field(i).String()+"%")
			}
		case reflect.Float64:
			if DTO.Field(i).Float() != 0 {
				value := strconv.FormatFloat(DTO.Field(i).Float(), 'g', 15, 64)
				sql += tagName + "=? AND "
				paramsResult = append(paramsResult, value)
			}
		}
	}
	if strings.Contains(sql, "AND") {
		sql = sql[:strings.LastIndex(sql, "AND")]
	} else {
		sql += "1=1"
	}
	field := "id"
	if strings.TrimSpace(orderField) != "" {
		field = orderField
	}
	if desc {
		sql += " order by " + field + " desc "
	} else {
		sql += " order by " + field
	}
	countStr := strings.ReplaceAll(sql, "*", " COUNT(1) as total ")
	page := PageInfo.FieldByName("Page").Int()
	pageSize := PageInfo.FieldByName("PageSize").Int()
	if page > 0 && pageSize > 0 {
		sql += " limit " + strconv.FormatInt((page-1)*pageSize, 10) + " , " + strconv.FormatInt(pageSize, 10)
	}
	return sql, paramsResult, countStr
}

// SafeSelectMT 多表查询语句生成(采用参数化查询，未直接拼接SQL语句), o 为DTO, a 为entity tbl 为表名称, otherFiledMap为表字段与sql映射, factors 为条件 tags为手动跳过的查找字段
// 返回值为带占位符的SQL以及对应的参数数组
func SafeSelectMT(o interface{}, tbl string, otherFiledMap map[string]string, factors []string, desc bool, tags ...string) (sqlStr string, params []interface{}, countSql string) {
	var paramsResult []interface{}
	selectSql := "SELECT * "
	if len(otherFiledMap) > 0 {
		selectSql += ","
		for filed, v := range otherFiledMap {
			selectSql += fmt.Sprintf(" (%s) as %s ,", v, filed)
		}
		if strings.Contains(selectSql, ",") {
			selectSql = selectSql[:strings.LastIndex(selectSql, ",")]
		}
	}
	sql := "FROM " + tbl + " WHERE "
	ov := reflect.ValueOf(o)
	ot := reflect.TypeOf(o)
	var DTO reflect.Value
	for i := 0; i < ot.NumField(); i++ {
		if ot.Field(i).Name != "PageInfo" {
			DTO = ov.Field(i)
		}
	}
	PageInfo := ov.FieldByName("PageInfo")
	dt := DTO.Type()
	for i := 0; i < dt.NumField(); i++ {
		tagName := dt.Field(i).Tag.Get(Tag)
		if containArray(tagName, tags) {
			continue
		}
		switch DTO.Field(i).Kind() {
		case reflect.Int:
			if DTO.Field(i).Int() != 0 {
				sql += tagName + "=? " + " AND "
				paramsResult = append(paramsResult, strconv.FormatInt(DTO.Field(i).Int(), 10))
			}
		case reflect.String:
			if DTO.Field(i).String() != "" {
				sql += tagName + " like ?" + " AND "
				paramsResult = append(paramsResult, "%"+DTO.Field(i).String()+"%")
			}
		case reflect.Float64:
			if DTO.Field(i).Float() != 0 {
				value := strconv.FormatFloat(DTO.Field(i).Float(), 'g', 15, 64)
				sql += tagName + "=? AND "
				paramsResult = append(paramsResult, value)
			}
		}
	}
	for i := 0; i < len(factors) && strings.TrimSpace(factors[i]) != ""; i++ {
		sql += fmt.Sprintf(" %s AND ", factors[i])
	}
	if strings.Contains(sql, "AND") {
		sql = sql[:strings.LastIndex(sql, "AND")]
	} else {
		sql += "1=1"
	}
	if desc {
		sql += " order by " + "id" + " desc "
	} else {
		sql += " order by id  "
	}
	countStr := "SELECT COUNT(1) as total " + sql
	page := PageInfo.FieldByName("Page").Int()
	pageSize := PageInfo.FieldByName("PageSize").Int()
	if page > 0 && pageSize > 0 {
		sql += " limit " + strconv.FormatInt((page-1)*pageSize, 10) + " , " + strconv.FormatInt(pageSize, 10)
	}
	return selectSql + sql, paramsResult, countStr
}

// SafeSelectMTP [更加安全-防止sql注入] 多表查询语句生成(采用参数化查询，未直接拼接SQL语句), o 为DTO, a 为entity tbl 为表名称, otherFiledMap为表字段与sql映射, factorsMap 为条件和参数的map tags为手动跳过的查找字段
// 返回值为带占位符的SQL以及对应的参数数组
func SafeSelectMTP(o interface{}, tbl string, otherFiledMap map[string]string, factorsMap map[string][]interface{}, desc bool, tags ...string) (sqlStr string, params []interface{}, countSql string) {
	var paramsResult []interface{}
	selectSql := "SELECT * "
	if len(otherFiledMap) > 0 {
		selectSql += ","
		for filed, v := range otherFiledMap {
			selectSql += fmt.Sprintf(" (%s) as %s ,", v, filed)
		}
		if strings.Contains(selectSql, ",") {
			selectSql = selectSql[:strings.LastIndex(selectSql, ",")]
		}
	}
	sql := "FROM " + tbl + " WHERE "
	ov := reflect.ValueOf(o)
	ot := reflect.TypeOf(o)
	var DTO reflect.Value
	for i := 0; i < ot.NumField(); i++ {
		if ot.Field(i).Name != "PageInfo" {
			DTO = ov.Field(i)
		}
	}
	PageInfo := ov.FieldByName("PageInfo")
	dt := DTO.Type()
	for i := 0; i < dt.NumField(); i++ {
		tagName := dt.Field(i).Tag.Get(Tag)
		if containArray(tagName, tags) {
			continue
		}
		switch DTO.Field(i).Kind() {
		case reflect.Int:
			if DTO.Field(i).Int() != 0 {
				sql += tagName + "=? " + " AND "
				paramsResult = append(paramsResult, strconv.FormatInt(DTO.Field(i).Int(), 10))
			}
		case reflect.String:
			if DTO.Field(i).String() != "" {
				sql += tagName + " like ?" + " AND "
				paramsResult = append(paramsResult, "%"+DTO.Field(i).String()+"%")
			}
		case reflect.Float64:
			if DTO.Field(i).Float() != 0 {
				value := strconv.FormatFloat(DTO.Field(i).Float(), 'g', 15, 64)
				sql += tagName + "=? AND "
				paramsResult = append(paramsResult, value)
			}
		}
	}
	for k, v := range factorsMap {
		if strings.TrimSpace(k) != "" {
			sql += fmt.Sprintf(" %s AND ", k)
			params := util.AnythingToSlice(v)
			for i := 0; i < len(params); i++ {
				paramsResult = append(paramsResult, params[i])
			}
		}
	}

	if strings.Contains(sql, "AND") {
		sql = sql[:strings.LastIndex(sql, "AND")]
	} else {
		sql += "1=1"
	}
	if desc {
		sql += " order by " + "id" + " desc "
	} else {
		sql += " order by id  "
	}
	countStr := "SELECT COUNT(1) as total " + sql
	page := PageInfo.FieldByName("Page").Int()
	pageSize := PageInfo.FieldByName("PageSize").Int()
	if page > 0 && pageSize > 0 {
		sql += " limit " + strconv.FormatInt((page-1)*pageSize, 10) + " , " + strconv.FormatInt(pageSize, 10)
	}
	return selectSql + sql, paramsResult, countStr
}

// SafeSelectP [更加安全-防止sql注入] 多表查询语句生成(采用参数化查询，未直接拼接SQL语句), o 为DTO, a 为entity tbl 为表名称, otherFiledMap为表字段与sql映射, factorsMap 为条件和参数的map tags为手动跳过的查找字段
// 返回值为带占位符的SQL以及对应的参数数组
func SafeSelectP(o interface{}, tbl string, tags ...string) (sqlStr string, params []interface{}, countSql string) {
	var paramsResult []interface{}
	// 查询表字段
	selectSql := "SELECT * "
	ov := reflect.ValueOf(o)
	PageInfo := ov.FieldByName("PageInfo")
	// 查询外表字段
	fieldsMap := PageInfo.FieldByName("fieldsMap")
	fieldsR := fieldsMap.MapRange()
	selectSql += ","
	for fieldsR.Next() {
		k := fieldsR.Key().String()
		v := fieldsR.Value().String()
		selectSql += fmt.Sprintf(" (%s) as %s ,", v, k)
	}
	if strings.Contains(selectSql, ",") {
		selectSql = selectSql[:strings.LastIndex(selectSql, ",")]
	}
	sql := "FROM " + tbl + " WHERE "
	// 查询条件-表字段
	ot := reflect.TypeOf(o)
	var DTO reflect.Value
	for i := 0; i < ot.NumField(); i++ {
		if ot.Field(i).Name != "PageInfo" {
			DTO = ov.Field(i)
		}
	}
	dt := DTO.Type()
	for i := 0; i < dt.NumField(); i++ {
		tagName := dt.Field(i).Tag.Get(Tag)
		if containArray(tagName, tags) {
			continue
		}
		switch DTO.Field(i).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if DTO.Field(i).Int() != 0 {
				sql += tagName + "=? " + " AND "
				paramsResult = append(paramsResult, strconv.FormatInt(DTO.Field(i).Int(), 10))
			}
		case reflect.String:
			if DTO.Field(i).String() != "" {
				sql += tagName + " like ?" + " AND "
				paramsResult = append(paramsResult, "%"+DTO.Field(i).String()+"%")
			}
		case reflect.Float64, reflect.Float32:
			if DTO.Field(i).Float() != 0 {
				value := strconv.FormatFloat(DTO.Field(i).Float(), 'g', 15, 64)
				sql += tagName + "=? AND "
				paramsResult = append(paramsResult, value)
			}
		case reflect.Bool:
			sql += tagName + "=? AND "
			paramsResult = append(paramsResult, DTO.Field(i).Bool())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if DTO.Field(i).Uint() != 0 {
				sql += tagName + " =? AND "
				paramsResult = append(paramsResult, DTO.Field(i).Uint())
			}
		}
	}
	// 查询条件-复杂条件
	factorsMap := PageInfo.FieldByName("factorsMap")
	factorsR := factorsMap.MapRange()
	for factorsR.Next() {
		k := factorsR.Key().String()
		v := factorsR.Value()
		if strings.TrimSpace(k) != "" {
			sql += fmt.Sprintf(" %s AND ", k)
			for i := 0; i < v.Len(); i++ {
				e := v.Index(i).Elem()
				switch e.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					paramsResult = append(paramsResult, e.Int())
				case reflect.String:
					paramsResult = append(paramsResult, e.String())
				case reflect.Float64, reflect.Float32:
					paramsResult = append(paramsResult, e.Float())
				case reflect.Bool:
					paramsResult = append(paramsResult, e.Bool())
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					paramsResult = append(paramsResult, e.Uint())
				}
			}
		}
	}
	if strings.Contains(sql, "AND") {
		sql = sql[:strings.LastIndex(sql, "AND")]
	} else {
		sql += "1=1"
	}
	// 分组处理
	groupsMap := PageInfo.FieldByName("groupsMap")
	groupsR := groupsMap.MapRange()
	if groupsMap.Len() > 0 {
		sql += " group by "
	}
	for groupsR.Next() {
		k := groupsR.Key().String()
		sql += k + " , "
	}
	if groupsMap.Len() > 0 {
		sql = sql[:strings.LastIndex(sql, ",")]
	}
	// 排序处理
	ordersMap := PageInfo.FieldByName("ordersMap")
	ordersR := ordersMap.MapRange()
	if ordersMap.Len() > 0 {
		sql += " order by "
	}
	for ordersR.Next() {
		k := factorsR.Key().String()
		v := factorsR.Value().Bool()
		if v {
			sql += k + " desc , "
		} else {
			sql += k + " , "
		}
	}
	if ordersMap.Len() > 0 {
		sql = sql[:strings.LastIndex(sql, ",")]
	}
	countStr := "SELECT COUNT(1) as total " + sql
	page := PageInfo.FieldByName("Page").Int()
	pageSize := PageInfo.FieldByName("PageSize").Int()
	if page > 0 && pageSize > 0 {
		sql += " limit " + strconv.FormatInt((page-1)*pageSize, 10) + " , " + strconv.FormatInt(pageSize, 10)
	}
	return selectSql + sql, paramsResult, countStr
}

// SafeSelectWithFactor 安全的可手动介入查询条件的查询语句生成
// 返回值为带占位符的SQL以及对应的参数数组
func SafeSelectWithFactor(o interface{}, tbl string, factors []string, tags ...string) (sqlStr string, params []interface{}, countSql string) {
	var paramsResult []interface{}
	ov := reflect.ValueOf(o)
	ot := reflect.TypeOf(o)
	var DTO reflect.Value
	for i := 0; i < ot.NumField(); i++ {
		if ot.Field(i).Name != "PageInfo" {
			DTO = ov.Field(i)
		}
	}
	PageInfo := ov.FieldByName("PageInfo")
	dt := DTO.Type()
	sql := "SELECT "
	for i := 0; i < dt.NumField(); i++ {
		f := dt.Field(i).Tag.Get(Tag)
		if DTO.Field(i).Kind() == reflect.Struct {
			continue
		}
		sql += f + ","
	}
	if strings.Contains(sql, ",") {
		sql = sql[:strings.LastIndex(sql, ",")]
	}
	sql += " FROM " + tbl + " WHERE "
	for i := 0; i < dt.NumField(); i++ {
		tagName := dt.Field(i).Tag.Get(Tag)
		if containArray(tagName, tags) {
			continue
		}
		switch DTO.Field(i).Kind() {
		case reflect.Int:
			if DTO.Field(i).Int() != 0 {
				sql += tagName + "=? " + " AND "
				paramsResult = append(paramsResult, strconv.FormatInt(DTO.Field(i).Int(), 10))
			}
		case reflect.String:
			if DTO.Field(i).String() != "" {
				sql += tagName + " like ?" + " AND "
				paramsResult = append(paramsResult, "%"+DTO.Field(i).String()+"%")
			}
		case reflect.Float64:
			if DTO.Field(i).Float() != 0 {
				value := strconv.FormatFloat(DTO.Field(i).Float(), 'g', 15, 64)
				sql += tagName + "=? AND "
				paramsResult = append(paramsResult, value)
			}
		}
	}
	for i := 0; i < len(factors) && strings.TrimSpace(factors[i]) != ""; i++ {
		sql += fmt.Sprintf(" %s AND ", factors[i])
	}
	if strings.Contains(sql, "AND") {
		sql = sql[:strings.LastIndex(sql, "AND")]
	} else {
		sql += "1=1"
	}
	countStr := strings.ReplaceAll(sql, "*", " COUNT(1) as total ")
	page := PageInfo.FieldByName("Page").Int()
	pageSize := PageInfo.FieldByName("PageSize").Int()
	if page > 0 && pageSize > 0 {
		sql += " limit " + strconv.FormatInt((page-1)*pageSize, 10) + " , " + strconv.FormatInt(pageSize, 10)
	}
	return sql, paramsResult, countStr
}

func SafeSelectWithFactorOrder(o interface{}, tbl string, desc bool, orderField string, factors []string, tags ...string) (sqlStr string, params []interface{}, countSql string) {
	var paramsResult []interface{}
	ov := reflect.ValueOf(o)
	ot := reflect.TypeOf(o)
	var DTO reflect.Value
	for i := 0; i < ot.NumField(); i++ {
		if ot.Field(i).Name != "PageInfo" {
			DTO = ov.Field(i)
		}
	}
	PageInfo := ov.FieldByName("PageInfo")
	dt := DTO.Type()
	sql := "SELECT "
	for i := 0; i < dt.NumField(); i++ {
		f := dt.Field(i).Tag.Get(Tag)
		if DTO.Field(i).Kind() == reflect.Struct {
			continue
		}
		sql += f + ","
	}
	if strings.Contains(sql, ",") {
		sql = sql[:strings.LastIndex(sql, ",")]
	}
	sql += " FROM " + tbl + " WHERE "
	for i := 0; i < dt.NumField(); i++ {
		tagName := dt.Field(i).Tag.Get(Tag)
		if containArray(tagName, tags) {
			continue
		}
		switch DTO.Field(i).Kind() {
		case reflect.Int:
			if DTO.Field(i).Int() != 0 {
				sql += tagName + "=? " + " AND "
				paramsResult = append(paramsResult, strconv.FormatInt(DTO.Field(i).Int(), 10))
			}
		case reflect.String:
			if DTO.Field(i).String() != "" {
				sql += tagName + " like ?" + " AND "
				paramsResult = append(paramsResult, "%"+DTO.Field(i).String()+"%")
			}
		case reflect.Float64:
			if DTO.Field(i).Float() != 0 {
				value := strconv.FormatFloat(DTO.Field(i).Float(), 'g', 15, 64)
				sql += tagName + "=? AND "
				paramsResult = append(paramsResult, value)
			}
		}
	}
	for i := 0; i < len(factors) && strings.TrimSpace(factors[i]) != ""; i++ {
		sql += fmt.Sprintf(" %s AND ", factors[i])
	}
	if strings.Contains(sql, "AND") {
		sql = sql[:strings.LastIndex(sql, "AND")]
	} else {
		sql += "1=1"
	}
	field := "id"
	if strings.TrimSpace(orderField) != "" {
		field = orderField
	}
	if desc {
		sql += " order by " + field + " desc "
	} else {
		sql += " order by " + field
	}
	countStr := strings.ReplaceAll(sql, "*", " COUNT(1) as total ")
	page := PageInfo.FieldByName("Page").Int()
	pageSize := PageInfo.FieldByName("PageSize").Int()
	if page > 0 && pageSize > 0 {
		sql += " limit " + strconv.FormatInt((page-1)*pageSize, 10) + " , " + strconv.FormatInt(pageSize, 10)
	}
	return sql, paramsResult, countStr
}

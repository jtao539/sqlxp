package sqlFactory

import (
	"fmt"
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
		return contain(tagName, "id", "code", "status")
	})
}

// SafeUpdateP 安全的更新语句生成，o 为DTO, a 为entity tbl 为表名称， 通过对比o和a获取跳过的字段，args 为需要手动跳过的字段
// 返回值包含带占位符的sql和参数数组
func SafeUpdateP(o interface{}, a interface{}, tbl string, args ...string) (string, []interface{}) {
	return safeLocalUpdate(o, a, tbl, func(tagName string) bool {
		return containArray(tagName, args)
	}, func(tagName string) bool {
		return contain(tagName, "id", "code", "status")
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
func SafeSelect(o interface{}, tbl string, tags ...string) (sqlStr string, params []interface{}) {
	var paramsResult []interface{}
	sql := "SELECT * FROM " + tbl + " WHERE "
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
		if tagName == "id" || containArray(tagName, tags) {
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
	if PageInfo.FieldByName("CreateUserId").Int() != 0 {
		sql += "create_user_id" + "=?" + " AND "
		paramsResult = append(paramsResult, strconv.FormatInt(PageInfo.FieldByName("CreateUserId").Int(), 10))
	}
	if strings.Contains(sql, "AND") {
		sql = sql[:strings.LastIndex(sql, "AND")]
	} else {
		sql += "1=1"
	}
	page := PageInfo.FieldByName("Page").Int()
	pageSize := PageInfo.FieldByName("PageSize").Int()
	if page > 0 && pageSize > 0 {
		sql += " limit " + strconv.FormatInt((page-1)*pageSize, 10) + " , " + strconv.FormatInt(pageSize, 10)
	}
	return sql, paramsResult
}

// SafeSelectWithFactor 安全的可手动介入查询条件的查询语句生成
// 返回值为带占位符的SQL以及对应的参数数组
func SafeSelectWithFactor(o interface{}, tbl string, factors []string, tags ...string) (sqlStr string, params []interface{}) {
	var paramsResult []interface{}
	sql := "SELECT * FROM " + tbl + " WHERE "
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
		if tagName == "id" || containArray(tagName, tags) {
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
	if PageInfo.FieldByName("CreateUserId").Int() != 0 {
		sql += "create_user_id" + "=?" + " AND "
		paramsResult = append(paramsResult, strconv.FormatInt(PageInfo.FieldByName("CreateUserId").Int(), 10))
	}
	for i := 0; i < len(factors) && strings.TrimSpace(factors[i]) != ""; i++ {
		sql += fmt.Sprintf(" %s AND ", factors[i])
	}
	if strings.Contains(sql, "AND") {
		sql = sql[:strings.LastIndex(sql, "AND")]
	} else {
		sql += "1=1"
	}
	page := PageInfo.FieldByName("Page").Int()
	pageSize := PageInfo.FieldByName("PageSize").Int()
	if page > 0 && pageSize > 0 {
		sql += " limit " + strconv.FormatInt((page-1)*pageSize, 10) + " , " + strconv.FormatInt(pageSize, 10)
	}
	return sql, paramsResult
}

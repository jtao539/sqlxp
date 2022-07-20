package util

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// 用于sqlNull类型与基本数据类型转换的工具
// finish by JiangTao in 2022.07

const TAG = "db"

// N2Basic 将字段类型为sql.NullXXX的实体 转换为字段类型为普通类型，转换过程以普通类型的结构体为主。 tags 为需要跳过转换的字段, 常用于某个字段类型不一致
func N2Basic(a interface{}, b interface{}, tags ...string) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	tb := reflect.TypeOf(b)
	ovb := vb.Elem()
	otb := tb.Elem()
	for i := 0; i < ovb.NumField(); i++ {
		field := otb.Field(i)
		if tag := field.Tag.Get(TAG); len(tags) > 0 && containArray(tag, tags) {
			continue
		}
		switch ovb.Field(i).Kind() {
		case reflect.String:
			sValue := va.FieldByName(field.Name)
			if sValue.Kind() == reflect.Struct && sValue.Type().Name() == "NullString" {
				for j := 0; j < sValue.NumField(); j++ {
					if sValue.Field(j).Kind() == reflect.Bool && !sValue.Field(j).Bool() {
						ovb.FieldByName(field.Name).SetString("")
						break
					}
					if sValue.Field(j).Kind() == reflect.String {
						ovb.FieldByName(field.Name).SetString(sValue.Field(j).String())
					}
				}
			} else if sValue.Kind() == reflect.String {
				ovb.FieldByName(field.Name).SetString(sValue.String())
			}
		case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Int16:
			sValue := va.FieldByName(field.Name)
			if sValue.Kind() == reflect.Struct && (sValue.Type().Name() == "NullInt32" || sValue.Type().Name() == "NullInt64") {
				for j := 0; j < sValue.NumField(); j++ {
					if sValue.Field(j).Kind() == reflect.Bool && !sValue.Field(j).Bool() {
						ovb.FieldByName(field.Name).SetInt(0)
						break
					}
					if sValue.Field(j).Kind() == reflect.Int64 || sValue.Field(j).Kind() == reflect.Int32 || sValue.Field(j).Kind() == reflect.Int {
						ovb.FieldByName(field.Name).SetInt(sValue.Field(j).Int())
					}
				}
			} else if sValue.Kind() == reflect.Int32 || sValue.Kind() == reflect.Int64 || sValue.Kind() == reflect.Int8 || sValue.Kind() == reflect.Int16 || sValue.Kind() == reflect.Int {
				ovb.FieldByName(field.Name).SetInt(sValue.Int())
			}
		case reflect.Float64, reflect.Float32:
			sValue := va.FieldByName(field.Name)
			if sValue.Kind() == reflect.Struct && (sValue.Type().Name() == "NullFloat64" || sValue.Type().Name() == "NullFloat32") {
				for j := 0; j < sValue.NumField(); j++ {
					if sValue.Field(j).Kind() == reflect.Float64 || sValue.Field(j).Kind() == reflect.Float32 {
						ovb.FieldByName(field.Name).SetFloat(sValue.Field(j).Float())
					}
				}
			} else if sValue.Kind() == reflect.Float32 || sValue.Kind() == reflect.Float64 {
				ovb.FieldByName(field.Name).SetFloat(sValue.Float())
			}
		}
	}
}

// Basic2N 将基本数据类型的实体转换为sql.NullXXX的实体，转换过程以BasicEntity为主
func Basic2N(a interface{}, b interface{}, tags ...string) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	tb := reflect.TypeOf(b)
	ovb := vb.Elem()
	otb := tb.Elem()
	for i := 0; i < ovb.NumField(); i++ {
		field := otb.Field(i)
		if tag := field.Tag.Get(TAG); len(tags) > 0 && containArray(tag, tags) {
			continue
		}
		/*if ovb.Field(i).Kind() == reflect.Struct {*/
		fieldA := va.FieldByName(field.Name)
		switch fieldA.Kind() {
		case reflect.String:
			if ovb.FieldByName(field.Name).Kind() == reflect.Struct && ovb.FieldByName(field.Name).Type().Name() == "NullString" {
				if fieldA.String() == "" {
					ovb.FieldByName(field.Name).FieldByName("Valid").SetBool(false)
				} else {
					ovb.FieldByName(field.Name).FieldByName("Valid").SetBool(true)
					ovb.FieldByName(field.Name).FieldByName("String").SetString(fieldA.String())
				}
			} else if ovb.FieldByName(field.Name).Kind() == reflect.String {
				ovb.FieldByName(field.Name).SetString(fieldA.String())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if ovb.FieldByName(field.Name).Kind() == reflect.Struct && (ovb.FieldByName(field.Name).Type().Name() == "NullInt32" || ovb.FieldByName(field.Name).Type().Name() == "NullInt64") {
				if fieldA.Int() == 0 {
					ovb.FieldByName(field.Name).FieldByName("Valid").SetBool(false)
				} else {
					for j := 0; j < ovb.FieldByName(field.Name).NumField(); j++ {
						innerField := ovb.FieldByName(field.Name).Field(j)
						if innerField.Kind() == reflect.Bool {
							innerField.SetBool(true)
						} else if innerField.Kind() == reflect.Int32 || innerField.Kind() == reflect.Int64 {
							innerField.SetInt(fieldA.Int())
						}
					}
				}
			} else if ovb.FieldByName(field.Name).Kind() == reflect.Int || ovb.FieldByName(field.Name).Kind() == reflect.Int32 || ovb.FieldByName(field.Name).Kind() == reflect.Int64 {
				ovb.FieldByName(field.Name).SetInt(fieldA.Int())
			}
		case reflect.Float32, reflect.Float64:
			if ovb.FieldByName(field.Name).Kind() == reflect.Struct && (ovb.FieldByName(field.Name).Type().Name() == "NullFloat32" || ovb.FieldByName(field.Name).Type().Name() == "NullFloat64") {
				if fieldA.Float() == 0 {
					ovb.FieldByName(field.Name).FieldByName("Valid").SetBool(false)
				} else {
					for j := 0; j < ovb.FieldByName(field.Name).NumField(); j++ {
						innerField := ovb.FieldByName(field.Name).Field(j)
						if innerField.Kind() == reflect.Bool {
							innerField.SetBool(true)
						} else if innerField.Kind() == reflect.Float32 || innerField.Kind() == reflect.Float64 {
							innerField.SetFloat(fieldA.Float())
						}
					}
				}
			} else if ovb.FieldByName(field.Name).Kind() == reflect.Float32 || ovb.FieldByName(field.Name).Kind() == reflect.Float64 {
				ovb.FieldByName(field.Name).SetFloat(fieldA.Float())
			}
		}
	}
}

func IntToNullInt32(a int) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(a), Valid: true}
}

func StringToNullString(a string) sql.NullString {
	return sql.NullString{String: a, Valid: true}
}

func anythingToSlice(a interface{}) []interface{} {
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
			t := reflect.TypeOf(result[i])
			fmt.Println("t = ", t)
		}
		return result
	default:
		panic("not supported")
	}
}

func GetLength(a interface{}) int {
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return v.Len()
	default:
		panic("not supported")
	}
}

func ArrayToString(array []string) string {
	var result string
	for i := 0; i < len(array); i++ {
		result += array[i] + ","
	}
	if strings.Contains(result, ",") {
		result = result[:strings.LastIndex(result, ",")]
	}
	return result
}

func containArray(tagName string, args []string) bool {
	for i := 0; i < len(args); i++ {
		if tagName == args[i] {
			return true
		}
	}
	return false
}

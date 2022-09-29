package util

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// 用于sqlNull类型与基本数据类型转换的工具
// finished by JiangTao in 2022.07

const TAG = "db"

// N2Basic 将字段类型为sql.NullXXX的实体 转换为字段类型为普通类型，转换过程以普通类型的结构体为主。 tags 为需要跳过转换的字段, 常用于某个字段类型不一致
func N2Basic(a interface{}, b interface{}, tags ...string) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	tb := reflect.TypeOf(b)
	ovb := vb.Elem()
	otb := tb.Elem()
	if va.Kind() != reflect.Struct || otb.Kind() != reflect.Struct {
		log.Panicln("kind is not Struct")
		return
	}
	localN2B(va, ovb, otb, tags...)
}

func N2BasicList(nList interface{}, bList interface{}, tags ...string) {
	vNList := reflect.ValueOf(nList)
	vBList := reflect.ValueOf(bList)
	if !(vBList.Kind() == reflect.Array || vBList.Kind() == reflect.Slice) || !(vNList.Kind() == reflect.Array || vNList.Kind() == reflect.Slice) {
		log.Panicln("kind is not Array or Slice")
		return
	}
	for x := 0; x < vBList.Len(); x++ {
		va := reflect.ValueOf(vNList.Index(x).Interface())
		vbAddr := reflect.ValueOf(vBList.Index(x).Addr().Interface())
		tb := reflect.TypeOf(vBList.Index(x).Interface())
		vb := vbAddr.Elem()
		localN2B(va, vb, tb, tags...)
	}
}

func localN2B(va, ovb reflect.Value, otb reflect.Type, tags ...string) {
	for i := 0; i < otb.NumField(); i++ {
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
		case reflect.Struct:
			flag := false
			vta := va.Type()
			for i := 0; i < vta.NumField(); i++ {
				if vta.Field(i).Name == field.Name {
					flag = true
					break
				}
			}
			if flag {
				x := ovb.FieldByName(field.Name).Addr().Elem()
				y := reflect.TypeOf(ovb.FieldByName(field.Name).Addr().Elem().Interface())
				localN2B(va.FieldByName(field.Name), x, y)
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
	localB2N(va, ovb, otb, tags...)
}

func Basic2NList(bList interface{}, nList interface{}, tags ...string) {
	vBList := reflect.ValueOf(bList)
	vNList := reflect.ValueOf(nList)
	if !(vBList.Kind() == reflect.Array || vBList.Kind() == reflect.Slice) || !(vNList.Kind() == reflect.Array || vNList.Kind() == reflect.Slice) {
		log.Panicln("kind is not Array or Slice")
		return
	}
	for x := 0; x < vBList.Len(); x++ {
		va := reflect.ValueOf(vBList.Index(x).Interface())
		vbAddr := reflect.ValueOf(vNList.Index(x).Addr().Interface())
		otb := reflect.TypeOf(vNList.Index(x).Interface())
		ovb := vbAddr.Elem()
		localB2N(va, ovb, otb, tags...)
	}
}

func localB2N(va, ovb reflect.Value, otb reflect.Type, tags ...string) {
	for i := 0; ovb.Kind() == reflect.Struct && i < ovb.NumField(); i++ {
		field := otb.Field(i)
		if tag := field.Tag.Get(TAG); len(tags) > 0 && containArray(tag, tags) {
			continue
		}
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
		case reflect.Struct:
			x := ovb.FieldByName(field.Name).Addr().Elem()
			y := reflect.TypeOf(ovb.FieldByName(field.Name).Addr().Elem().Interface())
			localB2N(fieldA, x, y)
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

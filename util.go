package sqlxp

import "github.com/jtao539/sqlxp/util"

// N2B nEntity 转换为 bEntity,
// nEntity是包含sql.NullXX类型的结构体 支持string,int,float
// bEntity是golang basic数据类型的结构体,需为指针类型
func N2B(nEntity interface{}, bEntity interface{}, tags ...string) {
	util.N2Basic(nEntity, bEntity, tags...)
}

// N2BList nEntityList 转换为 bEntityList,
// nEntityList是包含nEntity的数组或切片,nEntity是包含sql.NullXX类型的结构体 支持string,int,float
// bEntityList是包含bEntity的数组或切片, bEntity是golang basic数据类型的结构体
// 注意 : nEntityList, bEntityList 均为普通数组类型,无需传地址或指针
func N2BList(nEntityList interface{}, bEntityList interface{}, tags ...string) {
	util.N2BasicList(nEntityList, bEntityList, tags...)
}

// B2N bEntity 转换为 nEntity,
// bEntity是golang basic数据类型的结构体,支持string,int,float
// nEntity是包含sql.NullXX类型的结构体,需为指针类型
func B2N(bEntity interface{}, nEntity interface{}, tags ...string) {
	util.Basic2N(bEntity, nEntity, tags...)
}

// B2NList bEntityList 转换为 nEntityList,
// bEntityList是包含bEntity的数组或切片, bEntity是golang basic数据类型的结构体, 支持string,int,float
// nEntityList是包含nEntity的数组或切片,nEntity是包含sql.NullXX类型的结构体
// 注意 : nEntityList, bEntityList 均为普通数组类型,无需传地址或指针
func B2NList(bEntityList interface{}, nEntityList interface{}, tags ...string) {
	util.Basic2NList(bEntityList, nEntityList, tags...)
}

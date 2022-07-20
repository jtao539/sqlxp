package sqlxp

import "github.com/jtao539/sqlxp/util"

// N2B nEntity 转换为 bEntity,
// nEntity是包含sql.NullXX类型的结构体 支持string,int,float
// bEntity是golang basic数据类型的结构体
func N2B(nEntity interface{}, bEntity interface{}, tags ...string) {
	util.N2Basic(nEntity, bEntity, tags...)
}

// B2N bEntity 转换为 nEntity,
// nEntity是包含sql.NullXX类型的结构体 支持string,int,float
// bEntity是golang basic数据类型的结构体
func B2N(bEntity interface{}, nEntity interface{}, tags ...string) {
	util.Basic2N(bEntity, nEntity, tags...)
}

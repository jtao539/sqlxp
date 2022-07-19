package sqlxp

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	. "github.com/jtao539/sqlxp/sqlFactory"
)

type SqlxP struct {
	DB *sqlx.DB
}

// Select 列表查找, dest为要查找的数据类型数组, request为查询条件结构体, table 为表名称, tags为需要跳过的字段
func (s *SqlxP) Select(dest interface{}, request interface{}, table string, tags ...string) error {
	str, params := SafeSelect(request, table, tags...)
	err := s.DB.Select(dest, str, params...)
	return err
}

// SelectWithFactor 可手动介入查询条件的列表查找, dest为要查找的数据类型数组, request为查询条件结构体, table 为表名称, factors 为sql条件, tags为需要跳过的字段
func (s *SqlxP) SelectWithFactor(dest interface{}, request interface{}, table string, factors []string, tags ...string) error {
	str, params := SafeSelectWithFactor(request, table, factors, tags...)
	err := s.DB.Select(dest, str, params...)
	return err
}

// Update 数据更新, request为新数据的结构体, entity为SQLNULL实体 table 为表名称， 通过对比request和entity获取跳过的字段, tx 为事务支持
func (s *SqlxP) Update(request interface{}, entity interface{}, table string, tx ...*sqlx.Tx) error {
	var err error
	str, params := SafeUpdate(request, entity, table)
	if len(tx) > 0 {
		_, err = tx[0].Exec(str, params...)
	} else {
		_, err = s.DB.Exec(str, params...)
	}
	if err != nil {
		return err
	}
	return err
}

// UpdateP 数据更新, request为新数据的结构体, entity为SQLNULL实体 table 为表名称， fields为需要跳过更新的字段, 通过对比request和entity获取跳过的字段, tx 为事务支持
func (s *SqlxP) UpdateP(request interface{}, entity interface{}, table string, fields []string, tx ...*sqlx.Tx) error {
	var err error
	str, params := SafeUpdateP(request, entity, table, fields...)
	if len(tx) > 0 {
		_, err = tx[0].Exec(str, params...)
	} else {
		_, err = s.DB.Exec(str, params...)
	}
	if err != nil {
		return err
	}
	return err
}

func (s *SqlxP) GetOneById(one interface{}, table string, id int) error {
	str := fmt.Sprintf("select * from %s where id=?", table)
	return s.DB.Get(one, str, id)
}

func (s *SqlxP) InsertOne(one interface{}, table string, tx ...*sqlx.Tx) error {
	var err error
	var rows sql.Result
	str := SafeInsert(one, table)
	if len(tx) > 0 {
		rows, err = tx[0].NamedExec(str, one)
	} else {
		rows, err = s.DB.NamedExec(str, one)
	}
	if err != nil {
		return err
	}
	AffectedNum, _ := rows.RowsAffected()
	if AffectedNum == 0 {
		return InsertError
	}
	return err
}

func (s *SqlxP) DeleteOneById(table string, id int, tx ...*sqlx.Tx) error {
	var err error
	var rows sql.Result
	str := fmt.Sprintf("delete from %s where id = ?", table)
	if len(tx) > 0 {
		rows, err = tx[0].Exec(str, id)
	} else {
		rows, err = s.DB.Exec(str, id)
	}
	if err != nil {
		return err
	}
	AffectedNum, _ := rows.RowsAffected()
	if AffectedNum == 0 {
		return DeleteError
	}
	return err
}

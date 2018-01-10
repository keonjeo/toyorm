package toyorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type MySqlDialect struct {
}

func (dia MySqlDialect) HasTable(model *Model) ExecValue {
	return ExecValue{
		"SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE  table_schema = (SELECT DATABASE()) AND table_name = ?",
		[]interface{}{model.Name},
	}
}

func (dia MySqlDialect) CreateTable(model *Model) (execlist []ExecValue) {
	// lazy init model
	fieldStrList := []string{}
	for _, sqlField := range model.GetSqlFields() {
		s := fmt.Sprintf("%s %s", sqlField.Name, sqlField.Type)
		if sqlField.AutoIncrement {
			s += " AUTO_INCREMENT"
		}
		for k, v := range sqlField.CommonAttr {
			if v == "" {
				s += " " + k
			} else {
				s += " " + fmt.Sprintf("%s=%s", k, v)
			}
		}
		fieldStrList = append(fieldStrList, s)
	}
	primaryStrList := []string{}
	for _, p := range model.GetPrimary() {
		primaryStrList = append(primaryStrList, p.Name)
	}
	sqlStr := fmt.Sprintf("CREATE TABLE %s (%s, PRIMARY KEY(%s))",
		model.Name,
		strings.Join(fieldStrList, ","),
		strings.Join(primaryStrList, ","),
	)
	execlist = append(execlist, ExecValue{sqlStr, nil})

	indexStrList := []string{}
	for key, fieldList := range model.GetIndexMap() {
		fieldStrList := []string{}
		for _, f := range fieldList {
			fieldStrList = append(fieldStrList, f.Name)
		}
		indexStrList = append(indexStrList, fmt.Sprintf("CREATE INDEX %s ON %s(%s)", key, model.Name, strings.Join(fieldStrList, ",")))
	}
	uniqueIndexStrList := []string{}
	for key, fieldList := range model.GetUniqueIndexMap() {
		fieldStrList := []string{}
		for _, f := range fieldList {
			fieldStrList = append(fieldStrList, f.Name)
		}
		uniqueIndexStrList = append(uniqueIndexStrList, fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s(%s)", key, model.Name, strings.Join(fieldStrList, ",")))
	}
	for _, indexStr := range indexStrList {
		execlist = append(execlist, ExecValue{indexStr, nil})
	}
	for _, indexStr := range uniqueIndexStrList {
		execlist = append(execlist, ExecValue{indexStr, nil})
	}
	return
}

func (dia MySqlDialect) ToSqlType(_type reflect.Type) (sqlType string) {
	switch _type.Kind() {
	case reflect.Ptr:
		sqlType = dia.ToSqlType(_type.Elem())
	case reflect.Bool:
		sqlType = "BOOLEAN"
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		sqlType = "INTEGER"
	case reflect.Int64, reflect.Uint64, reflect.Int, reflect.Uint:
		sqlType = "BIGINT"
	case reflect.Float32, reflect.Float64:
		sqlType = "FLOAT"
	case reflect.String:
		sqlType = "VARCHAR(255)"
	case reflect.Struct:
		if _, ok := reflect.New(_type).Elem().Interface().(time.Time); ok {
			sqlType = "TIMESTAMP"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullBool); ok {
			sqlType = "BOOLEAN"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullInt64); ok {
			sqlType = "BIGINT"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullString); ok {
			sqlType = "VARCHAR(255)"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.NullFloat64); ok {
			sqlType = "FLOAT"
		} else if _, ok := reflect.New(_type).Elem().Interface().(sql.RawBytes); ok {
			sqlType = "VARCHAR(255)"
		}
	default:
		if _, ok := reflect.New(_type).Elem().Interface().([]byte); ok {
			sqlType = "VARCHAR(255)"
		}
	}
	return
}

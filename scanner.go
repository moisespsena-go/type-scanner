package scanner

import (
	"database/sql"
	"os"
	"reflect"
)

type Scanner interface {
	Bulk(rows *sql.Rows, do func(v interface{}) error) (err error)
	One(rows *sql.Rows, dst interface{}) (err error)
}

type Scan struct {
	Typ           reflect.Type
	Ptr           bool
	NameConverter NameConverter
}

func New(typ interface{}) (s Scan) {
	switch t := typ.(type) {
	case reflect.Type:
		s.Typ = t
	default:
		s.Typ = reflect.TypeOf(typ)
	}
	for s.Typ.Kind() == reflect.Ptr {
		s.Ptr = true
		s.Typ = s.Typ.Elem()
	}
	return
}

func (this Scan) Fields(columns ...string) (fields [][]int) {
	nc := this.NameConverter
	if nc == nil {
		nc = DefaultNameConverter
	}

	fields = make([][]int, len(columns))
	for i, col := range columns {
		fname := nc.Convert(col, FakeNameConverter)
		if f, ok := this.Typ.FieldByName(fname); ok {
			fields[i] = f.Index
		}
	}
	return
}

func (this Scan) Of(value reflect.Value, fields ...[]int) (recorde interface{}, args []interface{}) {
	args = make([]interface{}, len(fields))

	for i, index := range fields {
		if index == nil {
			args[i] = discardScan{}
		} else {
			args[i] = value.FieldByIndex(index).Addr().Interface()
		}
	}

	if this.Ptr {
		recorde = value.Addr().Interface()
	} else {
		recorde = value.Interface()
	}
	return
}

func (this Scan) New(fields ...[]int) (recorde interface{}, args []interface{}) {
	return this.Of(reflect.New(this.Typ).Elem(), fields...)
}

func (this Scan) Bulk(rows *sql.Rows, do func(v interface{}) error) (err error) {
	columns, _ := rows.Columns()
	if len(columns) == 0 {
		return os.ErrNotExist
	}

	fields := this.Fields(columns...)

	for rows.Next() {
		record, args := this.New(fields...)
		if err = rows.Scan(args...); err != nil {
			return
		}
		if err = do(record); err != nil {
			return
		}
	}
	return
}

func (this Scan) One(rows *sql.Rows, dst interface{}) (err error) {
	columns, _ := rows.Columns()
	if len(columns) == 0 {
		return os.ErrNotExist
	}

	fields := this.Fields(columns...)

	if rows.Next() {
		_, args := this.Of(reflect.ValueOf(dst).Elem(), fields...)
		return rows.Scan(args...)
	} else {
		return os.ErrNotExist
	}
}

type discardScan struct{}

func (discardScan) Scan(src interface{}) error {
	return nil
}

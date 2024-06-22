package goe

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
)

func handlerValues(conn Connection, sqlQuery string, args []any) {
	_, err := conn.ExecContext(context.Background(), sqlQuery, args...)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handlerValuesReturning(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string) {
	row := conn.QueryRowContext(context.Background(), sqlQuery, args...)

	//TODO: Better returning handler
	targetId := value.FieldByName(idName)
	id := returnTarget(targetId)
	err := row.Scan(id)
	if err != nil {
		fmt.Println(err)
		return
	}

	targetId.Set(reflect.ValueOf(id).Elem())
}

func returnTarget(targetId reflect.Value) any {
	switch targetId.Kind() {
	case reflect.Uint64:
		return new(uint64)
	case reflect.Uint32:
		return new(uint32)
	case reflect.Uint16:
		return new(uint16)
	case reflect.Uint8:
		return new(uint8)
	case reflect.Uint:
		return new(uint)
	case reflect.Int:
		return new(int)
	case reflect.Int8:
		return new(int8)
	case reflect.Int16:
		return new(int16)
	case reflect.Int32:
		return new(int32)
	case reflect.Int64:
		return new(int64)
	case reflect.String:
		return new(string)
	case reflect.Slice:
		return new([]byte)
	default:
		return new(any)
	}
}

func handlerResult(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string) {
	switch value.Kind() {
	case reflect.Slice:
		handlerQuery(conn, sqlQuery, value, args, structColumns)
	case reflect.Struct:
		//handlerQueryRow(conn, sqlQuery, value, args)
		fmt.Println("struct")
	default:
		fmt.Println("default")
	}
}

func handlerQuery(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string) {
	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)

	//TODO: Better error
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	//Prepare dest for query
	c, err := rows.Columns()

	if err != nil {
		fmt.Println(err)
		return
	}

	dest := make([]any, len(c))
	for i := range dest {
		dest[i] = new(sql.RawBytes)
	}

	//Check the result target
	switch value.Type().Elem().Kind() {
	case reflect.Struct:
		err = mapStructQuery(rows, dest, value, structColumns)

		//TODO: Better error
		if err != nil {
			fmt.Println(err)
			return
		}
	default:
		err = mapQuery(rows, dest, value)

		//TODO: Better error
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func mapStructQuery(rows *sql.Rows, dest []any, value reflect.Value, columns []string) (err error) {
	//TODO: add count for slices
	value.Set(reflect.MakeSlice(value.Type(), 0, 0))
	for i := 0; rows.Next(); i++ {
		err = rows.Scan(dest...)
		if err != nil {
			return err
		}
		s := reflect.New(value.Type().Elem()).Elem()
		//Fills the target
		for i, a := range dest {
			setValue(s.FieldByName(columns[i]), a)
		}
		value.Set(reflect.Append(value, s))
	}
	return err
}

func mapQuery(rows *sql.Rows, dest []any, value reflect.Value) (err error) {
	//TODO: Len of slice be the size of the query
	value.Set(reflect.MakeSlice(value.Type(), 0, 0))

	for i := 0; rows.Next(); i++ {
		err = rows.Scan(dest...)
		if err != nil {
			return err
		}
		s := reflect.New(value.Type().Elem()).Elem()
		for _, a := range dest {
			setValue(s, a)
		}
		value.Set(reflect.Append(value, s))
	}
	return err
}

func setValue(v reflect.Value, a any) {
	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(string(*(a.(*sql.RawBytes))))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(parseInt(*(a.(*sql.RawBytes))))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(parseUint(*(a.(*sql.RawBytes))))
	case reflect.Bool:
		v.SetBool((*(a.(*sql.RawBytes)))[0] == 116)
	case reflect.Slice:
		v.SetBytes((*(a.(*sql.RawBytes))))
	case reflect.Struct:
		var tm time.Time
		tm.UnmarshalText(*a.(*sql.RawBytes))
		v.Set(reflect.ValueOf((tm)))
	case reflect.Float64, reflect.Float32:
		f, _ := strconv.ParseFloat(string((*(a.(*sql.RawBytes)))), 32)
		v.SetFloat(f)
	}
}

func parseInt(bts []byte) int64 {
	var n int64
	var neg bool
	if bts[0] == '-' {
		neg = true
		bts = bts[1:]
	}
	for _, b := range bts {
		d := b - '0' //reduces the byte by the rune 0, if the byte is digit 0 will be: 48 - 48
		n *= int64(10)
		n1 := n + int64(d)
		n = n1
	}
	if neg {
		n = -n
	}
	return n
}

func parseUint(bts []byte) uint64 {
	var n uint64
	for _, b := range bts {
		d := b - '0' //reduces the byte by the rune 0, if the byte is digit 0 will be: 48 - 48
		n *= uint64(10)
		n1 := n + uint64(d)
		n = n1
	}
	return n
}

func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

package goe

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

func handlerValues(conn Connection, sqlQuery string, args []any) {
	defer conn.Close()

	_, err := conn.ExecContext(context.Background(), sqlQuery, args...)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handlerValuesReturning(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string) {
	defer conn.Close()

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
	defer conn.Close()

	switch value.Kind() {
	case reflect.Slice:
		handlerQuery(conn, sqlQuery, value, args, structColumns)
	case reflect.Struct:
		handlerStructQueryRow(conn, sqlQuery, value, args, structColumns)
	default:
		handlerQueryRow(conn, sqlQuery, value, args, structColumns)
	}
}

func handlerQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any, columns []string) {
	dest := make([]any, len(columns))
	for i := range dest {
		dest[i] = new(any)
	}
	err := conn.QueryRowContext(context.Background(), sqlQuery, args...).Scan(dest...)
	if err != nil {
		fmt.Println(err)
		return
	}

	setValue(value, dest[0])
}

func handlerStructQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any, columns []string) {
	dest := make([]any, len(columns))
	for i := range dest {
		dest[i] = new(any)
	}
	err := conn.QueryRowContext(context.Background(), sqlQuery, args...).Scan(dest...)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i, a := range dest {
		setValue(value.FieldByName(columns[i]), a)
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

	dest := make([]any, len(structColumns))
	for i := range dest {
		dest[i] = new(any)
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
	//TODO: Change sql.RawBytes to *any
	switch v.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt((*(a).(*any)).(int64))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint((*(a).(*any)).(uint64))
	case reflect.Float64, reflect.Float32:
		v.SetFloat((*(a).(*any)).(float64))
	default:
		v.Set(reflect.ValueOf(*(a).(*any)))
	}
}

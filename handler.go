package goe

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
)

func handlerValues(conn Connection, sqlQuery string, args []any) error {
	defer conn.Close()

	_, err := conn.ExecContext(context.Background(), sqlQuery, args...)
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturning(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string) error {
	defer conn.Close()

	row := conn.QueryRowContext(context.Background(), sqlQuery, args...)

	err := row.Scan(value.FieldByName(idName).Addr().Interface())
	if err != nil {
		return err
	}
	return nil
}

func handlerValuesReturningBatch(conn Connection, sqlQuery string, value reflect.Value, args []any, idName string) error {
	defer conn.Close()

	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)
	if err != nil {
		return err
	}

	i := 0
	for rows.Next() {
		err = rows.Scan(value.Index(i).FieldByName(idName).Addr().Interface())
		if err != nil {
			return err
		}
		i++
	}
	return nil
}

func handlerResult(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string, limit uint) error {
	defer conn.Close()

	switch value.Kind() {
	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Struct && value.Type().Elem().Name() != "Time" {
			return handlerStructQuery(conn, sqlQuery, value, args, structColumns, limit)
		}
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return handlerQueryRow(conn, sqlQuery, value, args)
		}
		return handlerQuery(conn, sqlQuery, value, args, limit)
	case reflect.Struct:
		if value.Type().Name() != "Time" {
			return handlerStructQueryRow(conn, sqlQuery, value, args, structColumns)
		}
		return handlerQueryRow(conn, sqlQuery, value, args)
	default:
		return handlerQueryRow(conn, sqlQuery, value, args)
	}
}

func handlerQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any) error {
	dest := make([]any, 1)
	for i := range dest {
		dest[i] = value.Addr().Interface()
	}
	err := conn.QueryRowContext(context.Background(), sqlQuery, args...).Scan(dest...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	value.Set(reflect.ValueOf(dest[0]).Elem())
	return nil
}

func handlerStructQueryRow(conn Connection, sqlQuery string, value reflect.Value, args []any, columns []string) error {
	dest := make([]any, len(columns))
	for i := range dest {
		t, _ := value.Type().FieldByName(columns[i])
		if t.Type == nil {
			dest[i] = reflect.New(value.Type().Field(i).Type).Interface()
			continue
		}
		dest[i] = reflect.New(t.Type).Interface()
	}
	err := conn.QueryRowContext(context.Background(), sqlQuery, args...).Scan(dest...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var field reflect.Value
	for i, a := range dest {
		field = value.FieldByName(columns[i])
		if !field.CanSet() {
			field = value.Field(i)
		}
		field.Set(reflect.ValueOf(a).Elem())
	}
	return nil
}

func handlerQuery(conn Connection, sqlQuery string, value reflect.Value, args []any, limit uint) error {
	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)

	if err != nil {
		return err
	}
	defer rows.Close()

	dest := make([]any, 1)
	for i := range dest {
		dest[i] = reflect.New(value.Type().Elem()).Interface()
	}

	err = mapQuery(rows, dest, value, limit)

	if err != nil {
		return err
	}
	return nil
}

func mapQuery(rows *sql.Rows, dest []any, value reflect.Value, limit uint) (err error) {
	value.Set(reflect.MakeSlice(value.Type(), 0, int(limit)))

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			return err
		}
		value.Set(reflect.Append(value, reflect.ValueOf(dest[0]).Elem()))
	}
	return err
}

func handlerStructQuery(conn Connection, sqlQuery string, value reflect.Value, args []any, structColumns []string, limit uint) error {
	rows, err := conn.QueryContext(context.Background(), sqlQuery, args...)

	if err != nil {
		return err
	}
	defer rows.Close()

	dest := make([]any, len(structColumns))
	for i := range dest {
		if t, ok := value.Type().Elem().FieldByName(structColumns[i]); ok {
			dest[i] = reflect.New(t.Type).Interface()
			continue
		}
		dest[i] = reflect.New(value.Type().Elem().Field(i).Type).Interface()
	}

	err = mapStructQuery(rows, dest, value, structColumns, limit)

	if err != nil {
		return err
	}
	return nil
}

func mapStructQuery(rows *sql.Rows, dest []any, value reflect.Value, columns []string, limit uint) (err error) {
	value.Set(reflect.MakeSlice(value.Type(), 0, int(limit)))
	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			return err
		}
		s := reflect.New(value.Type().Elem()).Elem()
		var f reflect.Value
		//Fills the target
		for i, a := range dest {
			f = s.FieldByName(columns[i])
			if !f.CanSet() {
				f = s.Field(i)
			}
			f.Set(reflect.ValueOf(a).Elem())
		}
		value.Set(reflect.Append(value, s))
	}
	return err
}

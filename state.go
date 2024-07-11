package goe

import (
	"errors"
	"reflect"
)

var ErrInvalidScan = errors.New("goe: invalid scan target. try sending a pointer for scan")

var ErrInvalidInsertValue = errors.New("goe: invalid insert value. try sending a pointer as value")
var ErrInvalidInsertInValue = errors.New("goe: invalid insertIn value. try sending only two values or a size even slice")

var ErrInvalidUpdateValue = errors.New("goe: invalid update value. try sending a struct or a pointer to struct as value")

type stateSelect struct {
	conn    Connection
	addrMap map[uintptr]field
	builder *builder
	err     error
}

func createSelectState(conn Connection, e error) *stateSelect {
	return &stateSelect{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateSelect) Where(brs ...operator) *stateSelect {
	s.builder.brs = brs
	return s
}

func (s *stateSelect) Join(t1 any, t2 any) *stateSelect {
	args, err := getArgsIn(t1, t2)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.buildSelectJoins(s.addrMap, "JOIN", args)
	return s
}

func (s *stateSelect) InnerJoin(t1 any, t2 any) *stateSelect {
	args, err := getArgsIn(t1, t2)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.buildSelectJoins(s.addrMap, "INNER JOIN", args)
	return s
}

func (s *stateSelect) RightJoin(t1 any, t2 any) *stateSelect {
	args, err := getArgsIn(t1, t2)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.buildSelectJoins(s.addrMap, "RIGHT JOIN", args)
	return s
}

func (s *stateSelect) LeftJoin(t1 any, t2 any) *stateSelect {
	args, err := getArgsIn(t1, t2)
	if err != nil {
		s.err = err
		return s
	}
	s.builder.buildSelectJoins(s.addrMap, "LEFT JOIN", args)
	return s
}

func (s *stateSelect) querySelect(args []uintptr) *stateSelect {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildSelect(s.addrMap)
	}
	return s
}

func (s *stateSelect) Scan(target any) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		return "", ErrInvalidScan
	}

	//generate query
	s.builder.buildSqlSelect()

	sql := s.builder.sql.String()
	handlerResult(s.conn, sql, value.Elem(), s.builder.argsAny, s.builder.structColumns)
	return sql, nil
}

/*
State Insert
*/
type stateInsert struct {
	conn    Connection
	builder *builder
	err     error
}

func createInsertState(conn Connection, e error) *stateInsert {
	return &stateInsert{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateInsert) queryInsert(args []uintptr, addrMap map[uintptr]field) *stateInsert {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildInsert(addrMap)
	}
	return s
}

func (s *stateInsert) Value(target any) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		return "", ErrInvalidInsertValue
	}

	value = value.Elem()

	if value.Kind() == reflect.Slice {
		return s.batchValue(value)
	}

	idName := s.builder.buildValues(value)

	sql := s.builder.sql.String()
	handlerValuesReturning(s.conn, sql, value, s.builder.argsAny, idName)
	return sql, nil
}

func (s *stateInsert) batchValue(value reflect.Value) (string, error) {
	idName := s.builder.buildBatchValues(value)

	sql := s.builder.sql.String()
	handlerValuesReturningBatch(s.conn, sql, value, s.builder.argsAny, idName)
	return sql, nil
}

type stateInsertIn struct {
	conn    Connection
	builder *builder
	err     error
}

func createInsertStateIn(conn Connection, e error) *stateInsertIn {
	return &stateInsertIn{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateInsertIn) queryInsertIn(args []uintptr, addrMap map[uintptr]field) *stateInsertIn {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildInsertIn(addrMap)
	}
	return s
}

func (s *stateInsertIn) Values(v ...any) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	switch len(v) {
	case 1:
		value := reflect.ValueOf(v[0])
		if value.Kind() == reflect.Pointer {
			value = value.Elem()
		}
		if value.Kind() != reflect.Slice || value.Len() < 2 {
			return "", ErrInvalidInsertInValue
		}

		s.builder.buildValuesInBatch(value)

		sql := s.builder.sql.String()
		handlerValues(s.conn, sql, s.builder.argsAny)
		return sql, nil
	case 2:
		s.builder.argsAny = append(s.builder.argsAny, v[0])
		s.builder.argsAny = append(s.builder.argsAny, v[1])

		s.builder.buildValuesIn()

		sql := s.builder.sql.String()
		handlerValues(s.conn, sql, s.builder.argsAny)
		return sql, nil
	default:
		return "", ErrInvalidInsertInValue
	}
}

/*
State Update
*/
type stateUpdate struct {
	conn    Connection
	builder *builder
	err     error
}

func createUpdateState(conn Connection, e error) *stateUpdate {
	return &stateUpdate{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateUpdate) Where(brs ...operator) *stateUpdate {
	s.builder.brs = brs
	return s
}

func (s *stateUpdate) queryUpdate(args []uintptr, addrMap map[uintptr]field) *stateUpdate {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildUpdate(addrMap)
	}
	return s
}

func (s *stateUpdate) Value(target any) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	value := reflect.ValueOf(target)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return "", ErrInvalidUpdateValue
	}

	s.builder.buildSet(value)

	//generate query
	s.builder.buildSqlUpdate()

	sql := s.builder.sql.String()
	handlerValues(s.conn, sql, s.builder.argsAny)
	return sql, nil
}

type stateUpdateIn struct {
	conn    Connection
	builder *builder
	err     error
}

func createUpdateInState(conn Connection, e error) *stateUpdateIn {
	return &stateUpdateIn{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateUpdateIn) Where(brs ...operator) *stateUpdateIn {
	s.builder.brs = brs
	return s
}

func (s *stateUpdateIn) queryUpdateIn(args []uintptr, addrMap map[uintptr]field) *stateUpdateIn {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildUpdateIn(addrMap)
	}
	return s
}

func (s *stateUpdateIn) Value(value any) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	s.builder.argsAny = append(s.builder.argsAny, value)

	s.builder.buildSetIn()

	s.builder.buildSqlUpdateIn()

	sql := s.builder.sql.String()
	handlerValues(s.conn, sql, s.builder.argsAny)
	return sql, nil
}

type stateDelete struct {
	conn    Connection
	builder *builder
	err     error
}

func createDeleteState(conn Connection, e error) *stateDelete {
	return &stateDelete{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateDelete) queryDelete(args []uintptr, addrMap map[uintptr]field) *stateDelete {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildDelete(addrMap)
	}
	return s
}

func (s *stateDelete) Where(brs ...operator) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	s.builder.brs = brs

	s.builder.buildSqlDelete()

	sql := s.builder.sql.String()
	handlerValues(s.conn, sql, s.builder.argsAny)
	return sql, nil
}

type stateDeleteIn struct {
	conn    Connection
	builder *builder
	err     error
}

func createDeleteInState(conn Connection, e error) *stateDeleteIn {
	return &stateDeleteIn{conn: conn, builder: createBuilder(), err: e}
}

func (s *stateDeleteIn) queryDeleteIn(args []uintptr, addrMap map[uintptr]field) *stateDeleteIn {
	if s.err == nil {
		s.builder.args = args
		s.builder.buildDeleteIn(addrMap)
	}
	return s
}

func (s *stateDeleteIn) Where(brs ...operator) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	s.builder.brs = brs

	s.builder.buildSqlDeleteIn()

	sql := s.builder.sql.String()
	handlerValues(s.conn, sql, s.builder.argsAny)
	return sql, nil
}

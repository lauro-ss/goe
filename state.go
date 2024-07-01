package goe

import (
	"fmt"
	"reflect"
)

type stateSelect struct {
	conn    Connection
	addrMap map[string]field
	builder *builder
}

func createSelectState(conn Connection, qt int8) *stateSelect {
	return &stateSelect{conn: conn, builder: createBuilder(qt)}
}

func (s *stateSelect) Where(brs ...operator) *stateSelect {
	where(s.builder, brs...)
	return s
}

func (s *stateSelect) Join(tables ...any) *stateSelect {
	s.builder.args = append(s.builder.args, getArgs(s.addrMap, tables...)...)
	s.builder.buildSelectJoins(s.addrMap)
	return s
}

func (s *stateSelect) querySelect(args []string) *stateSelect {
	s.builder.args = args
	s.builder.buildSelect(s.addrMap)
	return s
}

func (s *stateSelect) Result(target any) {
	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		fmt.Printf("%v: target result needs to be a pointer try &animals\n", pkg)
		return
	}

	//generate query
	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	handlerResult(s.conn, s.builder.sql.String(), value.Elem(), s.builder.argsAny, s.builder.structColumns)
}

/*
State Insert
*/
type stateInsert struct {
	conn    Connection
	builder *builder
}

func createInsertState(conn Connection, qt int8) *stateInsert {
	return &stateInsert{conn: conn, builder: createBuilder(qt)}
}

func (s *stateInsert) queryInsert(args []string, addrMap map[string]field) *stateInsert {
	s.builder.args = args
	s.builder.buildInsert(addrMap)
	return s
}

func (s *stateInsert) Value(target any) {
	value := reflect.ValueOf(target)

	//TODO: Handler value as struct or ptr
	if value.Kind() != reflect.Ptr {
		fmt.Printf("%v: target result needs to be a pointer try &animals\n", pkg)
		return
	}

	value = value.Elem()

	idName := s.builder.buildValues(value, s.builder.targetFksNames)

	//generate query
	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	handlerValuesReturning(s.conn, s.builder.sql.String(), value, s.builder.argsAny, idName)
}

type stateInsertIn struct {
	conn    Connection
	builder *builder
}

func createInsertStateIn(conn Connection, qt int8) *stateInsertIn {
	return &stateInsertIn{conn: conn, builder: createBuilder(qt)}
}

func (s *stateInsertIn) queryInsertIn(args []string, addrMap map[string]field) *stateInsertIn {
	s.builder.args = args
	s.builder.buildInsertIn(addrMap)
	return s
}

func (s *stateInsertIn) Values(v1 any, v2 any) {
	s.builder.argsAny = append(s.builder.argsAny, v1)
	s.builder.argsAny = append(s.builder.argsAny, v2)

	s.builder.buildValuesIn()

	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	handlerValues(s.conn, s.builder.sql.String(), s.builder.argsAny)
}

/*
State Update
*/
type stateUpdate struct {
	conn    Connection
	builder *builder
}

func createUpdateState(conn Connection, qt int8) *stateUpdate {
	return &stateUpdate{conn: conn, builder: createBuilder(qt)}
}

func (s *stateUpdate) Where(brs ...operator) *stateUpdate {
	where(s.builder, brs...)
	return s
}

func (s *stateUpdate) queryUpdate(args []string, addrMap map[string]field) *stateUpdate {
	s.builder.args = args
	s.builder.buildUpdate(addrMap)
	return s
}

func (s *stateUpdate) Value(target any) {
	value := reflect.ValueOf(target)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		fmt.Printf("%v: value for update needs to be a struct\n", pkg)
		return
	}

	s.builder.buildSet(value, s.builder.targetFksNames, s.builder.structColumns)

	//generate query
	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	handlerValues(s.conn, s.builder.sql.String(), s.builder.argsAny)

}

type stateUpdateIn struct {
	conn    Connection
	builder *builder
}

func createUpdateInState(conn Connection, qt int8) *stateUpdateIn {
	return &stateUpdateIn{conn: conn, builder: createBuilder(qt)}
}

func (s *stateUpdateIn) Where(brs ...operator) *stateUpdateIn {
	where(s.builder, brs...)
	return s
}

func (s *stateUpdateIn) queryUpdateIn(args []string, addrMap map[string]field) *stateUpdateIn {
	s.builder.args = args
	s.builder.buildUpdateIn(addrMap)
	return s
}

func (s *stateUpdateIn) Value(value any) {
	s.builder.argsAny = append(s.builder.argsAny, value)

	s.builder.buildSetIn()

	s.builder.buildSqlUpdateIn()

	fmt.Println(s.builder.sql)
	handlerValues(s.conn, s.builder.sql.String(), s.builder.argsAny)
}

type stateDelete struct {
	conn    Connection
	builder *builder
}

func createDeleteState(conn Connection, qt int8) *stateDelete {
	return &stateDelete{conn: conn, builder: createBuilder(qt)}
}

func (s *stateDelete) queryDelete(args []string, addrMap map[string]field) *stateDelete {
	s.builder.args = args
	s.builder.buildDelete(addrMap)
	return s
}

func (s *stateDelete) Where(brs ...operator) {
	where(s.builder, brs...)

	s.builder.buildSqlDelete()

	fmt.Println(s.builder.sql)
	handlerValues(s.conn, s.builder.sql.String(), s.builder.argsAny)
}

type stateDeleteIn struct {
	conn    Connection
	builder *builder
}

func createDeleteInState(conn Connection, qt int8) *stateDeleteIn {
	return &stateDeleteIn{conn: conn, builder: createBuilder(qt)}
}

func (s *stateDeleteIn) queryDeleteIn(args []string, addrMap map[string]field) *stateDeleteIn {
	s.builder.args = args
	s.builder.buildDeleteIn(addrMap)
	return s
}

func (s *stateDeleteIn) Where(brs ...operator) {
	where(s.builder, brs...)

	s.builder.buildSqlDeleteIn()

	fmt.Println(s.builder.sql)
	handlerValues(s.conn, s.builder.sql.String(), s.builder.argsAny)
}

func where(builder *builder, brs ...operator) {
	builder.brs = brs
	for _, br := range builder.brs {
		if op, ok := br.(complexOperator); ok {
			builder.tables.add(createStatement(op.pk.table, writeTABLE))
			builder.pks.add(op.pk)
		}
	}
}

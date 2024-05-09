package goe

import (
	"fmt"
	"reflect"
)

type state struct {
	conn    conn
	builder *builder
}

func createState(conn conn, qt int8) *state {
	return &state{conn: conn, builder: createBuilder(qt)}
}

func (s *state) Where(brs ...*booleanResult) State {
	s.builder.brs = brs
	for _, br := range s.builder.brs {
		switch br.tip {
		case EQUALS:
			s.builder.tables.add(createStatement(br.pk.table, writeTABLE))
			s.builder.pks.add(br.pk)
		}
	}
	return s
}

func (s *state) Result(target ...any) {
	switch s.builder.queryType {
	case querySELECT:
		s.resultSelect(target...)
	case queryINSERT:
		s.resultInsert(target...)
	case queryUPDATE:
		s.resultUpdate(target...)
	case queryDELETE:
	}
}

func (s *state) resultSelect(target ...any) {
	if len(target) != 1 {
		fmt.Printf("%v: invalid select query. try using only one target value\n", pkg)
		return
	}
	value := reflect.ValueOf(target[0])

	if value.Kind() != reflect.Ptr {
		fmt.Printf("%v: target result needs to be a pointer try &animals\n", pkg)
		return
	}

	//generate query
	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	handlerResult(s.conn, s.builder.sql.String(), value.Elem(), s.builder.argsAny)
}

func (s *state) resultInsert(target ...any) {
	switch len(target) {
	case 1:
		s.valueInsert(target[0])
	case 2:
		s.values(target[0], target[1])
	default:
		fmt.Printf("%v: invalid insert query\n", pkg)
		return
	}
}

func (s *state) resultUpdate(target ...any) {
	switch len(target) {
	case 1:
		s.valueUpdate(target[0])
	case 2:
		s.values(target[0], target[1])
	default:
		fmt.Printf("%v: invalid update query\n", pkg)
		return
	}
}

func (s *state) querySelect(args []string, addrMap map[string]any) State {
	s.builder.args = args
	s.builder.buildSelect(addrMap)
	return s
}

func (s *state) valueInsert(target any) {
	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Ptr {
		fmt.Printf("%v: target result needs to be a pointer try &animals\n", pkg)
		return
	}

	value = value.Elem()

	idName := s.builder.buildValues(value)

	//generate query
	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	handlerValuesReturning(s.conn, s.builder.sql.String(), value, s.builder.argsAny, idName)
}

func (s *state) queryInsert(args []string, addrMap map[string]any) State {
	s.builder.args = args
	s.builder.buildInsert(addrMap)
	return s
}

func (s *state) valueUpdate(target any) {
	value := reflect.ValueOf(target)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		fmt.Printf("%v: value for update needs to be a struct\n", pkg)
		return
	}

	s.builder.buildSet(value)

	//generate query
	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	handlerValues(s.conn, s.builder.sql.String(), s.builder.argsAny)

}

func (s *state) queryUpdate(args []string, addrMap map[string]any) State {
	s.builder.args = args
	s.builder.buildUpdate(addrMap)
	return s
}

func (s *state) values(v1 any, v2 any) {
	s.builder.argsAny = append(s.builder.argsAny, v1)
	s.builder.argsAny = append(s.builder.argsAny, v2)

	s.builder.buildValuesManyToMany()

	s.builder.buildSql()

	fmt.Println(s.builder.sql)
	handlerValues(s.conn, s.builder.sql.String(), s.builder.argsAny)
}

func (s *state) queryInsertManyToMany(args []string, addrMap map[string]any) State {
	s.builder.args = args
	s.builder.buildInsertManyToMany(addrMap)
	return s
}

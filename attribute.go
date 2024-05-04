package goe

type manyToMany struct {
	table string
	ids   map[string]string
}

type manyToOne struct {
	id      string
	hasMany bool
}

type pk struct {
	table         string
	selectName    string
	attributeName string
	autoIncrement bool
	fks           map[string]any
}

func createPk(table string, selectName string, attributeName string, autoIncrement bool) *pk {
	return &pk{
		table:         table,
		selectName:    selectName,
		attributeName: attributeName,
		autoIncrement: autoIncrement,
		fks:           make(map[string]any)}
}

type att struct {
	selectName    string
	attributeName string
	pk            *pk
}

func createAtt(selectName string, attributeName string, pk *pk) *att {
	return &att{selectName: selectName, attributeName: attributeName, pk: pk}
}

const (
	EQUALS = 1
)

type booleanResult struct {
	arg   string
	pk    *pk
	value any
	tip   int8
}

func createBooleanResult(arg string, pk *pk, value any, tip int8) *booleanResult {
	return &booleanResult{arg: arg, pk: pk, value: value, tip: tip}
}

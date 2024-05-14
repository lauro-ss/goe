package goe

type manyToMany struct {
	table string
	ids   map[string]attributeStrings
}

type manyToOne struct {
	id      string
	hasMany bool
}

type attributeStrings struct {
	selectName    string
	attributeName string
}

func createAttributeStrings(selectName string, attributeName string) attributeStrings {
	return attributeStrings{selectName: selectName, attributeName: attributeName}
}

type pk struct {
	table         string
	autoIncrement bool
	fks           map[string]any
	attributeStrings
}

func createPk(table string, selectName string, attributeName string, autoIncrement bool) *pk {
	return &pk{
		table:            table,
		attributeStrings: createAttributeStrings(selectName, attributeName),
		autoIncrement:    autoIncrement,
		fks:              make(map[string]any)}
}

type att struct {
	attributeStrings
	pk *pk
}

func createAtt(selectName string, attributeName string, pk *pk) *att {
	return &att{
		attributeStrings: createAttributeStrings(selectName, attributeName), pk: pk}
}

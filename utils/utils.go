package utils

import "strings"

func TableNamePattern(name string) string {
	name += "s"
	return strings.ToLower(name)
}

func ColumnNamePattern(name string) string {
	return strings.ToLower(name)
}

func ManyToManyNamePattern(column, table string) string {
	return strings.ToLower(column + table)
}

func ManyToOneNamePattern(column, table string) string {
	return ManyToManyNamePattern(column, table)
}

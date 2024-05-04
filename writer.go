package goe

import (
	"strings"
)

const (
	writeDML    int8 = 1 //DML as SELECT, INSERT, UPDATE and DELETE
	writeATT    int8 = 2 //Attribute
	writeTABLE  int8 = 3
	writeJOIN   int8 = 4
	writeMIDDLE int8 = 6
)

func writeSelect(sql *strings.Builder, q *statementQueue) {
	if q.head != nil {
		writeSelectStatement(sql, q.head)
		q.head = q.head.next
		q.size--
		writeSelect(sql, q)
		return
	}

	sql.WriteString(";")
}

func writeSelectStatement(sql *strings.Builder, n *node) {
	switch n.value.tip {
	case writeATT:
		// next node is a attribute
		if n.next != nil && n.next.value.tip == writeATT {
			sql.WriteString(n.value.keyword)
			sql.WriteRune(',')
			return
		}
		sql.WriteString(n.value.keyword)
		sql.WriteRune(' ')
	case writeDML:
		sql.WriteString(n.value.keyword)
		sql.WriteRune(' ')
	case writeTABLE:
		sql.WriteString(n.value.keyword)
	case writeJOIN:
		sql.WriteRune('\n')
		sql.WriteString(n.value.keyword)
	case writeMIDDLE:
		sql.WriteRune('\n')
		sql.WriteString(n.value.keyword)
		sql.WriteRune(' ')
	default:
		sql.WriteString(n.value.keyword)
	}
}

func writeInsert(sql *strings.Builder, q *statementQueue) {
	if q.head != nil {
		writeInsertStatement(sql, q.head)
		q.head = q.head.next
		q.size--
		writeInsert(sql, q)
		return
	}

	sql.WriteString(";")
}

func writeInsertStatement(sql *strings.Builder, n *node) {
	switch n.value.tip {
	case writeATT:
		// next node is a attribute
		if n.next != nil && n.next.value.tip == writeATT {
			sql.WriteString(n.value.keyword)
			sql.WriteRune(',')
			return
		}
		if n.next == nil {
			sql.WriteString(n.value.keyword + ")")
		} else {
			sql.WriteString(n.value.keyword + ") ")
		}
	case writeDML:
		sql.WriteString(n.value.keyword)
		sql.WriteRune(' ')
		if n.next != nil && n.next.value.tip == writeATT {
			sql.WriteRune('(')
		}
	case writeTABLE:
		sql.WriteString(n.value.keyword)
	default:
		sql.WriteString(n.value.keyword)
	}
}

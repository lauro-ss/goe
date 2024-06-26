package goe

type statementQueue struct {
	head *node
	size int
}

type node struct {
	value *statement
	next  *node
}

func createStatementQueue() *statementQueue {
	return &statementQueue{}
}

func (q *statementQueue) add(v *statement) {
	n := &node{value: v}

	if q.head == nil {
		q.head = n
		q.size++
		return
	}

	tail := getTail(q.head, n.value)
	if tail == nil {
		return
	}
	tail.next = n
	q.size++
}

func (q *statementQueue) get() *statement {
	if q.head == nil {
		return nil
	}

	n := q.head
	q.head = q.head.next
	q.size--
	return n.value
}

func getTail(n *node, v *statement) *node {
	if n.value.keyword == v.keyword && !v.allowCopies {
		return nil
	}
	if n.next != nil {
		return getTail(n.next, v)
	}
	return n
}

type pkQueue struct {
	head     *pkNode
	currrent *pkNode
	size     int
}

type pkNode struct {
	value *pk
	next  *pkNode
}

func createPkQueue() *pkQueue {
	return &pkQueue{}
}

func (q *pkQueue) add(v *pk) {
	n := &pkNode{value: v}

	if q.head == nil {
		q.head = n
		q.currrent = q.head
		q.size++
		return
	}

	tail := getPkNodeTail(q.head, n.value)
	if tail == nil {
		return
	}
	tail.next = n
	q.size++
}

func (q *pkQueue) get() *pk {
	if q.currrent == nil {
		// resets the queue
		q.currrent = q.head
		return nil
	}

	n := q.currrent
	q.currrent = q.currrent.next
	q.size--

	return n.value
}

func getPkNodeTail(n *pkNode, v *pk) *pkNode {
	if n.value.selectName == v.selectName {
		return nil
	}
	if n.next != nil {
		return getPkNodeTail(n.next, v)
	}
	return n
}

func (q *pkQueue) findPk(tableName string) *pk {
	return findPk(q.head, tableName)
}

func findPk(n *pkNode, tableName string) *pk {
	if n.value.table == tableName {
		return n.value
	}
	if n.next != nil {
		return findPk(n.next, tableName)
	}
	return nil
}

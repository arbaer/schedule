package algo

import (
	"log"
	"container/heap"
)

type LookupQueue struct {
	pq *PriorityQueue
	lookup map[string]*Item
}

func (lq *LookupQueue) Init() {
	lq.pq = &PriorityQueue{}
	lq.pq.Init()
	lq.lookup = make(map[string]*Item)
}

func (lq *LookupQueue) pushQueue(x pNode, cost int) {
	item := &Item{value: x, priority: cost}
	heap.Push(lq.pq, item)
	lq.lookup[x.hash()] = item
}

func (lq *LookupQueue) updateQueue(x pNode, cost int) {
	old_item, ok := lq.lookup[x.hash()]
	if !ok {
		log.Fatal("this is not ok")
	}
	lq.pq.update(old_item, x, cost)

//	item := &Item{value: x, priority: cost}
//	heap.Push(lq.pq, item)
//	lq.lookup[x.hash()] = true
}

func (lq *LookupQueue) popQueue() (pNode, int) {
	x := heap.Pop(lq.pq).(*Item)
	delete(lq.lookup, x.value.hash())
	return x.value, x.priority
}

func (lq LookupQueue) Len() int { return len(*lq.pq) }

func (lq *LookupQueue) existsQueue(x pNode) bool {
//	_,ok := lq.lookup[&x]
	_,ok := lq.lookup[x.hash()]
	return ok
}

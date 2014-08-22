// This example demonstrates a priority queue built using the heap interface.
package algo

import (
	"container/heap"
)

// An Item is something we manage in a priority queue.
type Item struct {
	value    pNode // The value of the item; arbitrary.
	priority int    // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Init() {
	heap.Init(pq)
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value pNode, priority int) {
	heap.Remove(pq, item.index)
	item.value = value
	item.priority = priority
	heap.Push(pq, item)
}

// This example inserts some items into a PriorityQueue, manipulates an item,
// and then removes the items in priority order.
/*func main() {
	// Some items and their priorities.
	items := map[string]int{
		"banana": 3, "apple": 2, "pear": 4,
	}

	// Create a priority queue and put the items in it.
	pq := &PriorityQueue{}
	heap.Init(pq)
	for value, priority := range items {
		item := &Item{
			value:    value,
			priority: priority,
		}
		heap.Push(pq, item)
	}

	// Insert a new item and then modify its priority.
	item := &Item{
		value:    "orange",
		priority: 1,
	}
	heap.Push(pq, item)
	pq.update(item, item.value, 5)

	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := heap.Pop(pq).(*Item)
		fmt.Printf("%.2d:%s \n", item.priority, item.value)
	}
}
*/

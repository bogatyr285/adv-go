package hw12

import (
	"container/heap"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Task struct {
	Identifier int
	Priority   int
	index      int
}

type PriorityQueue []*Task

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	task, ok := x.(*Task)
	if !ok {
		panic("got not a task")
	}
	task.index = len(*pq)
	*pq = append(*pq, task)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	task := old[n-1]
	*pq = old[0 : n-1]
	return task
}

// ---
type Scheduler struct {
	pq      PriorityQueue
	taskMap map[int]*Task
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		taskMap: make(map[int]*Task),
	}
}

func (s *Scheduler) AddTask(task Task) {
	t := &Task{
		Identifier: task.Identifier,
		Priority:   task.Priority,
	}
	s.taskMap[task.Identifier] = t
	heap.Push(&s.pq, t)
}

func (s *Scheduler) ChangeTaskPriority(taskID int, newPriority int) {
	if task, exists := s.taskMap[taskID]; exists {
		task.Priority = newPriority
		heap.Fix(&s.pq, task.index)
	}

}

func (s *Scheduler) GetTask() Task {
	if s.pq.Len() == 0 {
		return Task{}
	}
	task := heap.Pop(&s.pq).(*Task)
	delete(s.taskMap, task.Identifier)
	return Task{
		Identifier: task.Identifier,
		Priority:   task.Priority,
	}
}

func TestTrace(t *testing.T) {
	task1 := Task{Identifier: 1, Priority: 10}
	task2 := Task{Identifier: 2, Priority: 20}
	task3 := Task{Identifier: 3, Priority: 30}
	task4 := Task{Identifier: 4, Priority: 40}
	task5 := Task{Identifier: 5, Priority: 50}

	scheduler := NewScheduler()
	scheduler.AddTask(task1)
	scheduler.AddTask(task2)
	scheduler.AddTask(task3)
	scheduler.AddTask(task4)
	scheduler.AddTask(task5)

	task := scheduler.GetTask()
	assert.Equal(t, task5, task)

	task = scheduler.GetTask()
	assert.Equal(t, task4, task)

	scheduler.ChangeTaskPriority(1, 100)
	task1.Priority = 100 // поменяли приоритет в основной задаче для сравнения
	task = scheduler.GetTask()

	assert.Equal(t, task1, task)

	task = scheduler.GetTask()
	assert.Equal(t, task3, task)
}

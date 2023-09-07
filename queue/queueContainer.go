package queue

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type QueueContainer struct {
	Queues   map[string]*Queue
	QMux     sync.Mutex
	QLocks   map[string]*sync.RWMutex
	QSignals map[string]chan struct{}
}

func NewQueueContainer() *QueueContainer {
	return &QueueContainer{
		Queues:   make(map[string]*Queue),
		QMux:     sync.Mutex{},
		QLocks:   make(map[string]*sync.RWMutex),
		QSignals: make(map[string]chan struct{}),
	}
}

func (c *QueueContainer) CleanupOldElements() {
	c.QMux.Lock()
	for queueName, singleQueue := range c.Queues {
		c.QMux.Unlock()
		singleQueue.cleanupOldElements(
			c.QLocks[queueName],
		)
		c.QMux.Lock()
	}
	c.QMux.Unlock()
}

func (c *QueueContainer) EnqueueElement(queueName string, element Element) {
	c.EnsureQueueExists(queueName)

	c.QMux.Lock()
	q, _ := c.Queues[queueName]
	c.QMux.Unlock()

	q.enqueueElement(element, c.QSignals[queueName], c.QLocks[queueName])

	fmt.Println("--> Element of type " + strconv.Itoa(element.Type) + " enqueued to " + queueName)

}

func (c *QueueContainer) DequeueElement(
	queueName string,
	timeoutDuration time.Duration,
	elementType int,
	maxResponseElements int,
	lockRead bool,
) ([]Element, error) {
	c.EnsureQueueExists(queueName)

	c.QMux.Lock()
	q, _ := c.Queues[queueName]
	c.QMux.Unlock()

	return q.dequeueElement(
		timeoutDuration,
		elementType,
		maxResponseElements,
		c.QLocks[queueName],
		c.QSignals[queueName],
		queueName,
		lockRead,
	)
}

func (c *QueueContainer) EnsureQueueExists(queueName string) *Queue {
	c.QMux.Lock()
	defer c.QMux.Unlock()
	_, exists := c.Queues[queueName]
	if !exists {
		c.Queues[queueName] = &Queue{Id: queueName, LockRead: false}
		c.QLocks[queueName] = &sync.RWMutex{}
		c.QSignals[queueName] = make(chan struct{})

		fmt.Println("--> Queue created " + queueName)
	}

	return c.Queues[queueName]
}

func (c *QueueContainer) ClearQueue(queueName string) {
	c.QMux.Lock()
	defer c.QMux.Unlock()

	if queueName == "" {
		c.Queues = make(map[string]*Queue)
		return
	}

	_, exists := c.Queues[queueName]
	if exists {
		c.Queues[queueName] = &Queue{Id: queueName, LockRead: false}
		c.QLocks[queueName] = &sync.RWMutex{}
		c.QSignals[queueName] = make(chan struct{})

		fmt.Println("--> Queue created " + queueName)
	}
}

func (c *QueueContainer) GetQueueByName(queueName string) *Queue {
	c.QMux.Lock()
	defer c.QMux.Unlock()

	return c.Queues[queueName]
}

func (c *QueueContainer) GetAllQueues() map[string]*Queue {
	c.QMux.Lock()
	defer c.QMux.Unlock()

	return c.Queues
}

func (c *QueueContainer) UnlockRead(queueName string) {
	c.EnsureQueueExists(queueName)

	c.QMux.Lock()
	q, _ := c.Queues[queueName]
	c.QMux.Unlock()

	q.UnlockRead(c.QLocks[queueName])
}

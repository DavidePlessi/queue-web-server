package queue

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Queue struct {
	Id       string
	Elements []Element
}

func (q *Queue) cleanupOldElements(
	locker *sync.Mutex,
) {
	locker.Lock()
	updatedElements := make([]Element, 0)
	for _, e := range q.Elements {
		elementExpirationTime, err := time.Parse(time.RFC3339, e.ExpirationTime)
		if err != nil {
			updatedElements = append(updatedElements, e)
			continue
		}

		if elementExpirationTime.Sub(time.Now()) > 0 {
			updatedElements = append(updatedElements, e)
		}
	}
	if expiredElementsCount := len(q.Elements) - len(updatedElements); expiredElementsCount > 0 {
		fmt.Println("--> Expired elements " + strconv.Itoa(expiredElementsCount))
	}

	q.Elements = updatedElements
	locker.Unlock()
}

func (q *Queue) enqueueElement(
	element Element,
	signal chan struct{},
	locker *sync.Mutex,
) {
	locker.Lock()
	q.Elements = append(q.Elements, element)
	locker.Unlock()

	select {
	case signal <- struct{}{}:
	default:
	}
}

func (q *Queue) dequeueElement(
	timeoutDuration time.Duration,
	elementType int,
	maxResponseElements int,
	locker *sync.Mutex,
	signal chan struct{},
	queueName string,
) []Element {
	var elements []Element
	found := false
	timer := time.NewTimer(timeoutDuration)
	for !found {
		locker.Lock()
		var elementsAddedCount int = 0
		for i := 0; i < len(q.Elements); i++ {
			e := q.Elements[i]
			if e.Type == elementType || elementType == -1 {
				elements = append(elements, e)
				q.Elements = append(q.Elements[:i], q.Elements[i+1:]...)
				i--
				elementsAddedCount++
				found = true
				fmt.Println("--> Element of type " + strconv.Itoa(e.Type) + " dequeued from " + queueName)
			}
			if elementsAddedCount == maxResponseElements {
				break
			}
		}
		locker.Unlock()

		if !found {
			fmt.Println("--> Waiting for an element to be added to " + queueName)
			select {
			case <-signal:
				found = false
			case <-timer.C:
				fmt.Println("--> Waiting for an element to be added to " + queueName + " TIMEOUT")
				found = true
			}
		}
	}

	return elements
}

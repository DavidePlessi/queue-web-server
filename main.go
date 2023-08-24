// main.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"queue-web-server/queue"
)

var (
	queues       = make(map[string]*queue.Queue)
	queuesMux    = sync.Mutex{}
	queueLocks   = make(map[string]*sync.Mutex)
	queueSignals = make(map[string]chan struct{})
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/create", createQueue).Methods("POST")
	router.HandleFunc("/{queueName}/enqueue", enqueueElement).Methods("POST")
	router.HandleFunc(
		"/{queueName}/dequeue",
		dequeueElement).Methods("GET")
	router.HandleFunc("/queues", getAllQueues).Methods("GET")

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}

func innerCreateQueue(queueName string) {
	queuesMux.Lock()
	defer queuesMux.Unlock()

	if _, exists := queues[queueName]; !exists {
		queues[queueName] = &queue.Queue{Id: queueName}
		queueLocks[queueName] = &sync.Mutex{}
		queueSignals[queueName] = make(chan struct{})
	}
}

func createQueue(w http.ResponseWriter, r *http.Request) {
	var id string
	err := json.NewDecoder(r.Body).Decode(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	innerCreateQueue(id)

	w.WriteHeader(http.StatusCreated)
}

func enqueueElement(w http.ResponseWriter, r *http.Request) {
	queueName := mux.Vars(r)["queueName"]
	innerCreateQueue(queueName)

	var element queue.Element
	err := json.NewDecoder(r.Body).Decode(&element)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queuesMux.Lock()
	q, exists := queues[queueName]
	queuesMux.Unlock()

	if !exists {
		http.Error(w, "Queue not found", http.StatusNotFound)
		return
	}

	queueLocks[queueName].Lock()
	q.Elements = append(q.Elements, element)
	queueLocks[queueName].Unlock()

	select {
	case queueSignals[queueName] <- struct{}{}:
	default:
	}

	w.WriteHeader(http.StatusCreated)
}

func dequeueElement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queueName := vars["queueName"]

	queryParams := r.URL.Query()
	innerCreateQueue(queueName)
	elementType := queryParams.Get("elementType")
	maxResponseElements, err := strconv.Atoi(queryParams.Get("maxResponseElements"))
	if err != nil {
		maxResponseElements = 1
	}

	timeout := queryParams.Get("timeout")
	timeoutDuration, err := time.ParseDuration(timeout + "s")
	if err != nil {
		timeoutDuration, _ = time.ParseDuration("30s")
	}

	queuesMux.Lock()
	q, exists := queues[queueName]
	queuesMux.Unlock()

	if !exists {
		http.Error(w, "Queue not found", http.StatusNotFound)
		return
	}

	var elements []queue.Element

	found := false
	timer := time.NewTimer(timeoutDuration)
	for !found {

		queueLocks[queueName].Lock()
		var elementsAddedCount int = 0
		for i := 0; i < len(q.Elements); i++ {
			e := q.Elements[i]
			if e.Type == elementType || elementType != "" {
				elements = append(elements, e)
				q.Elements = append(q.Elements[:i], q.Elements[i+1:]...)
				i--
				elementsAddedCount++
				found = true
			}
			if elementsAddedCount == maxResponseElements {
				break
			}
		}
		queueLocks[queueName].Unlock()

		if !found {
			select {
			case <-queueSignals[queueName]:
				found = false
			case <-timer.C:
				found = true
			}
		}
	}

	acceptHeader := r.Header.Get("Accept")
	if acceptHeader == "application/csv" {
		csvSeparator := r.Header.Get("Csv-Separator")
		if csvSeparator == "" {
			csvSeparator = ";"
		}

		csvLineSeparator := r.Header.Get("Csv-LineSeparator")
		if csvLineSeparator == "" {
			csvLineSeparator = "\n"
		}

		w.Header().Set("Content-Type", "application/csv")

		//var firstElement queue.Element
		//if len(elements) > 0 {
		//	//Header
		//	firstElement = elements[0]
		//	w.Write([]byte("type" + csvSeparator))
		//	for _, key := range firstElement.Body.Keys() {
		//		w.Write([]byte(key + csvSeparator))
		//	}
		//	w.Write([]byte(csvLineSeparator))
		//}

		for _, value := range elements {
			w.Write([]byte(value.Type + csvSeparator))
			for _, key := range value.Body.Keys() {
				if value, exists := value.Body.Get(key); exists {
					w.Write([]byte(fmt.Sprintf("%v", value) + csvSeparator))
				} else {
					w.Write([]byte(csvSeparator))
				}

			}
			w.Write([]byte(csvLineSeparator))
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(elements)
	}
}

func getAllQueues(w http.ResponseWriter, r *http.Request) {
	queuesMux.Lock()
	defer queuesMux.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(queues)
}

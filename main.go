// main.go
package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"queue-web-server/queue"
)

var (
	queues     = make(map[string]*queue.Queue)
	queuesMux  = sync.Mutex{}
	queueLocks = make(map[string]*sync.Mutex)
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/create", createQueue).Methods("POST")
	router.HandleFunc("/{queueName}/enqueue", enqueueElement).Methods("POST")
	router.HandleFunc("/{queueName}/dequeue/{elementType}", dequeueElement).Methods("GET")
	router.HandleFunc("/queues", getAllQueues).Methods("GET")

	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}

func createQueue(w http.ResponseWriter, r *http.Request) {
	var id string
	err := json.NewDecoder(r.Body).Decode(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queuesMux.Lock()
	defer queuesMux.Unlock()

	if _, exists := queues[id]; exists {
		http.Error(w, "Queue already exists", http.StatusBadRequest)
		return
	}

	queues[id] = &queue.Queue{Id: id}
	queueLocks[id] = &sync.Mutex{}
	w.WriteHeader(http.StatusCreated)
}

func enqueueElement(w http.ResponseWriter, r *http.Request) {
	queueName := mux.Vars(r)["queueId"]
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
	defer queueLocks[queueName].Unlock()

	q.Elements = append(q.Elements, element)
	w.WriteHeader(http.StatusCreated)
}

func dequeueElement(w http.ResponseWriter, r *http.Request) {
	queueName := mux.Vars(r)["queueName"]

	queuesMux.Lock()
	defer queuesMux.Unlock()

	queuesMux.Lock()
	q, exists := queues[queueName]
	queuesMux.Unlock()

	if !exists {
		http.Error(w, "Queue not found", http.StatusNotFound)
		return
	}

	queueLocks[queueName].Lock()
	defer queueLocks[queueName].Unlock()
	var element queue.Element

	if elementType := mux.Vars(r)["elementType"]; elementType != "" {
		// Find and remove the element of the specified type
		for i, e := range q.Elements {
			if e.Type == elementType {
				element = e
				q.Elements = append(q.Elements[:i], q.Elements[i+1:]...)
				break
			}
		}
	} else {
		// Dequeue the first element
		if len(q.Elements) > 0 {
			element = q.Elements[0]
			q.Elements = q.Elements[1:]
		} else {
			http.Error(w, "Queue is empty", http.StatusNotFound)
			return
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

		// Respond with CSV
		w.Header().Set("Content-Type", "application/csv")
		for _, value := range element.Body.([]interface{}) {
			w.Write([]byte(value.(string) + csvSeparator))
		}
		//w.Write([]byte(csvLineSeparator))
	} else {
		// Respond with JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(element)
	}
}

func getAllQueues(w http.ResponseWriter, r *http.Request) {
	queuesMux.Lock()
	defer queuesMux.Unlock()

	var response []map[string]interface{}
	for _, q := range queues {
		queueCopy := make([]queue.Element, len(q.Elements))
		copy(queueCopy, q.Elements) // Create a copy of the queue elements
		response = append(response, map[string]interface{}{
			"queueId":  q.Id,
			"elements": queueCopy,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

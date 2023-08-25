// main.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"queue-web-server/queue"
)

var queueContainer = queue.NewQueueContainer()

func main() {
	port := ":8080"
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}

	router := mux.NewRouter()
	router.HandleFunc("/create", createQueue).Methods("POST")
	router.HandleFunc("/{queueName}/enqueue", enqueueElement).Methods("POST")
	router.HandleFunc(
		"/{queueName}/dequeue",
		dequeueElement).Methods("GET")
	router.HandleFunc("/queues", getQueues).Methods("GET")
	router.HandleFunc("/clear", clearQueue).Methods("GET")

	http.Handle("/", router)
	fmt.Println("--> All ready on port" + port)

	//Check for old element every n minutes
	cleanupTimer := time.NewTicker(1 * time.Minute)

	go func() {
		fmt.Println("--> Start clear routine")
		for range cleanupTimer.C {
			queueContainer.CleanupOldElements()
		}
	}()

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func createQueue(w http.ResponseWriter, r *http.Request) {
	var id string
	err := json.NewDecoder(r.Body).Decode(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queueContainer.EnsureQueueExists(id)

	w.WriteHeader(http.StatusCreated)
}

func enqueueElement(w http.ResponseWriter, r *http.Request) {
	queueName := mux.Vars(r)["queueName"]

	var element queue.Element
	err := json.NewDecoder(r.Body).Decode(&element)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queueContainer.EnqueueElement(queueName, element)

	jsonOutput, err := json.MarshalIndent(element, "", "    ")
	if err != nil {
		fmt.Println("Error during json encoding for console:", err)
	}
	fmt.Println("----------ENQUEUED-JSON---------")
	fmt.Println(string(jsonOutput))
	fmt.Println("--------------------------------")

	w.WriteHeader(http.StatusCreated)
}

func dequeueElement(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queueName := vars["queueName"]

	queryParams := r.URL.Query()

	elementTypeStr := queryParams.Get("elementType")
	var elementType = -1
	if elementTypeStr != "" {
		elementType, _ = strconv.Atoi(elementTypeStr)
	}

	maxResponseElements, err := strconv.Atoi(queryParams.Get("maxResponseElements"))
	if err != nil {
		maxResponseElements = 1
	}

	timeout := queryParams.Get("timeout")
	timeoutDuration, err := time.ParseDuration(timeout + "s")
	if err != nil {
		timeoutDuration, _ = time.ParseDuration("30s")
	}

	var elements = queueContainer.DequeueElement(
		queueName,
		timeoutDuration,
		elementType,
		maxResponseElements,
	)

	jsonOutput, err := json.MarshalIndent(elements, "", "    ")
	if err != nil {
		fmt.Println("Error during json encoding for console:", err)
	}
	fmt.Println("----------DEQUEUED-JSON---------")
	fmt.Println(string(jsonOutput))
	fmt.Println("--------------------------------")

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
			w.Write([]byte(strconv.Itoa(value.Type) + csvSeparator))
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

func getQueues(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	queueName := queryParams.Get("queueId")

	w.Header().Set("Content-Type", "application/json")

	if queueName == "" {
		var queues = queueContainer.GetAllQueues()
		json.NewEncoder(w).Encode(queues)
	} else {
		var queues = make(map[string]*queue.Queue)
		queues[queueName] = queueContainer.GetQueueByName(queueName)
		json.NewEncoder(w).Encode(queues)
	}

}

func clearQueue(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	queueName := queryParams.Get("queueId")
	queueContainer.ClearQueue(queueName)
}

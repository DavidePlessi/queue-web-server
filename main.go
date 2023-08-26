// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"queue-web-server/queue"
)

var queueContainer = queue.NewQueueContainer()

func main() {
	port, expiredElementCheckInterval := parseArguments()
	startExpiredElementCheckRoutine(expiredElementCheckInterval)
	startWebServer(port)
}

// --- HELP FUNCTIONS
func parseArguments() (string, int) {

	var port string
	var expiredElementCheckInterval int

	flag.StringVar(&port, "p", "8080", "Port to listen on")
	flag.IntVar(&expiredElementCheckInterval, "i", 60, "Interval in seconds to check for expired elements")

	flag.Parse()

	if port != "" {
		port = ":" + port
	} else {
		port = ":8080"
	}

	return port, expiredElementCheckInterval
}

func startExpiredElementCheckRoutine(interval int) {
	if interval > 0 {
		cleanupTimer := time.NewTicker(time.Duration(interval) * time.Second)

		go func() {
			fmt.Println("--> Start clear routine")
			for range cleanupTimer.C {
				queueContainer.CleanupOldElements()
			}
		}()
	}
}

func startWebServer(port string) {
	router := mux.NewRouter()

	router.HandleFunc("/create", createQueue).Methods("POST")
	router.HandleFunc("/{queueName}/enqueue", enqueueElement).Methods("POST")
	router.HandleFunc("/{queueName}/dequeue", dequeueElement).Methods("GET")
	router.HandleFunc("/queues", getQueues).Methods("GET")
	router.HandleFunc("/clear", clearQueue).Methods("GET")

	http.Handle("/", router)

	fmt.Println("--> All ready on port" + port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err)
	}
}

//--- HELP FUNCTIONS END

// --- WEB HANDLERS
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

//--- WEB HANDLERS END

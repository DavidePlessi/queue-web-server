# Queue Web Server with Go

This project demonstrates a simple queue web server implemented in Go. It provides API endpoints for managing multiple queues, enqueueing elements, dequeuing elements, and retrieving queue information.

## Prerequisites

- Go (Golang) installed on your system
- Basic understanding of HTTP, RESTful APIs, and Go programming

## Project Structure

The project structure is organized as follows:

my-queue-server/  
├── main.go  
├── queue/  
│ ├── queue.go  
│ └── element.go  
└── README.md

- `main.go`: The main entry point of the application containing the HTTP server setup and API endpoints.
- `queue/queue.go`: Defines the `Queue` struct and related functions.
- `queue/element.go`: Defines the `Element` struct for queue elements.

## Getting Started

1. Clone this repository:

   ```bash
   git clone https://github.com/DavidePlessi/queue-web-server.git
   cd my-queue-server
   ```
2. Install required packages:

   ```bash
   go get -u github.com/gorilla/mux
   go get -u github.com/iancoleman/orderedmap
   ```
3. Run the server:

   ```bash
   go run main.go
   ```
4. Interact with the API endpoints using tools like curl or Postman.

## API Endpoints
### Create a Queue
- Endpoint: /create
- Method: POST
- Request Body: id
- Response: HTTP status code indicating success or failure
### Enqueue an Element
- Endpoint: /{queueId}/enqueue
- Method: POST
- Request Body: Element type and body
- Response: HTTP status code indicating success or failure
### Dequeue an Element
- Endpoint: /{queueName}/dequeue
- Query Params: timeout, maxResponseElements, elementType
  - timeout: call timeout, default value 30s
  - maxResponseElements: max number of elements to dequeue, default value 5
  - elementType: 
- Method: GET
- Response: JSON or CSV response based on Accept header
### Get All Queues
- Endpoint: /queues
- Method: GET
- Response: JSON response containing all queues and their elements

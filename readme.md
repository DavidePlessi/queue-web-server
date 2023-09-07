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
│ ├── queueContainer.go  
│ ├── queue.go  
│ └── element.go  
└── README.md

- `main.go`: The main entry point of the application containing the HTTP server setup and API endpoints.
- `queue/queueContainer.go`: Defines the `QueueContainer` struct and related functions.
- `queue/queue.go`: Defines the `Queue` struct and related functions.
- `queue/element.go`: Defines the `Element` struct for queue elements.

## Getting Started

1. Clone this repository:

   ```bash
   git clone https://github.com/DavidePlessi/queue-web-server.git
   cd queue-web-server
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
   Can accept several arguments:
    - p: the port on which the server will listen, default value 8080
    - i: Interval in seconds to check for expired elements, default value 60s
   ```bash
   go run main.go -p 8100 -i 60
   ```
4. Interact with the API endpoints using tools like curl or Postman.

## Build
The script `build.go` is designed to help you build QWS for different target operating systems 
and architectures using cross-compilation.
Usage:
```bash
go run build.go -os <os> -arch <arch>
```
available options for os and arch can be found inside the script.  
Builds are stored inside `dist` folder `"./dist/qws_<version>_<targetOS>_<targetArch>(|.exe)`.  
Is it possible to compile all common os/arch combination using the flag -all
```bash
go run build.go -all
```

## API Endpoints
### Create a Queue
- Endpoint: /create
- Method: POST
- Request Body: id
- Response: HTTP status code indicating success or failure
### Enqueue an Element
- Endpoint: /{queueId}/enqueue
- Method: POST
- Request Body: Element type, body, time, expirationTime(optional)
```json lines
{
    "type": 1,
    "body": {
        "1": "6",
        "3": false,
        "2": 7
    },
    "time": "2023-08-25T16:06:21.683798+02:00",
    "expirationTime": "2023-08-25T16:30:21.683798+02:00" //optional
}
```
- Response: HTTP status code indicating success or failure
### Dequeue an Element
- Endpoint: /{queueName}/dequeue
- Query Params: timeout, maxResponseElements, elementType
  - timeout: call timeout, default value 30s
  - maxResponseElements: max number of elements to dequeue, default value 5
  - elementType: the type of element
  - lockQueueRead: lock the queue on read, default value false (needs to be unlocked with /{queueName}/unlock-read)
- Method: GET
- Response: JSON or CSV response based on Accept header
### Get Queues
- Endpoint: /queues
- Query Params: queueId (get specific queue by Id)
- Method: GET
- Response: JSON response containing all queues and their elements
### Clear Queues
- Endpoint: /clear
- Query Params: queueId (get specific queue by Id)
- Method: POST
- Response: HTTP status code indicating success or failure
### Unlock Queue
- Endpoint: /{queueName}/unlock-read
- Method: POST
- Response: HTTP status code indicating success or failure

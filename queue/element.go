package queue

type Element struct {
	Type string      `json:"type"`
	Body interface{} `json:"body"`
}

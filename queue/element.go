package queue

import "github.com/iancoleman/orderedmap"

type Element struct {
	Type int                   `json:"type"`
	Body orderedmap.OrderedMap `json:"body"`
	Time string                `json:"time"`
}

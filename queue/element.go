package queue

import "github.com/iancoleman/orderedmap"

type Element struct {
	Type string                `json:"type"`
	Body orderedmap.OrderedMap `json:"body"`
}

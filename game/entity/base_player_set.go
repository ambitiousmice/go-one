package entity

import (
	"bytes"
)

// BasePlayerSet is the data structure for a set of entities
type BasePlayerSet map[*BasePlayer]struct{}

// Add adds an entity to the BasePlayerSet
func (es BasePlayerSet) Add(entity *BasePlayer) {
	es[entity] = struct{}{}
}

// Del deletes an entity from the BasePlayerSet
func (es BasePlayerSet) Del(entity *BasePlayer) {
	delete(es, entity)
}

// Contains returns if the entity is in the BasePlayerSet
func (es BasePlayerSet) Contains(entity *BasePlayer) bool {
	_, ok := es[entity]
	return ok
}

func (es BasePlayerSet) ForEach(f func(e *BasePlayer)) {
	for e := range es {
		f(e)
	}
}

func (es BasePlayerSet) String() string {
	b := bytes.Buffer{}
	b.WriteString("{")
	first := true
	for entity := range es {
		if !first {
			b.WriteString(", ")
		} else {
			first = false
		}
		b.WriteString(entity.String())
	}
	b.WriteString("}")
	return b.String()
}

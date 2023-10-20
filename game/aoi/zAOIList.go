package aoi

import "go-one/game/common"

type zAOIList struct {
	aoidist common.Coord
	head    *xzaoi
	tail    *xzaoi
}

func newYAOIList(aoidist common.Coord) *zAOIList {
	return &zAOIList{aoidist: aoidist}
}

func (sl *zAOIList) Insert(aoi *xzaoi) {
	insertCoord := aoi.aoi.z
	if sl.head != nil {
		p := sl.head
		for p != nil && p.aoi.z < insertCoord {
			p = p.zNext
		}
		// now, p == nil or p.coord >= insertCoord
		if p == nil { // if p == nil, insert xzaoi at the end of list
			tail := sl.tail
			tail.zNext = aoi
			aoi.zPrev = tail
			sl.tail = aoi
		} else { // otherwise, p >= xzaoi, insert xzaoi before p
			prev := p.zPrev
			aoi.zNext = p
			p.zPrev = aoi
			aoi.zPrev = prev

			if prev != nil {
				prev.zNext = aoi
			} else { // p is the head, so xzaoi should be the new head
				sl.head = aoi
			}
		}
	} else {
		sl.head = aoi
		sl.tail = aoi
	}
}

func (sl *zAOIList) Remove(aoi *xzaoi) {
	prev := aoi.zPrev
	next := aoi.zNext
	if prev != nil {
		prev.zNext = next
		aoi.zPrev = nil
	} else {
		sl.head = next
	}
	if next != nil {
		next.zPrev = prev
		aoi.zNext = nil
	} else {
		sl.tail = prev
	}
}

func (sl *zAOIList) Move(aoi *xzaoi, oldCoord common.Coord) {
	coord := aoi.aoi.z
	if coord > oldCoord {
		// moving to next ...
		next := aoi.zNext
		if next == nil || next.aoi.z >= coord {
			// no need to adjust in list
			return
		}
		prev := aoi.zPrev
		//fmt.Println(1, prev, next, prev == nil || prev.zNext == xzaoi)
		if prev != nil {
			prev.zNext = next // remove xzaoi from list
		} else {
			sl.head = next // xzaoi is the head, trim it
		}
		next.zPrev = prev

		//fmt.Println(2, prev, next, prev == nil || prev.zNext == next)
		prev, next = next, next.zNext
		for next != nil && next.aoi.z < coord {
			prev, next = next, next.zNext
			//fmt.Println(2, prev, next, prev == nil || prev.zNext == next)
		}
		//fmt.Println(3, prev, next)
		// no we have prev.X < coord && (next == nil || next.X >= coord), so insert between prev and next
		prev.zNext = aoi
		aoi.zPrev = prev
		if next != nil {
			next.zPrev = aoi
		} else {
			sl.tail = aoi
		}
		aoi.zNext = next

		//fmt.Println(4)
	} else {
		// moving to prev ...
		prev := aoi.zPrev
		if prev == nil || prev.aoi.z <= coord {
			// no need to adjust in list
			return
		}

		next := aoi.zNext
		if next != nil {
			next.zPrev = prev
		} else {
			sl.tail = prev // xzaoi is the head, trim it
		}
		prev.zNext = next // remove xzaoi from list

		next, prev = prev, prev.zPrev
		for prev != nil && prev.aoi.z > coord {
			next, prev = prev, prev.zPrev
		}
		// no we have next.X > coord && (prev == nil || prev.X <= coord), so insert between prev and next
		next.zPrev = aoi
		aoi.zNext = next
		if prev != nil {
			prev.zNext = aoi
		} else {
			sl.head = aoi
		}
		aoi.zPrev = prev
	}
}

func (sl *zAOIList) Mark(aoi *xzaoi) {
	prev := aoi.zPrev
	coord := aoi.aoi.z

	minCoord := coord - sl.aoidist
	for prev != nil && prev.aoi.z >= minCoord {
		prev.markVal += 1
		prev = prev.zPrev
	}

	next := aoi.zNext
	maxCoord := coord + sl.aoidist
	for next != nil && next.aoi.z <= maxCoord {
		next.markVal += 1
		next = next.zNext
	}
}

func (sl *zAOIList) ClearMark(aoi *xzaoi) {
	prev := aoi.zPrev
	coord := aoi.aoi.z

	minCoord := coord - sl.aoidist
	for prev != nil && prev.aoi.z >= minCoord {
		prev.markVal = 0
		prev = prev.zPrev
	}

	next := aoi.zNext
	maxCoord := coord + sl.aoidist
	for next != nil && next.aoi.z <= maxCoord {
		next.markVal = 0
		next = next.zNext
	}
}

package aoi

import (
	"github.com/ambitiousmice/go-one/game/common"
	"sync"
)

type xzaoi struct {
	aoi          *AOI
	neighbors    map[*xzaoi]struct{}
	xPrev, xNext *xzaoi
	zPrev, zNext *xzaoi
	markVal      int
}

// XZListAOIManager is an implementation of AOICalculator using XZ lists
type XZListAOIManager struct {
	aoidist    common.Coord
	xSweepList *xAOIList
	zSweepList *zAOIList
	mutex      sync.Mutex
}

// NewXZListAOIManager creates a new XZListAOIManager
func NewXZListAOIManager(aoidist common.Coord) AOIManager {
	return &XZListAOIManager{
		aoidist:    aoidist,
		xSweepList: newXAOIList(aoidist),
		zSweepList: newYAOIList(aoidist),
	}
}

// Enter is called when Entity enters Space
func (aoiman *XZListAOIManager) Enter(aoi *AOI, x, z common.Coord) {
	aoiman.mutex.Lock()
	defer aoiman.mutex.Unlock()
	aoi.dist = aoiman.aoidist

	xzaoi := &xzaoi{
		aoi:       aoi,
		neighbors: map[*xzaoi]struct{}{},
	}
	aoi.x, aoi.z = x, z
	aoi.implData = xzaoi
	aoiman.xSweepList.Insert(xzaoi)
	aoiman.zSweepList.Insert(xzaoi)
	aoiman.adjust(xzaoi)
}

// Leave is called when Entity leaves Space
func (aoiman *XZListAOIManager) Leave(aoi *AOI) {
	aoiman.mutex.Lock()
	defer aoiman.mutex.Unlock()
	xzaoi := aoi.implData.(*xzaoi)
	aoiman.xSweepList.Remove(xzaoi)
	aoiman.zSweepList.Remove(xzaoi)
	aoiman.adjust(xzaoi)
}

// Moved is called when Entity moves in Space
func (aoiman *XZListAOIManager) Moved(aoi *AOI, x, z common.Coord) {
	aoiman.mutex.Lock()
	defer aoiman.mutex.Unlock()
	oldX := aoi.x
	oldZ := aoi.z
	aoi.x, aoi.z = x, z
	xzaoi := aoi.implData.(*xzaoi)
	if oldX != x {
		aoiman.xSweepList.Move(xzaoi, oldX)
	}
	if oldZ != z {
		aoiman.zSweepList.Move(xzaoi, oldZ)
	}
	aoiman.adjust(xzaoi)
}

// adjust is called by Entity to adjust neighbors
func (aoiman *XZListAOIManager) adjust(aoi *xzaoi) {
	aoiman.xSweepList.Mark(aoi)
	aoiman.zSweepList.Mark(aoi)
	// AOI marked twice are neighbors
	for neighbor := range aoi.neighbors {
		if neighbor.markVal == 2 {
			// neighbors kept
			neighbor.markVal = -2 // mark this as neighbor
		} else { // markVal < 2
			// was neighbor, but not any more
			delete(aoi.neighbors, neighbor)
			aoi.aoi.callback.OnLeaveAOI(neighbor.aoi)
			delete(neighbor.neighbors, aoi)
			neighbor.aoi.callback.OnLeaveAOI(aoi.aoi)
		}
	}

	// travel in X list again to find all new neighbors, whose markVal == 2
	aoiman.xSweepList.GetClearMarkedNeighbors(aoi)
	// travel in Z list again to unmark all
	aoiman.zSweepList.ClearMark(aoi)
}

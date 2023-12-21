package aoi

import "github.com/ambitiousmice/go-one/game/common"

// Coord is the type for coordinate axes values

type AOI struct {
	x    common.Coord
	z    common.Coord
	dist common.Coord
	Data interface{}

	callback AOICallback
	implData interface{}
}

func InitAOI(aoi *AOI, dist common.Coord, data interface{}, callback AOICallback) {
	aoi.dist = dist
	aoi.Data = data
	aoi.callback = callback
}

type AOICallback interface {
	OnEnterAOI(other *AOI)
	OnLeaveAOI(other *AOI)
}

type AOIManager interface {
	Enter(aoi *AOI, x, z common.Coord)
	Leave(aoi *AOI)
	Moved(aoi *AOI, x, z common.Coord)
}

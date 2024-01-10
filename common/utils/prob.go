package utils

import "math/rand"

// ProbabilityInterval 表示概率区间
type ProbabilityInterval struct {
	Index int
	Start int
	End   int
}

// ProbabilityGenerator 表示概率生成器
type ProbabilityGenerator struct {
	TotalProbability     int
	ProbabilityIntervals []ProbabilityInterval
}

// GenerateResult 生成随机结果
func (pg *ProbabilityGenerator) GenerateResult() int {
	random := rand.Intn(pg.TotalProbability)
	for _, probabilityInterval := range pg.ProbabilityIntervals {
		if random >= probabilityInterval.Start && random < probabilityInterval.End {
			return probabilityInterval.Index
		}
	}
	return -1
}

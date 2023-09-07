package idgenerator

type IDGenerator interface {
	NextIDStr() string
	NextID() int64
}

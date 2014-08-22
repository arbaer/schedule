package types

//node used in simple algorithms
type JobNode struct {
	Id int

	IOTime  float64

	Name string

	Inputs []string
	InIds []int

	OutIds []int

	CpuTime float64
	Depth	int
}


package engine

type Interface interface {
	Execute(id string, ex string) (string, error)
	SetBase(inputBase int, outputBase int)
	Base() (int, int)
}

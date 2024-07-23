package engine

type Interface interface {
	Execute(id string, ex string) (string, error)
}

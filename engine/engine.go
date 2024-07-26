package engine

type Interface interface {
	Execute(id string, ex string) (string, error)
	SetBase(Base)
	Base() Base
	WithErrorSupression(with func())
	WithNumberBase(base Base, with func())
}

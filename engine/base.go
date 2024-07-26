package engine

type Base struct {
	Input  int
	Output int
}

// OutputOnly returns an instance of the Base type where the input base is the
// same as the output base
func (bs Base) OutputOnly() Base {
	return Base{
		Input:  bs.Output,
		Output: bs.Output,
	}
}

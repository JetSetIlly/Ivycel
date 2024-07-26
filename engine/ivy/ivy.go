package ivy

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"robpike.io/ivy/config"
	"robpike.io/ivy/exec"
	"robpike.io/ivy/parse"
	"robpike.io/ivy/run"
	"robpike.io/ivy/scan"
	"robpike.io/ivy/value"
)

const contextName = "ivy"

type Ivy struct {
	conf    config.Config
	context value.Context

	lastResultBuffer []byte
	lastErrorBuffer  []byte
	lastResult       *bytes.Buffer
	lastError        *bytes.Buffer

	inputBase  int
	outputBase int

	errorSuppression bool
	lastErr          error
}

func New() Ivy {
	var iv Ivy

	iv.lastResult = bytes.NewBuffer(iv.lastResultBuffer)
	iv.lastError = bytes.NewBuffer(iv.lastErrorBuffer)
	iv.conf.SetOutput(iv.lastResult)
	iv.conf.SetErrOutput(iv.lastError)

	iv.context = exec.NewContext(&iv.conf)
	iv.SetBase(10, 10)

	return iv
}

// log error normalises the error message and assigns it to the lastErr field.
// it also returns the lastErr field (ie. the normalised form)
func (iv *Ivy) logError(err error) error {
	spl := strings.SplitN(err.Error(), ":", 3)
	if len(spl) > 0 {
		err = errors.New(spl[len(spl)-1])
	}
	if !iv.errorSuppression {
		iv.lastErr = err
	}
	return err
}

func (iv Ivy) LastError() error {
	return iv.lastErr
}

// run the supplied function but with the error suppression flag set
func (iv *Ivy) WithErrorSupression(with func()) {
	iv.errorSuppression = true
	with()
	iv.errorSuppression = false
}

func (iv *Ivy) execute(ex string) (string, error) {
	if !iv.errorSuppression {
		iv.lastErr = nil
	}

	iv.lastResult.Reset()
	iv.lastError.Reset()

	scanner := scan.New(iv.context, contextName, strings.NewReader(ex))
	parser := parse.NewParser(contextName, scanner, iv.context)

	ok := run.Run(parser, iv.context, false)
	if !ok {
		return "", errors.New(iv.lastError.String())
	}

	return iv.lastResult.String(), nil
}

func (iv *Ivy) Execute(ref string, ex string) (string, error) {
	ref, ex = normaliseCellReferences(ref, ex)

	_, err := iv.execute(fmt.Sprintf("%s = %s", ref, ex))
	if err != nil {
		return "", iv.logError(err)
	}

	result, err := iv.execute(ref)
	if err != nil {
		return "", iv.logError(err)
	}

	return result, nil
}

func (iv *Ivy) SetBase(inputBase int, outputBase int) {
	iv.inputBase = inputBase
	iv.outputBase = outputBase

	var err error

	_, err = iv.execute(fmt.Sprintf(")ibase %d", iv.inputBase))
	_, err = iv.execute(fmt.Sprintf(")obase %d", iv.outputBase))

	if err != nil {
		iv.logError(err)
		return
	}
}

func (iv Ivy) Base() (int, int) {
	return iv.inputBase, iv.outputBase
}

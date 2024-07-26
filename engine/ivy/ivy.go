package ivy

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/jetsetilly/ivycel/engine"
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

	base engine.Base

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
	iv.SetBase(engine.Base{Input: 10, Output: 10})

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

// run the supplied function but with the error suppression flag set
func (iv *Ivy) WithNumberBase(base engine.Base, with func()) {
	iv.setBase(base)
	with()
	iv.setBase(iv.base)
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

func (iv *Ivy) setBase(base engine.Base) {
	var err error

	_, err = iv.execute(fmt.Sprintf(")ibase %d", base.Input))
	if err != nil {
		iv.logError(err)
	}

	_, err = iv.execute(fmt.Sprintf(")obase %d", base.Output))
	if err != nil {
		iv.logError(err)
	}
}

func (iv *Ivy) SetBase(base engine.Base) {
	iv.base = base
	iv.setBase(base)
}

func (iv Ivy) Base() engine.Base {
	return iv.base
}

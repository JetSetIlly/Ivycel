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

	lastErr error
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

func (iv *Ivy) logError(err error) {
	spl := strings.SplitN(err.Error(), ":", 3)
	if len(spl) > 0 {
		iv.lastErr = errors.New(spl[len(spl)-1])
	}
}

func (iv Ivy) LastError() error {
	return iv.lastErr
}

func (iv *Ivy) execute(ex string) (string, error) {
	iv.lastErr = nil

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

func (iv *Ivy) Execute(id string, ex string) (string, error) {
	id = fmt.Sprintf("v%s", id)

	_, err := iv.execute(fmt.Sprintf("%s = %s", id, ex))
	if err != nil {
		iv.logError(err)
		return "", err
	}

	result, err := iv.execute(id)
	if err != nil {
		iv.logError(err)
		return "", err
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

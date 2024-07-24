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

// log error normalises the error message and assigns it to the lastErr field.
// it also returns the lastErr field (ie. the normalised form)
func (iv *Ivy) logError(err error) error {
	spl := strings.SplitN(err.Error(), ":", 3)
	if len(spl) > 0 {
		iv.lastErr = errors.New(spl[len(spl)-1])
	}
	return iv.lastErr
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

func (iv *Ivy) Execute(ref string, ex string) (string, error) {
	ref = iv.WrapCellReference(ref)

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

// cell references are prefixed with this rune in order to prevent them looking
// like hexadecimal numbers, which would be a problem in some contexts
//
// this is only required due to the current method of storing variables in ivy.
// there may be a better solution and so might be removed in the future
const cellReferencePrefix = 'v'

// cell reference is converted so that it is safe to use with ivy in all
// instances. it only needs to be called when the cell appears in a cell entry.
// ie. something that will be evaluated by ivy
func (iv Ivy) WrapCellReference(ref string) string {
	return fmt.Sprintf("%c%s", cellReferencePrefix, ref)
}

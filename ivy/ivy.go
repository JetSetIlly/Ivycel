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

type Ivy struct {
	conf    config.Config
	context value.Context

	lastResultBuffer []byte
	lastErrorBuffer  []byte
	lastResult       *bytes.Buffer
	lastError        *bytes.Buffer
}

func New() Ivy {
	var iv Ivy

	iv.lastResult = bytes.NewBuffer(iv.lastResultBuffer)
	iv.lastError = bytes.NewBuffer(iv.lastErrorBuffer)
	iv.conf.SetOutput(iv.lastResult)
	iv.conf.SetErrOutput(iv.lastError)

	iv.context = exec.NewContext(&iv.conf)
	return iv
}

func (iv Ivy) execute(ex string) (string, error) {
	iv.lastResult.Reset()
	iv.lastError.Reset()

	scanner := scan.New(iv.context, "ivy", strings.NewReader(ex))
	parser := parse.NewParser("ivy", scanner, iv.context)

	ok := run.Run(parser, iv.context, false)
	if !ok {
		return "", errors.New(iv.lastError.String())
	}

	return iv.lastResult.String(), nil
}

func (iv Ivy) Execute(id string, ex string) (string, error) {
	_, err := iv.execute(fmt.Sprintf("%s = %s", id, ex))
	if err != nil {
		return "", err
	}

	result, err := iv.execute(id)
	if err != nil {
		return "", err
	}

	return result, nil
}

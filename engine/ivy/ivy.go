package ivy

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jetsetilly/ivycel/engine"
	"github.com/jetsetilly/ivycel/references"
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

	// currBase is the most recent base setting to be given to ivy. it is not
	// the same as the base field which should be though of as the current
	// default base for spreadsheet cells
	currBase engine.Base

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

func (iv *Ivy) tidyError(err error) error {
	msg := strings.TrimSpace(err.Error())
	msg = references.EngineToCellReference(msg)

	spl := strings.SplitN(msg, ": ", 3)
	if len(spl) > 0 {
		err = errors.New(spl[len(spl)-1])
	} else {
		err = errors.New(msg)
	}
	return err
}

// log error normalises the error message and assigns it to the lastErr field.
// it also returns the normalised form for further use
//
// the original error message is also printed with the log package
func (iv *Ivy) logError(err error) error {
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
	// calls to WithErrorSuppression() may be nested
	errorSuppression := iv.errorSuppression
	iv.errorSuppression = true
	with()
	iv.errorSuppression = errorSuppression
}

// run the supplied function but with the error suppression flag set
func (iv *Ivy) WithNumberBase(base engine.Base, with func()) {
	// calls to WithNumberBase() may be nested
	currBase := iv.currBase
	iv.setBase(base)
	with()
	iv.setBase(currBase)
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
	ref, ex = references.CellToEngineReference(ref, ex)

	if strings.HasPrefix(ex, ")") {
		return "", iv.logError(errors.New("Special Commands Not Supported"))
	}

	if strings.HasPrefix(ex, "opdelete ") {
		return "", iv.logError(errors.New("User-Defined Operations Not Supported"))
	}

	// check for user-defined operator keyword
	if strings.HasPrefix(ex, "op ") {
		return "", iv.logError(errors.New("User-Defined Operations Not Supported"))
	}

	// other expressions are executed and assigned to a variable name
	// representing a cell

	_, err := iv.execute(fmt.Sprintf("%s = %s", ref, ex))
	if err != nil {
		return "", iv.logError(iv.tidyError(err))
	}

	result, err := iv.execute(ref)
	if err != nil {
		return "", iv.logError(iv.tidyError(err))
	}

	return result, nil
}

// shape of the value at the supplied reference. ref should not be wrapped
func (iv *Ivy) Shape(ref string) string {
	ref, _ = references.CellToEngineReference(ref, "")

	var shp string
	var err error

	iv.WithErrorSupression(func() {
		shp, err = iv.execute(fmt.Sprintf("rho %s", ref))
		if err != nil {
			log.Printf("ivy: Shape: %s", iv.tidyError(err))
		}
	})

	return shp
}

func (iv *Ivy) setBase(base engine.Base) {
	iv.WithErrorSupression(func() {
		_, err := iv.execute(fmt.Sprintf(")ibase %d", base.Input))
		if err != nil {
			log.Printf("ivy: setBase: %s", iv.tidyError(err))
		}

		_, err = iv.execute(fmt.Sprintf(")obase %d", base.Output))
		if err != nil {
			log.Printf("ivy: setBase: %s", iv.tidyError(err))
		}
	})

	iv.currBase = base
}

func (iv *Ivy) SetBase(base engine.Base) {
	iv.base = base
	iv.setBase(base)
}

func (iv Ivy) Base() engine.Base {
	return iv.base
}

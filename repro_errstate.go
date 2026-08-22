//go:build ignore

package main

import (
	"errors"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

func main() {
	inner := fmt.Errorf("%w: approved interpretation is required", model.ErrState)
	wrapped := fmt.Errorf("export report failed: %v", inner)
	fixed := fmt.Errorf("export report failed: %w", inner)

	println("inner   errors.Is(ErrState) =", errors.Is(inner, model.ErrState))
	println("wrapped errors.Is(ErrState) =", errors.Is(wrapped, model.ErrState))
	println("fixed   errors.Is(ErrState) =", errors.Is(fixed, model.ErrState))
}

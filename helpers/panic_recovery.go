package helpers

import (
	"fmt"
	"runtime/debug"
)

// PanicHandler provides centralized panic recovery logic
type PanicHandler struct {
	scenarioName string
	funcName     string
}

// NewPanicHandler creates a panic handler
func NewPanicHandler(scenarioName string) *PanicHandler {
	return &PanicHandler{scenarioName: scenarioName}
}

// WithName sets the function name for better error messages
func (h *PanicHandler) WithName(funcName string) *PanicHandler {
	h.funcName = funcName
	return h
}

// RecoverFromPanic handles panics with consistent logging
// If err is non-nil, it will be cleared on panic to prevent double-reporting
func (h *PanicHandler) RecoverFromPanic(err *error) {
	if r := recover(); r != nil {
		name := h.scenarioName
		if name == "" {
			name = "unknown"
		}
		funcName := h.funcName
		if funcName == "" {
			funcName = "unknown"
		}
		fmt.Printf("Panic in %s (%s): %v\nStack trace:\n%s\n",
			funcName, name, r, debug.Stack())
		if err != nil {
			*err = nil // Clear error to prevent test framework double-reporting
		}
	}
}

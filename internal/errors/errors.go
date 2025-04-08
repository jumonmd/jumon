// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package errors

import (
	"errors"
	"fmt"
)

type ServiceError struct {
	// Status Code
	Code        int
	Description string
}

func New(code int, description string) *ServiceError {
	return &ServiceError{Code: code, Description: description}
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Description)
}

func (e *ServiceError) Wrap(err error) error {
	return fmt.Errorf("%s: %w", e.Error(), err)
}

// ServiceError returns the error for NATS micro service.
func (e *ServiceError) ServiceError(err error) (string, string, []byte) {
	return fmt.Sprintf("%d", e.Code), e.Description + ": " + err.Error(), nil
}

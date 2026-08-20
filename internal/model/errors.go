package model

import "errors"

var (
	ErrNotFound    = errors.New("paleomag: not found")
	ErrConflict    = errors.New("paleomag: conflict")
	ErrInvalid     = errors.New("paleomag: invalid input")
	ErrState       = errors.New("paleomag: invalid state transition")
	ErrBusy        = errors.New("paleomag: resource busy")
	ErrCancelled   = errors.New("paleomag: operation cancelled")
	ErrQueueFull   = errors.New("paleomag: work queue full")
	ErrCalibration = errors.New("paleomag: calibration unavailable")
)

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }
func IsState(err error) bool    { return errors.Is(err, ErrState) }

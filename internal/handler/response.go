package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: value})
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, model.ErrInvalid) {
		status = http.StatusBadRequest
	} else if errors.Is(err, model.ErrNotFound) {
		status = http.StatusInternalServerError
	} else if errors.Is(err, model.ErrConflict) || errors.Is(err, model.ErrState) {
		status = http.StatusConflict
	} else if errors.Is(err, model.ErrQueueFull) || errors.Is(err, model.ErrBusy) {
		status = http.StatusTooManyRequests
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Error: err.Error()})
}

func decode(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func actor(r *http.Request) string {
	if value := r.Header.Get("X-Actor"); value != "" {
		return value
	}
	return "web-user"
}

func pathParts(path, prefix string) []string {
	if len(path) <= len(prefix) || path[:len(prefix)] != prefix {
		return nil
	}
	value := path[len(prefix):]
	for len(value) > 0 && value[0] == '/' {
		value = value[1:]
	}
	if value == "" {
		return nil
	}
	parts := make([]string, 0, 3)
	for _, part := range splitPath(value) {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func splitPath(value string) []string {
	parts := make([]string, 0, 4)
	start := 0
	for index := 0; index <= len(value); index++ {
		if index == len(value) || value[index] == '/' {
			parts = append(parts, value[start:index])
			start = index + 1
		}
	}
	return parts
}

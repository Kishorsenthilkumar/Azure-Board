package main

import (
	"errors"
	"net/http"
	"strconv"
)

var errInvalidID = errors.New("invalid ID")

func readIDParam(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errInvalidID
	}

	return id, nil
}

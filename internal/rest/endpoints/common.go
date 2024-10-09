/*
 * Copyright (C) 2024 by Jason Figge
 */

package endpoints

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	managers2 "us.figge.auto-ssh/internal/managers"
)

const (
	id = "id"
)

var (
	ErrWriterFlush  = fmt.Errorf("unable to flush writer")
	ErrEncodeOutput = fmt.Errorf("failed to encode output")
)

func handleErrorResponse(resp http.ResponseWriter, err error) {
	httpStatus := http.StatusInternalServerError
	switch {
	case errors.Is(errors.Unwrap(err), managers2.ErrHostNotFound):
		httpStatus = http.StatusNotFound
	case errors.Is(errors.Unwrap(err), managers2.ErrTunnelNotFound):
		httpStatus = http.StatusNotFound
	}
	resp.WriteHeader(httpStatus)
	resp.Write([]byte(err.Error()))
}

func handleOutputResponse(resp http.ResponseWriter, output any) {
	if output == nil || reflect.ValueOf(output).IsNil() {
		resp.WriteHeader(http.StatusNoContent)
	} else {
		b := bytes.Buffer{}
		writer := bufio.NewWriter(&b)
		if err := json.NewEncoder(writer).Encode(output); err != nil {
			handleErrorResponse(resp, fmt.Errorf("%w: %v", ErrEncodeOutput, err))
			return
		}
		if err := writer.Flush(); err != nil {
			handleErrorResponse(resp, fmt.Errorf("%w: %v", ErrWriterFlush, err))
			return
		}

		resp.Header().Set("Content-Type", "application/json")
		resp.Header().Set("Content-Length", fmt.Sprintf("%d", len(b.Bytes())))
		resp.WriteHeader(http.StatusOK)
		resp.Write(b.Bytes())
	}
}

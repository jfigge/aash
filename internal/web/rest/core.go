/*
 * Copyright (C) 2024 by Jason Figge
 */

package rest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"us.figge.auto-ssh/internal/web/managers"
	managerModels "us.figge.auto-ssh/internal/web/models"
)

var (
	ErrWriterFlush  = fmt.Errorf("unable to flush writer")
	ErrEncodeOutput = fmt.Errorf("failed to encode output")
)

func extractOptions(req *http.Request) []managerModels.HostOptionFunc {
	var options []managerModels.HostOptionFunc
	return options
}

func handleErrorResponse(resp http.ResponseWriter, err error) {
	httpStatus := http.StatusInternalServerError
	switch {
	case errors.Is(errors.Unwrap(err), managers.ErrHostNotFound):
		httpStatus = http.StatusNotFound
	}
	resp.Write([]byte(err.Error()))
	resp.WriteHeader(httpStatus)
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

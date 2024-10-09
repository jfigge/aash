/*
 * Copyright (C) 2024 by Jason Figge
 */

package managers

import (
	"math/rand"
	"strings"
	"time"
	"unsafe"

	"us.figge.auto-ssh/internal/core/utils/cache"
	"us.figge.auto-ssh/internal/rest/models"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var (
	src = rand.NewSource(time.Now().UnixNano())
)

func Page[S any](items []S, p models.PaginationInput, listCache *cache.Cache[string, []S]) ([]S, *string) {
	if p.MaxResults > len(items) {
		p.MaxResults = len(items)
	}
	var more *string
	if p.MaxResults < len(items) {
		more = RandString(16)
		listCache.Add(*more, items[p.MaxResults:])
	}
	return items[:p.MaxResults], more
}

func RandString(n int) *string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return (*string)(unsafe.Pointer(&b))
}

func contains(sliceA, sliceB []string) bool {
	for _, a := range sliceA {
		for _, b := range sliceB {
			if strings.EqualFold(a, b) {
				return true
			}
		}
	}
	return false
}

func ExtractTunnelOptions(opts []models.TunnelOptionFunc) *models.TunnelOptions {
	options := &models.TunnelOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

package goscf

import "strings"

type Header map[string]string

func (h Header) Get(key string) string {
	v, ok := h[strings.ToLower(key)]
	if !ok {
		return ""
	}
	return v
}

func (h Header) Set(key, value string) {
	h[key] = value
}

func (h Header) Add(key, value string) {
	v, ok := h[key]
	if !ok {
		h[key] = value
	} else {
		h[key] = v + "," + value
	}
}

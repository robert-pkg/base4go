package config

import (
	"sync/atomic"
)

var (
	apiMapping atomic.Value // 存 map[string]string
)

type apiMappintItem struct {
	Scheme  string
	Service string
	Method  string
}

func (a *apiMappintItem) Equal(b *apiMappintItem) bool {
	if a == nil || b == nil {
		return a == b
	}

	return a.Scheme == b.Scheme &&
		a.Service == b.Service &&
		a.Method == b.Method
}

func setApiMapping(v map[string]*apiMappintItem) {
	apiMapping.Store(v)
}

func GetApiMappingMap() map[string]*apiMappintItem {
	anyValue := apiMapping.Load()
	if anyValue == nil {
		return map[string]*apiMappintItem{}
	}

	return anyValue.(map[string]*apiMappintItem)
}

func GetApiMapping(addr string) (scheme, service, method string, ok bool) {
	anyValue := apiMapping.Load()
	if anyValue == nil {
		ok = false
		return
	}

	var v *apiMappintItem
	m := anyValue.(map[string]*apiMappintItem)
	if v, ok = m[addr]; ok {
		scheme = v.Scheme
		service = v.Service
		method = v.Method
	}
	return
}

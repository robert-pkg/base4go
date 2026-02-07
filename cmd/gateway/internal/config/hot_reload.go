package config

import (
	"github.com/robert-pkg/base4go/cmd/gateway/internal/utils"
	"github.com/robert-pkg/base4go/log"
)

func hotReload(cfg *Config) {

	newApiMapping := cfg.apiMapping
	oldApiMapping := GetApiMappingMap()

	if !utils.IsMappingEqual(newApiMapping, oldApiMapping, func(a, b *apiMappintItem) bool {
		return a.Equal(b)
	}) {
		for k, v := range newApiMapping {
			log.Infof("api mapping changed. key=%v, old=%v, new=%v", k, oldApiMapping[k], v)
		}

		setApiMapping(newApiMapping)
	}

}

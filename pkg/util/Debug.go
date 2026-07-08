package util

import (
	"fmt"
)

func (cfg *Config) DebugLog(message ...any) {
	if cfg.Logging {
		fmt.Println(message...)
	}
}

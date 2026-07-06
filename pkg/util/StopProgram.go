package util

import (
	"fmt"
	"os"
)

func StopProgram(code int) {
	fmt.Scanln()
	os.Exit(code)
}

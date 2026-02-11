package lib

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var loadDotenvOnce sync.Once

func GetEnv(name string) (string, error) {
	loadDotenvOnce.Do(func() {
		_ = godotenv.Load()
	})

	msg := fmt.Sprintf("%s is not defined in environment.", name)

	env := os.Getenv(name)
	if env == "" {
		// Retry loading to avoid sticky misses when env is populated after first access.
		_ = godotenv.Load()
		env = os.Getenv(name)
	}

	if env == "" {
		return "", errors.New(msg)
	}

	return env, nil
}

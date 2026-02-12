package lib

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func GetEnv(name string) (string, error) {
	godotenv.Load()

	env := os.Getenv(name)
	if env == "" {
		return "", fmt.Errorf("%s is not defined in environment.", name)
	}

	return env, nil
}

package handler

import (
	"fmt"
	"strings"
	"sync"

	"s2-leads-api/lib"

	"github.com/gofiber/fiber/v3"
	unkey "github.com/unkeyed/sdks/api/go/v2"
	"github.com/unkeyed/sdks/api/go/v2/models/components"
)

var (
	unkeyClient   *unkey.Unkey
	unkeyInitErr  error
	unkeyInitOnce sync.Once
)

func getUnkeyClient() (*unkey.Unkey, error) {
	unkeyInitOnce.Do(func() {
		rootKey, err := lib.GetEnv("UNKEY_ROOT_KEY")
		if err != nil {
			unkeyInitErr = fmt.Errorf("failed to load UNKEY_ROOT_KEY: %w", err)
			return
		}

		unkeyClient = unkey.New(
			unkey.WithSecurity(rootKey),
		)
	})

	return unkeyClient, unkeyInitErr
}

func unkeyAuth(c fiber.Ctx) error {
	client, err := getUnkeyClient()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "verification service unavailable",
		})
	}

	apiKey := extractAPIKey(c.Get("Authorization"))
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing or invalid Authorization header",
		})
	}

	res, err := client.Keys.VerifyKey(c.Context(), components.V2KeysVerifyKeyRequestBody{
		Key: apiKey,
	})
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "verification service unavailable",
		})
	}

	if res == nil || res.V2KeysVerifyKeyResponseBody == nil || !res.V2KeysVerifyKeyResponseBody.Data.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid API key",
		})
	}

	if res.V2KeysVerifyKeyResponseBody.Data.KeyID != nil {
		c.Locals("keyId", *res.V2KeysVerifyKeyResponseBody.Data.KeyID)
	}

	return c.Next()
}

func extractAPIKey(authHeader string) string {
	header := strings.TrimSpace(authHeader)
	if header == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[len("bearer "):])
	}

	// Accept raw token value for simplicity.
	return header
}

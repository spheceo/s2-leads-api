package lib

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	unkey "github.com/unkeyed/sdks/api/go/v2"
	"github.com/unkeyed/sdks/api/go/v2/models/components"
)

func UnkeyAuth(c fiber.Ctx) error {
	rootKey, err := GetEnv("UNKEY_ROOT_KEY")
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "verification service unavailable",
		})
	}

	client := unkey.New(
		unkey.WithSecurity(rootKey),
	)

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

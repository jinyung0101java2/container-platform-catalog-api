package handler

import "github.com/gofiber/fiber/v2"

type HealthStatus struct {
	Status string   `json:"status"`
	Groups []string `json:"groups,omitempty"`
}

const StatusUp = "UP"

func Health(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(HealthStatus{
		Status: StatusUp,
		Groups: []string{"liveness", "readiness"},
	})
}
func HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(HealthStatus{
		Status: StatusUp,
	})
}

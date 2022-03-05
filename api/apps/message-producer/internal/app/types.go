package app

import "github.com/gofiber/fiber/v2"

type App interface {
	GetRouter() fiber.Router
	Init(*fiber.App)
}

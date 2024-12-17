package middleware

import (
	"errors"
	"fmt"
	"github.com/gofiber/contrib/fiberi18n/v2"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang-jwt/jwt/v5"
	"go-api/common"
	"go-api/config"
	"golang.org/x/text/language"
)

func FiberMiddleware(a *fiber.App) {
	a.Use(
		// Add i18n
		fiberi18n.New(&fiberi18n.Config{
			RootPath:         "./localize",
			AcceptLanguages:  []language.Tag{language.English, language.Korean},
			DefaultLanguage:  language.English,
			FormatBundleFile: "json",
		}),
		// Add CORS to each route.
		cors.New(),
		// Add simple logger.
		logger.New(),
		// Add JWT Middleware
		jwtware.New(jwtware.Config{
			KeyFunc: customKeyFunc(),
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				return customErrorHandler(c, err)
			},
		}),
	)
}

func customKeyFunc() jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		// Always check the signing method
		if t.Method.Alg() != jwtware.HS512 {
			return nil, fmt.Errorf("Unexpected jwt signing method=%v", t.Header["alg"])
		}
		return []byte(config.Env.JwtSecret), nil
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	log.Errorf("JWT ErrorHandler:: %v", err.Error())
	switch {
	case errors.Is(err, jwtware.ErrJWTMissingOrMalformed):
		return common.RespErrStatus(c, fiber.StatusBadRequest, fmt.Errorf(common.MISSING_OR_MALFORMED_JWT))
	case errors.Is(err, jwt.ErrTokenExpired):
		return common.RespErrStatus(c, fiber.StatusUnauthorized, fmt.Errorf(common.TOKEN_EXPIRED))
	}
	return common.RespErrStatus(c, fiber.StatusUnauthorized, fmt.Errorf(common.TOKEN_FAILED))
}

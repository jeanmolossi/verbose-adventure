package stripslash

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func StripBackslash(s string) string {
	return strings.TrimPrefix(s, "/")
}

func StripTrailslash(s string) string {
	return strings.TrimSuffix(s, "/")
}

func ParamWithoutBackslash(c echo.Context, key string) string {
	return StripBackslash(c.Param(key))
}

func ParamWithoutAnySlashes(c echo.Context, key string) string {
	return StripBackslash(
		StripTrailslash(c.Param(key)),
	)
}

package ctxinfo

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jtorz/temp-secure-notes/app/config"
)

type modeKeyType string

const modeKey modeKeyType = "_mode_"

// LoggingLevel returns the level of logging int he context.
func LoggingLevel(ctx context.Context) config.LogginLvl {
	lvl, _ := ctx.Value(modeKey).(config.LogginLvl)
	return lvl
}

// LogginAllowed verifies if the logging is allowed in the context.
// Only the loggin is allowed when the required log is bigger (or equal) than the context.LogginLvl.
func LogginAllowed(ctx context.Context, lvl config.LogginLvl) bool {
	ctxLvl := LoggingLevel(ctx)
	return lvl >= ctxLvl
}

// SetLoggingLevel sets level of logging to the gin.Context.
func SetLoggingLevel(c *gin.Context, lvl config.LogginLvl) {
	c.Set(string(modeKey), lvl)
}

// SetLoggingLevel sets level of logging to the gin.Context.
func SetLoggingLevelC(parent context.Context, lvl config.LogginLvl) context.Context {
	return context.WithValue(parent, modeKey, lvl)
}

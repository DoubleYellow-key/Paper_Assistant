package middleware

import (
	"net/http"
	"strconv"
	"strings"

	apperrors "paper-assistant-backend/internal/pkg/errors"
	"paper-assistant-backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const userIDKey = "user_id"

// 演示版 token 约定: Authorization: Bearer uid-<number>
func AuthJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "missing token")
			c.Abort()
			return
		}
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "invalid token format")
			c.Abort()
			return
		}
		if !strings.HasPrefix(parts[1], "uid-") {
			response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "invalid token")
			c.Abort()
			return
		}
		userID, err := strconv.ParseUint(strings.TrimPrefix(parts[1], "uid-"), 10, 64)
		if err != nil || userID == 0 {
			response.Fail(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, "invalid token")
			c.Abort()
			return
		}
		c.Set(userIDKey, userID)
		c.Next()
	}
}

func MustUserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(userIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := v.(uint64)
	return userID, ok
}

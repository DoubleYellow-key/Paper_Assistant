package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{
		Code:    0,
		Message: "ok",
		Data:    data,
		TraceID: traceIDFromContext(c),
	})
}

func Fail(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Envelope{
		Code:    code,
		Message: message,
		TraceID: traceIDFromContext(c),
	})
}

func traceIDFromContext(c *gin.Context) string {
	v, ok := c.Get("trace_id")
	if !ok {
		return ""
	}
	traceID, _ := v.(string)
	return traceID
}

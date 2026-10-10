package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func http_MaxBytes(c *gin.Context, n int64) io.ReadCloser {
	return http.MaxBytesReader(c.Writer, c.Request.Body, n)
}

func httpDetect(b []byte) string { return http.DetectContentType(b) }

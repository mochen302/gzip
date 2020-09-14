package main

import (
	"compress/gzip"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"github.com/mochen302/gzipx"
)

func main() {
	r := gin.Default()
	r.Use(gzipx.Gzip(gzip.DefaultCompression))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong "+fmt.Sprint(time.Now().Unix()))
	})

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

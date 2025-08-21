package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type configData struct {
	port      int
	uploadDir string
}

// Version of the service
const Version = "1.0.0"

func main() {
	log.Printf("===> ILLiad upload WS is staring up <===")
	var cfg configData
	flag.IntVar(&cfg.port, "port", 8080, "API service port (default 8080)")
	flag.StringVar(&cfg.uploadDir, "dir", "", "Upload directory")
	flag.Parse()

	if cfg.uploadDir == "" {
		log.Fatal("Parameter dir is required")
	}

	log.Printf("[CONFIG] port          = [%d]", cfg.port)
	log.Printf("[CONFIG] dir           = [%s]", cfg.uploadDir)

	// Set routes and start server
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.GET("/", versionHandler)
	router.GET("/version", versionHandler)
	router.GET("/healthcheck", healthCheckHandler)

	portStr := fmt.Sprintf(":%d", cfg.port)
	log.Printf("INFO: start service on port %s with CORS support enabled", portStr)
	log.Fatal(router.Run(portStr))
}

func versionHandler(c *gin.Context) {
	build := "unknown"

	files, _ := filepath.Glob("../buildtag.*")
	if len(files) == 1 {
		build = strings.Replace(files[0], "../buildtag.", "", 1)
	}

	vMap := make(map[string]string)
	vMap["version"] = Version
	vMap["build"] = build
	c.JSON(http.StatusOK, vMap)
}

func healthCheckHandler(c *gin.Context) {
	type healthcheck struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message"`
	}

	hcMap := make(map[string]healthcheck)
	hcMap["illiad-upload"] = healthcheck{Healthy: true}

	c.JSON(http.StatusOK, hcMap)
}

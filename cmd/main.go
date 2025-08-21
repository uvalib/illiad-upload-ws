package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/virgo4-jwt/v4jwt"
)

type serviceContext struct {
	port      int
	uploadDir string
	jwtKey    string
}

// Version of the service
const Version = "1.0.0"

func main() {
	log.Printf("===> ILLiad upload WS is staring up <===")
	var ctx serviceContext
	flag.IntVar(&ctx.port, "port", 8080, "API service port (default 8080)")
	flag.StringVar(&ctx.uploadDir, "dir", "", "Upload directory")
	flag.StringVar(&ctx.jwtKey, "jwtkey", "", "V4 JWT key")
	flag.Parse()

	if ctx.uploadDir == "" {
		log.Fatal("Parameter dir is required")
	}
	if ctx.jwtKey == "" {
		log.Fatal("Parameter jwtkey is required")
	}

	log.Printf("[CONFIG] port          = [%d]", ctx.port)
	log.Printf("[CONFIG] dir           = [%s]", ctx.uploadDir)

	// Set routes and start server
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.GET("/", versionHandler)
	router.GET("/version", versionHandler)
	router.GET("/healthcheck", healthCheckHandler)
	router.POST("/upload", ctx.authMiddleware, ctx.uploadHandler)

	portStr := fmt.Sprintf(":%d", ctx.port)
	log.Printf("INFO: start service on port %s with CORS support enabled", portStr)
	log.Fatal(router.Run(portStr))
}

func (svc *serviceContext) authMiddleware(c *gin.Context) {
	tokenStr, err := getBearerToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		log.Printf("Authentication failed: [%s]", err.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if tokenStr == "undefined" {
		log.Printf("Authentication failed; bearer token is undefined")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("Validating JWT auth token...")
	v4Claims, jwtErr := v4jwt.Validate(tokenStr, svc.jwtKey)
	if jwtErr != nil {
		log.Printf("JWT signature for %s is invalid: %s", tokenStr, jwtErr.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// add the parsed claims and signed JWT string to the request context so other handlers can access it.
	c.Set("jwt", tokenStr)
	c.Set("claims", v4Claims)
	log.Printf("got bearer token: [%s]: %+v", tokenStr, v4Claims)
}

func (svc *serviceContext) uploadHandler(c *gin.Context) {
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

func getBearerToken(authorization string) (string, error) {
	components := strings.Split(strings.Join(strings.Fields(authorization), " "), " ")

	// must have two components, the first of which is "Bearer", and the second a non-empty token
	if len(components) != 2 || components[0] != "Bearer" || components[1] == "" {
		return "", fmt.Errorf("Invalid Authorization header: [%s]", authorization)
	}

	return components[1], nil
}

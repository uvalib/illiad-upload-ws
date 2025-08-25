package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
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
	router.GET("/favicon.ico", ignoreFavicon)
	router.GET("/version", versionHandler)
	router.GET("/healthcheck", ctx.healthCheckHandler)
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
	log.Printf("INFO: received new upload request with content type %s", c.ContentType())
	log.Printf("Dump all request headers ==================================")
	for name, values := range c.Request.Header {
		for _, value := range values {
			log.Printf("%s=%s\n", name, value)
		}
	}
	log.Printf("END header dump ===========================================")

	formData, err := c.MultipartForm()
	if err != nil {
		log.Printf("ERROR: unable to get multipartform data from request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	formFile := formData.File["file"][0]
	destFile := path.Join(svc.uploadDir, formFile.Filename)
	log.Printf("INFO: request contains file %s, save it to %s", formFile.Filename, destFile)
	err = c.SaveUploadedFile(formFile, destFile)
	if err != nil {
		log.Printf("ERROR: unable to save %s: %s", formFile.Filename, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("received %s", formFile.Filename))
}

func ignoreFavicon(c *gin.Context) {
	// no-op; just here to prevent errors when request made from browser
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

func (svc *serviceContext) healthCheckHandler(c *gin.Context) {
	type healthcheck struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message"`
	}

	hcMap := make(map[string]healthcheck)
	hcMap["illiad-upload"] = healthcheck{Healthy: true}

	fi, err := os.Stat(svc.uploadDir)
	if err != nil {
		log.Printf("ERROR: upload directory %s check failed: %s", svc.uploadDir, err.Error())
	} else {
		groupID := os.Getegid()
		userID := os.Geteuid()
		modeStr := fmt.Sprintf("%v", fi.Mode())
		log.Printf("INFO: upload directory %s permissions [%s]; current user:group [%d:%d] ", svc.uploadDir, modeStr, userID, groupID)
	}

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

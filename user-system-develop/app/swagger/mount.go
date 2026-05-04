package swagger

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml index.html
var assets embed.FS

// Mount регистрирует Swagger UI и спецификацию OpenAPI по путям /swagger/...
func Mount(engine *gin.Engine) {
	engine.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})
	engine.GET("/swagger/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})
	engine.GET("/swagger/index.html", serve("index.html", "text/html; charset=utf-8"))
	engine.GET("/swagger/openapi.yaml", serve("openapi.yaml", "application/yaml; charset=utf-8"))
}

func serve(name, contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := assets.ReadFile(name)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, contentType, data)
	}
}

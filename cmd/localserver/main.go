package main

import (
	"net/http"

	"github.com/cheriot/kubenav/internal/app"

	log "github.com/sirupsen/logrus"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/api/contexts", func(c echo.Context) error {
		ctxNames, err := app.KubeContextList()
		if err != nil {
			log.Errorf("error from KubeContextList: %w", err)
		}
		return c.JSON(http.StatusOK, ctxNames)
	})

	e.GET("/api/context/:ctx/namespaces", func(c echo.Context) error {
		nsNames, err := app.KubeNamespaceList(c.Request().Context(), c.Param("ctx"))
		if err != nil {
			log.Errorf("error from KubeNamespaceList: %w", err)
		}
		return c.JSON(http.StatusOK, nsNames)
	})

	e.POST("/api/context/:ctx/namespace/:ns/default", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"result": "ok"})
	})

	e.GET("/api/context/:ctx/namespace/:ns/query", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string][]string{
			"pods":     {"pod/mything-blue-asdf", "pod/mything-blue-fdsa", "pod/mything-blue-vzbs", "pod/mything-green-hgfd"},
			"services": {"service/mything", "service/mything-blue", "service/mything-green"},
		})
	})

	e.Logger.Fatal(e.Start(":4000"))
}

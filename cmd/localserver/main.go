package main

import (
	"net/http"

	"github.com/cheriot/kubenav/pkg/app"

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
			log.Errorf("error from KubeContextList: %s", err)
		}
		return c.JSON(http.StatusOK, ctxNames)
	})

	e.GET("/api/context/:ctx/namespaces", func(c echo.Context) error {
		ctx := c.Request().Context()
		ctxParam := c.Param("ctx")

		kc, err := app.GetOrMakeKubeCluster(ctx, ctxParam)
		if err != nil {
			log.Errorf("error getting kubecluster for %s: %s", ctxParam, err)
		}

		nsNames, err := kc.KubeNamespaceList(ctx)
		if err != nil {
			log.Errorf("error from KubeNamespaceList: %s", err)
		}
		return c.JSON(http.StatusOK, nsNames)
	})

	e.GET("/api/context/:ctx/namespace/:ns/query/:query", func(c echo.Context) error {
		ctx := c.Request().Context()
		ctxParam := c.Param("ctx")
		nsParam := c.Param("ns")
		queryParam := c.Param("query")

		kc, err := app.GetOrMakeKubeCluster(ctx, ctxParam)
		if err != nil {
			log.Errorf("error getting kubecluster for %s: %s", ctxParam, err)
		}

		resourceTables, err := kc.Query(ctx, nsParam, queryParam)
		if err != nil {
			log.Errorf("error query %s for %s: %s", queryParam, ctxParam, err)
		}

		return c.JSON(http.StatusOK, resourceTables)
	})

	e.Logger.Fatal(e.Start(":4000"))
}

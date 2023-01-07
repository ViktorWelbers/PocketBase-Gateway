// main.go
package main

import (
	"api-gateway/config"
	"api-gateway/handlers"
	"api-gateway/httpclient"
	"api-gateway/messaging"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"log"
	"net/http"
)

func AddHealthcheck(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		_, err := e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "api/health",
			Handler: func(c echo.Context) error {
				return c.JSON(http.StatusOK, map[string]string{"status": "OK"})
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func addHandlerToRouter(app *pocketbase.PocketBase, handler handlers.Handler, method string, path string) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		_, err := e.Router.AddRoute(echo.Route{
			Method:  method,
			Path:    path,
			Handler: handler.Handle,
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				apis.RequireRecordAuth(),
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func main() {
	app := pocketbase.New()
	AddHealthcheck(app)
	httpClient := httpclient.NewImageHttpClient(config.ImageServiceUrl, config.ImageServiceApiKey)
	queueClient := messaging.NewQueueClient(config.ReceiveQueueName, config.PublishQueueName, config.BusConnectionString)
	addHandlerToRouter(app, handlers.NewMessageProducerHandler(queueClient), "GET", "api/send_message")
	addHandlerToRouter(app, handlers.NewImageDownloadHandler(queueClient, httpClient), "GET", "api/get_image")
	addHandlerToRouter(app, handlers.NewAuthenticationHandler(httpClient), "GET", "api/check_auth")
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}

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
)

func addHandlerToRouter(app *pocketbase.PocketBase, handler handlers.Handler, method string, path string) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		_, err := e.Router.AddRoute(echo.Route{
			Method:  method,
			Path:    path,
			Handler: handler.Handle,
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				//apis.RequireRecordAuth(),
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
	httpClient := httpclient.NewImageHttpClient(config.ImageServiceUrl, config.ImageServiceApiKey)
	queueClient := messaging.NewQueueClient(config.ReceiveQueueName, config.PublishQueueName, config.BusConnectionString)
	addHandlerToRouter(app, handlers.NewMessageProducerHandler(queueClient), "GET", "api/send_message")
	addHandlerToRouter(app, handlers.NewImageDownloadHandler(queueClient, httpClient), "GET", "api/get_image")
	addHandlerToRouter(app, handlers.NewImageUploadHandler(httpClient), "POST", "api/upload/")
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}

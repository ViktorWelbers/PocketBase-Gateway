// main.go
package main

import (
	"api-gateway/internal/handlers"
	"api-gateway/internal/httpclient"
	"api-gateway/internal/messaging"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"log"
	"net/http"
	"os"
)

type Handler interface {
	Handle(c echo.Context) error
}

func addHealthcheck(app *pocketbase.PocketBase) {
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

func addHandlerToRouter(app *pocketbase.PocketBase, handler Handler, method string, path string) {
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

func setupRoutes(app *pocketbase.PocketBase, queueClient *messaging.QueueClient, httpClient *httpclient.ImageHttpClient) {
	addHealthcheck(app)
	addHandlerToRouter(app, handlers.NewMessageProducerHandler(queueClient), "GET", "api/send_message")
	addHandlerToRouter(app, handlers.NewImageDownloadHandler(queueClient, httpClient), "GET", "api/get_image")
	addHandlerToRouter(app, handlers.NewAuthenticationHandler(), "GET", "api/check_auth")
}

func main() {
	BusConnectionString := os.Getenv("AZURE_SERVICEBUS_CONNECTION_STRING")
	PublishQueueName := os.Getenv("AZURE_PUBLISH_QUEUE_NAME")
	ReceiveQueueName := os.Getenv("AZURE_RECEIVE_QUEUE_NAME")
	ImageServiceUrl := os.Getenv("IMAGE_SERVICE_URL")
	ImageServiceApiKey := os.Getenv("IMAGE_SERVICE_API_KEY")
	if BusConnectionString == "" || PublishQueueName == "" || ReceiveQueueName == "" || ImageServiceUrl == "" || ImageServiceApiKey == "" {
		log.Fatal("Missing environment variables")
	}
	app := pocketbase.New()
	httpClient := httpclient.NewImageHttpClient(ImageServiceUrl, ImageServiceApiKey)
	queueClient := messaging.NewQueueClient(BusConnectionString, PublishQueueName, ReceiveQueueName)
	defer queueClient.Close()
	setupRoutes(app, queueClient, httpClient)
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}

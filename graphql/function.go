package main

import (
	"context"
	"crypto/sha256"
	"net/http"
	"os"
	"sync"

	"github.com/graphql-editor/azure-functions-golang-worker/api"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/handlers"
	azuredriver "github.com/graphql-editor/stucco/pkg/providers/azure/driver"
	azurehandler "github.com/graphql-editor/stucco/pkg/providers/azure/handler"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/utils"
	gqlhandler "github.com/graphql-go/handler"
)

var (
	lock      sync.Mutex
	configSHA [sha256.Size]byte
	handler   azurehandler.Handler
)

// HTTPTrigger is an example httpTrigger
type HTTPTrigger struct {
	Request  *api.Request `azfunc:"httpTrigger"`
	Response api.Response `azfunc:"res"`
}

// Run implements function behaviour
func (h *HTTPTrigger) Run(ctx context.Context, logger api.Logger) {
	handler, err := getHandler()
	if err != nil {
		logger.Errorf("could not get handler: %v", err)
		h.Response = api.Response{
			Headers: http.Header{
				"content-type": []string{"text/plain"},
			},
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}
		return
	}
	h.Response = handler.ServeHTTP(ctx, logger, h.Request)
}

// Function exports function entry point
var Function HTTPTrigger

func configChecksum() [sha256.Size]byte {
	return sha256.Sum256([]byte(os.Getenv(router.SchemaEnv) + os.Getenv(utils.StuccoConfigEnv)))
}

func getHandler() (azurehandler.Handler, error) {
	lock.Lock()
	rhandler := handler
	checksum := configSHA
	lock.Unlock()
	if rhandler.Handler != nil && checksum == configChecksum() {
		return rhandler, nil
	}
	driver.Register(driver.Config{
		Provider: "azure",
		Runtime:  "function",
	}, &azuredriver.Driver{})
	router.SetDefaultEnvironment(router.Environment{
		Provider: "azure",
		Runtime:  "function",
	})
	var cfg router.Config
	if err := utils.LoadConfigFile("", &cfg); err != nil {
		return azurehandler.Handler{}, err
	}
	router, err := router.NewRouter(cfg)
	if err != nil {
		return azurehandler.Handler{}, err
	}
	lock.Lock()
	handler = azurehandler.Handler{
		Handler: handlers.WithProtocolInContext(gqlhandler.New(&gqlhandler.Config{
			Schema:   &router.Schema,
			Pretty:   true,
			GraphiQL: true,
		})),
	}
	configSHA = configChecksum()
	rhandler = handler
	lock.Unlock()
	return rhandler, nil
}

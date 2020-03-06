package main

import (
	"context"
	"fmt"
	"os"

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
	handler azurehandler.Handler
)

// HTTPTrigger is an example httpTrigger
type HTTPTrigger struct {
	Request  *api.Request `azfunc:"httpTrigger"`
	Response api.Response `azfunc:"res"`
}

// Run implements function behaviour
func (h *HTTPTrigger) Run(ctx context.Context, logger api.Logger) {
	h.Response = handler.ServeHTTP(ctx, logger, h.Request)
}

// Function exports function entry point
var Function HTTPTrigger

func init() {
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
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	router, err := router.NewRouter(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	handler = azurehandler.Handler{
		Handler: handlers.WithProtocolInContext(gqlhandler.New(&gqlhandler.Config{
			Schema:   &router.Schema,
			Pretty:   true,
			GraphiQL: true,
		})),
	}
}

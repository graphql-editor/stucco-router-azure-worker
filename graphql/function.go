package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/graphql-editor/azure-functions-golang-worker/api"
	azuredriver "github.com/graphql-editor/stucco-router-azure-worker/driver"
	azurehandler "github.com/graphql-editor/stucco-router-azure-worker/handler"
	"github.com/graphql-editor/stucco/pkg/driver"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/utils"
	gqlhandler "github.com/graphql-go/handler"
)

var (
	handler azurehandler.Handler
	config  = "./stucco"
)

func withProtocolInContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(
			rw,
			r.WithContext(
				context.WithValue(
					r.Context(),
					router.ProtocolKey, map[string]interface{}{
						"headers": r.Header,
					},
				),
			),
		)
	})
}

func recoveryHandler(next http.Handler, logger api.Logger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				logger.Errorf("%v\n", err)
				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte("There was an internal server error"))
			}
		}()
		next.ServeHTTP(rw, r)
	})
}

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
	if schema := os.Getenv("STUCCO_SCHEMA"); schema != "" {
		cfg.Schema = schema
	}
	if envConfig := os.Getenv("STUCCO_CONFIG"); envConfig != "" {
		config = envConfig
	}
	if err := utils.LoadConfigFile(config, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	router, err := router.NewRouter(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	handler = azurehandler.Handler{
		Handler: withProtocolInContext(gqlhandler.New(&gqlhandler.Config{
			Schema:   &router.Schema,
			Pretty:   true,
			GraphiQL: true,
		})),
	}
}

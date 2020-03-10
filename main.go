package main

import (
	"reflect"

	"github.com/graphql-editor/stucco-router-azure-worker/graphql/httptrigger"

	"github.com/graphql-editor/azure-functions-golang-worker/cmd/userworker"
)

func main() {
	userworker.Execute(map[string]reflect.Type{
		"graphql.Function": reflect.TypeOf(httptrigger.HTTPTrigger{}),
	})
}

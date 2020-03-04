GOBUILD=go build -mod=readonly -a -installsuffix cgo -ldflags="-w -s" -i

all: build_router build_worker 

build_worker:
	${GOBUILD} -o worker github.com/graphql-editor/azure-functions-golang-worker/cmd/worker

build_router:
	${GOBUILD} -buildmode=plugin -o graphql/function.so graphql/function.go

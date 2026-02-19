package server

import (
	"fmt"
	"net/http"
	"os"

	"gogogo/cmd/cli/commands/server/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func Run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	fmt.Printf("Starting GraphQL server on http://localhost:%s/\n", port)
	fmt.Printf("GraphQL playground available at http://localhost:%s/\n", port)
	return http.ListenAndServe(":"+port, nil)
}

package main

import (
	"gateway/internal/graph"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	// Create a new router
	router := http.NewServeMux()

	// Apply our simple auth middleware to the query endpoint
	router.Handle("/query", graph.AuthMiddleware(srv))
	router.Handle("/", playground.Handler("GraphQL playground", "/query"))

	log.Printf("🚀 Gateway connecting to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

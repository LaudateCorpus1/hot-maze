package main

import (
	"fmt"
	"log"
	"os"

	"context"

	hotmaze "github.com/Deleplace/hot-maze/B2"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	hotmaze.GetServer().BackendBaseURL = fmt.Sprintf("http://localhost:%s", port)

	ctx := context.Background()
	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/B2_SecureURLs", hotmaze.B2_SecureURLs); err != nil {
		log.Fatalf("funcframework.RegisterHTTPFunctionContext: %v\n", err)
	}
	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/B2_Get/", hotmaze.B2_Get); err != nil {
		log.Fatalf("funcframework.RegisterHTTPFunctionContext: %v\n", err)
	}
	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/B2_Forget", hotmaze.B2_Forget); err != nil {
		log.Fatalf("funcframework.RegisterHTTPFunctionContext: %v\n", err)
	}

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}

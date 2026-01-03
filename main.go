package main

import (
	"context"
	"rest_api_poc/config"
)

func main() {
	// create context
	ctx := context.Background()

	// Load config
	cfg := config.LoadConfig()

	// Start server, get shutdown function
	webDispose := startServer(cfg)
	defer webDispose(ctx)
}

/*
	1. Connect to postgres with retry mechanism and logging
	2. Create MODEL for persons
	3. API endpoints to CRUD persons with request validations
	4. Login mechanism with JWT token, check in headers and also for token
	5. Permission check to perform Write operations on persons api using a middleware
	6. Error middleware to handle error globally
	7. Error classes and different types of errors
	8. Logging middleware
	9. Tracing middleware
	10. CORS Middleware
	11. Combine all necessary middlewares into one
	12. Internationalization
	13. Swagger docs
	14. Docker compose
	15. Make file
*/

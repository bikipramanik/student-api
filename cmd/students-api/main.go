// package main tells the Go compiler that this file should compile as an executable program,
// rather than a reusable library/package.
package main

import (
	"context"
	"log" // Package log provides simple logging functions (e.g., logging errors and terminating)
	"log/slog"
	"net/http" // Package net/http provides HTTP client and server implementations
	"os"
	"os/signal"
	"syscall"
	"time"

	// Import our custom configuration package from the internal directory
	"github.com/bikipramanik/students-api/internal/config"
	"github.com/bikipramanik/students-api/internal/http/handlers/student"
	"github.com/bikipramanik/students-api/internal/storage/sqlite"
)

// main is the entry point function of the program. Execution begins here.
func main() {
	// -------------------------------------------------------------
	// 1. Load Configuration
	// -------------------------------------------------------------
	// MustLoad reads the YAML config file (specified via flags or env vars).
	// If loading fails, it will log the error and terminate the program immediately.
	cfg := config.MustLoad()

	// -------------------------------------------------------------
	// 2. Database Setup (placeholder for future implementation)
	// -------------------------------------------------------------
	// In future steps, database connection and migrations will be initialized here.

	storage, err := sqlite.New(cfg)

	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Storage Initiliazed", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))
	// -------------------------------------------------------------
	// 3. Setup Router (Request Multiplexer)
	// -------------------------------------------------------------
	// http.NewServeMux creates a new HTTP request multiplexer (router).
	// It matches the URL of each incoming request against a list of registered routes
	// and calls the corresponding handler.
	router := http.NewServeMux()

	// Register a route handler:
	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}", student.GetById(storage))
	router.HandleFunc("GET /api/students", student.GetList(storage))

	// -------------------------------------------------------------
	// 4. Configure HTTP Server
	// -------------------------------------------------------------
	// We create an instance of http.Server and set:
	// - Addr: the host and port to listen on (e.g., "localhost:8082"), read from our config.
	// - Handler: the router (ServeMux) that will dispatch incoming requests to our endpoints.
	server := http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}

	// -------------------------------------------------------------
	// 5. Start the Server with Graceful Shutdown
	// -------------------------------------------------------------
	slog.Info("Server started", slog.String("address", cfg.HTTPServer.Addr))

	// Step 1: Create a channel (a message pipe) to listen for OS signals.
	// Capacity 1 prevents missing a signal if the OS sends one quickly.
	done := make(chan os.Signal, 1)

	// Step 2: Tell the OS: "When the user presses Ctrl+C (SIGINT) or the system
	// asks to stop the process (SIGTERM), send that signal into our 'done' channel."
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Step 3: Run the server in a separate background thread (goroutine).
	// We do this because ListenAndServe() blocks execution forever while running.
	go func() {
		err := server.ListenAndServe()

		// If the server fails to start (e.g. port already occupied), log fatal and exit.
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %s", err.Error())
		}
	}()

	// Step 4: Block and pause here until a signal arrives in the 'done' channel.
	// The arrow '<-done' means: "wait and pull the signal out of the channel".
	<-done

	// Step 5: Begin graceful shutdown process using structured logger (slog).
	slog.Info("Shutting down this server...")

	// Step 6: Create a Context with a 5-second deadline.
	// This gives existing requests up to 5 seconds to finish before forcing a stop.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Step 7: Clean up resources used by the timeout timer when main() finishes.
	defer cancel()

	// Step 8: server.Shutdown stops accepting new requests and waits for ongoing
	// requests to complete within the 5-second context limit.
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shut down server", slog.String("Error", err.Error()))
	}

	slog.Info("Server shutdown successfully!!")

}

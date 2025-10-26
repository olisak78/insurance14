package testutils

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

// TestMain runs before all tests and ensures proper cleanup
// This ensures Docker cleanup even when running `go test ./...` directly
func TestMain(m *testing.M) {
	// Set up signal handling for graceful cleanup on interruption (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Run cleanup in a goroutine that listens for signals
	go func() {
		<-c
		log.Println("\n🛑 Received interrupt signal, cleaning up Docker containers...")
		CleanupSharedContainer()
		os.Exit(1)
	}()

	// Run tests
	log.Println("🧪 Starting test suite with Docker cleanup enabled...")
	code := m.Run()

	// Always cleanup when tests finish normally
	log.Println("✅ Tests completed, cleaning up Docker containers...")
	CleanupSharedContainer()

	// Exit with the test result code
	os.Exit(code)
}

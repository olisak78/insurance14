//go:build integration
// +build integration

package repository

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"developer-portal-backend/internal/testutils"
)

// TestMain runs before all repository tests and ensures proper Docker cleanup
func TestMain(m *testing.M) {
	// Set up signal handling for graceful cleanup on interruption (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Run cleanup in a goroutine that listens for signals
	go func() {
		<-c
		log.Println("\n🛑 Repository tests interrupted, cleaning up Docker containers...")
		testutils.CleanupSharedContainer()
		os.Exit(1)
	}()

	// Run tests
	log.Println("🧪 Starting repository integration tests...")
	code := m.Run()

	// Always cleanup when tests finish normally
	log.Println("✅ Repository tests completed, cleaning up Docker containers...")
	testutils.CleanupSharedContainer()

	// Exit with the test result code
	os.Exit(code)
}

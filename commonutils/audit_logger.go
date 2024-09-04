package commonutils

import (
	"log"
	"os"
)

// Create a logger
var auditLogger *log.Logger

// Log a message to the audit log
func auditLog(message string) {
	auditLogger.Println(message)
}

func initAuditLogger() {
	file, err := os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open audit log file: %v", err)
	}

	auditLogger = log.New(file, "", log.LstdFlags)
}

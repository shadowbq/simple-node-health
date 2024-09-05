package audit

import (
	"log"
	"os"
)

// Create a logger
var AuditLogger *log.Logger

// Log a message to the audit log
func AuditLog(message string) {
	AuditLogger.Println(message)
}

func InitAuditLogger() {
	file, err := os.OpenFile("/var/log/snh.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open audit log file: %v", err)
	}

	AuditLogger = log.New(file, "", log.LstdFlags)
}

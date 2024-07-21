package smartapigo

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// CustomFormatter is a custom formatter for Logrus.
type CustomFormatter struct{}

// Format renders a single log entry.
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Format(time.RFC3339)
	level := strings.ToUpper(entry.Level.String())[:3]
	pc, file, line, ok := runtime.Caller(9) // Adjust the stack depth as needed
	if !ok {
		file = "???"
		line = 0
	} else {
		file = filepath.Base(file)
	}
	funcName := runtime.FuncForPC(pc).Name()
	funcName = filepath.Base(funcName)

	logMessage := fmt.Sprintf("%s | %s | %s | %s:%d | %s\n",
		timestamp, level, file, funcName, line, entry.Message)
	return []byte(logMessage), nil
}

package utillog

// #2025-02-12

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"runtime"
)

func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(2) // Adjust depth as needed
	if !ok {
		return ""
	}
	// funcName := "unknown"
	// if fn := runtime.FuncForPC(pc); fn != nil {
	// 	// Extract only the function name (without full package path)
	// 	parts := strings.Split(fn.Name(), "/")
	// 	funcName = parts[len(parts)-1] // Get last part
	// }
	// return fmt.Sprintf("%s:%d", funcName, line)
	dir := filepath.Base(filepath.Dir(file))
	if dir == "." {
		dir = "" // Avoid returning "./filename:line"
	}
	name := filepath.Base(file)
	return fmt.Sprintf("%s/%s:%d", dir, name, line)
}

// var DefaultWriter = bufio.NewWriterSize(os.Stdout, 4096*10) // need mutex

var DefaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	// Level: slog.LevelInfo, // Set the minimum log level to Warning
}))

// var DefaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
// 	// Level: slog.LevelInfo, // Set the minimum log level to Warning
// 	// AddSource: true,
// }))

// var DefaultLogger = log.New(os.Stdout, "", log.LUTC)

func Info(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Info(msg, slog.String("src", getCallerInfo()))
}

func Error(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Error(msg, slog.String("src", getCallerInfo()))

}

func Panic(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Error(msg, slog.String("src", getCallerInfo()))

	log.Panic(msg)

}
func Debug(format string, v ...any) {

	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Debug(msg, slog.String("src", getCallerInfo()))
}
func Warn(format string, v ...any) {

	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Warn(msg, slog.String("src", getCallerInfo()))
}
func Sync() {
	// if zap
	fmt.Print("log sync...")
}

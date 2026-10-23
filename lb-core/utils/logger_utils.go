package utils

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"

	logger "github.com/sirupsen/logrus"
)

func InitLogger() {
	logger.SetReportCaller(true)
	logger.SetFormatter(NewTextFormatter(true))
	logger.SetOutput(os.Stdout)
}

type textFormatter struct {
	needColor bool
}

func NewTextFormatter(needColor bool) logger.Formatter {
	return &textFormatter{
		needColor: needColor,
	}
}

func (formatter *textFormatter) Format(entry *logger.Entry) ([]byte, error) {
	buf := &bytes.Buffer{}

	// Время
	writeColorText(buf, entry.Time.Format("2006-01-02 15:04:05"), colorTime, formatter.needColor)
	buf.WriteByte('\t')

	// Уровень
	writeLogLevel(buf, strings.ToUpper(entry.Level.String()), formatter.needColor)
	buf.WriteByte('\t')

	// Точка вызова
	writeColorText(buf, getExecutionPoint(entry), colorFile, formatter.needColor)
	buf.WriteByte('\t')

	// Сообщение + ошибка
	hasError := entry.Data["error"] != nil
	writeColorText(
		buf,
		fmt.Sprintf(": %s%s", entry.Message, getErrorString(entry)),
		colorErrorLevel,
		formatter.needColor && hasError,
	)
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

const (
	colorTraceLevel   = "\033[1;34m"
	colorDebugLevel   = "\033[1;36m"
	colorInfoLevel    = "\033[1;32m"
	colorWarningLevel = "\033[1;33m"
	colorErrorLevel   = "\033[1;31m"

	colorTime    = "\033[2;37m" // Серый для времени
	colorService = "\033[2;37m" // Серый для сервиса
	colorFile    = "\033[94m"   // Синий для файла
	colorReset   = "\033[0m"    // Сброс цвета
)

func writeColorText(buf *bytes.Buffer, text string, colorString string, needColor bool) {
	if !needColor {
		buf.WriteString(text)
		return
	}

	buf.WriteString(colorString)
	buf.WriteString(text)
	buf.WriteString(colorReset)
}

func writeLogLevel(buf *bytes.Buffer, level string, needColor bool) {
	switch level {
	case "TRACE":
		writeColorText(buf, level, colorTraceLevel, needColor)
	case "DEBUG":
		writeColorText(buf, level, colorDebugLevel, needColor)
	case "INFO":
		writeColorText(buf, level, colorInfoLevel, needColor)
	case "WARNING":
		writeColorText(buf, level, colorWarningLevel, needColor)
	case "ERROR":
		writeColorText(buf, level, colorErrorLevel, needColor)
	default:
		buf.WriteString(level)
	}
}

func getExecutionPoint(entry *logger.Entry) string {
	executionPointFile := "unknown"
	executionPointLine := ""
	if entry.Caller != nil {
		_, fileName := SubstringLast(entry.Caller.File, "/")
		executionPointFile = fileName
		executionPointLine = fmt.Sprintf("%d", entry.Caller.Line)
	}

	if pad := 28 - len(executionPointFile) - len(executionPointLine); pad > 0 {
		for range pad {
			executionPointFile = string(' ') + executionPointFile
		}
	}

	return fmt.Sprintf("%s:%s", executionPointFile, executionPointLine)
}

func getErrorString(entry *logger.Entry) string {
	var result string

	if err := entry.Data["error"]; err != nil {
		printStack := false
		if needStack, ok := entry.Data["needStack"]; ok {
			if val, ok := needStack.(bool); ok {
				printStack = val
			}
		}

		if err, ok := err.(error); ok {
			if printStack {
				result = fmt.Sprintf(": %s\n%s", err.Error(), getStack())
			} else {
				result = fmt.Sprintf(": %s", err.Error())
			}
		}
		result = strings.TrimSuffix(result, "\n")
	}

	return result
}

func getStack() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			break
		}
		buf = make([]byte, 2*len(buf))
	}
	return string(buf)
}

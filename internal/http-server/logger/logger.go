package logger

import (
	"fmt"
	"os"
	"time"
)

const (
	DefaultLogPath = "./logs.txt"
)

type Logger struct {
	Title       string
	Message     string
	Timestamp   time.Time
	Location    string
	MessageType string
	FilePath    string
}

func New(title, location string, logError error) (*Logger, error) {
	messageText := ""

	if logError != nil {
		messageText = logError.Error()
	}

	f, err := os.OpenFile(DefaultLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("не удалось инициализировать файл логов: %w", err)
	}
	f.Close()

	return &Logger{
		Title:     title,
		Location:  location,
		Message:   messageText,
		Timestamp: time.Now(),
		FilePath:  "./logs.txt",
	}, nil
}

func (l *Logger) Write() {
	f, err := os.OpenFile(l.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	logLine := fmt.Sprintf("[%s] - [%s]: %s - '%s' :: %s\n",
		l.Timestamp.Format(time.RFC3339),
		l.MessageType,
		l.Title,
		l.Message,
		l.Location,
	)

	if _, err := f.WriteString(logLine); err != nil {
		return
	}

	return
}

func Error(title, location string, logErr error) {
	l, err := New(title, location, logErr)
	if err != nil {
		fmt.Printf("Ошибка логгера: %v\n", err)
		return
	}
	l.MessageType = "ERROR"
	l.Write()
}

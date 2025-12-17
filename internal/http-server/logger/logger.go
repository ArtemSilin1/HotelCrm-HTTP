package logger

import (
	"fmt"
	"os"
	"time"
)

const (
	LogMessageTypeError = "--> Ошибка"
)

type LogMessage struct {
	Title       string
	Message     string
	Timestamp   time.Time
	Location    string
	MessageType string
}

func NewLog(title, location string, logEror error) error {
	errorText := ""
	messageType := ""

	if logEror != nil {
		errorText = logEror.Error()
		messageType = LogMessageTypeError
	}

	log := &LogMessage{
		MessageType: messageType,
		Message:     errorText,
		Title:       title,
		Location:    location,
		Timestamp:   time.Now(),
	}

	f, err := os.OpenFile("./logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	logLine := fmt.Sprintf("[%s] - [%s]: %s - '%s' :: %s\n",
		log.Timestamp.Format(time.RFC3339), log.MessageType, log.Title, log.Message, log.Location)
	if _, err := f.WriteString(logLine); err != nil {
		return err
	}

	return nil
}

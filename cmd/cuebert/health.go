package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// StatusMessage is used by the program to send status messages
// to the health check endpoint
type StatusMessage struct {
	Message     string         `json:"message"`
	Code        int            `json:"code"`
	DB          *DBStatus      `json:"db"`
	Poll        *RoutineStatus `json:"poll"`
	Check       *RoutineStatus `json:"check"`
	Diff        *RoutineStatus `json:"diff"`
	Respond     *BotStatus     `json:"respond"`
	DailyReport *BotStatus     `json:"daily_repost"`
}

// BotStatus is used to send the status of various parts of the bot
type BotStatus struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Error   error  `json:"error"`
	Time    string `json:"time"`
}

// DBStatus is used to send the status of the database connection
type DBStatus struct {
	Connected bool `json:"connected"`
}

// RoutineStatus is used to send the status of the main program routines
type RoutineStatus struct {
	Name          string `json:"name"`
	Start         string `json:"start_time"`
	Finish        string `json:"finish_time"`
	FinishNoError bool   `json:"exit_no_error"`
	Message       string `json:"message"`
}

type routineUpdate struct {
	start      bool
	finish     bool
	err        bool
	routine    *RoutineStatus
	statusChan chan StatusMessage
}

// StatusHandler is used to manage the status of the program
type StatusHandler struct {
	status     StatusMessage
	statusLock sync.RWMutex
}

// SetStatus is used to set the status of the program retrieved by the health check endpoint
func (sh *StatusHandler) SetStatus(status StatusMessage) {
	sh.statusLock.Lock()
	defer sh.statusLock.Unlock()
	sh.status = status
}

// GetStatus is used by other parts of the program to retrieve the status and only update
// the status message as it pertains to that part of the program.
func (sh *StatusHandler) GetStatus() StatusMessage {
	sh.statusLock.RLock()
	defer sh.statusLock.RUnlock()
	return sh.status
}

// StartHealthHandler is used to start the health check endpoint
func (sh *StatusHandler) StartHealthHandler() {
	server := &http.Server{
		Addr:              ":8888",
		ReadHeaderTimeout: 3 * time.Second,
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Read the status from the handler
		status := sh.GetStatus()

		// Convert the status message to JSON
		jsonData, err := json.Marshal(status)
		if err != nil {
			log.Println("Error:", err)
			return
		}

		// Set the appropriate content type
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON data to the response writer
		_, err = io.WriteString(w, string(jsonData))
		if err != nil {
			log.Println("Error:", err)
			return
		}
	})
	log.Fatal(server.ListenAndServe())
}

func (sh *StatusHandler) readStatusMessages() StatusMessage {
	return sh.GetStatus()
}

func (sh *StatusHandler) sendStatusMessages(
	status StatusMessage,
	statusChannel chan StatusMessage) {
	for {
		sh.SetStatus(status)
		// Send the status message to the channel
		statusChannel <- status
	}
}

func (sh *StatusHandler) updateStatus(u *routineUpdate, routineType string) {
	status := sh.readStatusMessages()

	var routineStatus *RoutineStatus
	switch routineType {
	case "check":
		if status.Check == nil {
			status.Check = &RoutineStatus{}
		}
		routineStatus = status.Check
	case "diff":
		if status.Diff == nil {
			status.Diff = &RoutineStatus{}
		}
		routineStatus = status.Diff
	case "poll":
		if status.Poll == nil {
			status.Poll = &RoutineStatus{}
		}
		routineStatus = status.Poll
	default:
		// Handle invalid routine type
		return
	}

	routineStatus.Name = u.routine.Name
	routineStatus.Message = u.routine.Message

	if u.start {
		routineStatus.Start = u.routine.Start
	}

	if u.finish {
		routineStatus.Finish = u.routine.Finish
		routineStatus.FinishNoError = u.routine.FinishNoError
	}

	if u.err {
		routineStatus.FinishNoError = false
		routineStatus.Finish = u.routine.Finish
	}

	sh.sendStatusMessages(status, u.statusChan)
}

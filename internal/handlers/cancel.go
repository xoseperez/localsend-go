package handlers

import (
	"net/http"
	"sync"

	"github.com/meowrain/localsend-go/internal/utils/logger"
)

var (
	cancelHandlers = make(map[string]func())
	handlersLock   sync.RWMutex
)

// RegisterCancelHandler registers a cancel handler for a session
func RegisterCancelHandler(sessionID string, cancelFunc func()) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	cancelHandlers[sessionID] = cancelFunc
}

// UnregisterCancelHandler unregisters the cancel handler for a session
func UnregisterCancelHandler(sessionID string) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	delete(cancelHandlers, sessionID)
}

// HandleCancel handles a cancel request
func HandleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	logger.Debugf("Received cancel request for session: %s", sessionID)
	if sessionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handlersLock.RLock()
	cancelFunc, exists := cancelHandlers[sessionID]
	handlersLock.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cancelFunc()
	w.WriteHeader(http.StatusOK)
}

package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// handleVPNConnect handles VPN connection request
func (s *Server) handleVPNConnect(req *Request) *Response {
	var params VPNConnectParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	// Token and server must be provided - no persistence
	token := params.Token
	server := params.Server

	// Validate we have token and server
	if token == "" || server == "" {
		result := VPNConnectResult{Success: false, Message: "Token and server are required. Credentials are not saved - you must provide them on each connection."}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	// Generate session ID
	sessionID := fmt.Sprintf("conn-%d", time.Now().Unix())

	// Create context for this connection
	ctx, cancel := context.WithCancel(context.Background())

	conn := &VPNConnection{
		SessionID:  sessionID,
		Server:     server,
		StartTime:  time.Now(),
		CancelFunc: cancel,
		DoneChan:   make(chan struct{}),
	}

	s.connMutex.Lock()
	// Disconnect all existing connections before starting new one
	for existingSessionID, existingConn := range s.connections {
		fmt.Printf("Disconnecting existing connection: %s\n", existingSessionID)
		existingConn.CancelFunc()
		// Release lock temporarily to avoid deadlock while waiting
		s.connMutex.Unlock()
		// Wait for cleanup to complete with timeout
		select {
		case <-existingConn.DoneChan:
			fmt.Printf("Connection %s cleaned up successfully\n", existingSessionID)
		case <-time.After(10 * time.Second):
			fmt.Printf("Warning: Timeout waiting for connection %s cleanup\n", existingSessionID)
		}
		s.connMutex.Lock()
		delete(s.connections, existingSessionID)
	}
	// Add the new connection
	s.connections[sessionID] = conn
	s.connMutex.Unlock()

	// Start VPN connection in background with server and token
	go s.runVPNConnection(ctx, sessionID, server, token, conn.DoneChan)

	result := VPNConnectResult{
		Success:   true,
		Message:   "VPN connection started successfully",
		SessionID: sessionID,
	}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// handleVPNQuit handles VPN disconnection request
func (s *Server) handleVPNQuit(req *Request) *Response {
	// Find the active connection (we only support one connection at a time for now)
	s.connMutex.Lock()
	var sessionID string
	var conn *VPNConnection
	for id, c := range s.connections {
		sessionID = id
		conn = c
		break
	}
	s.connMutex.Unlock()

	if conn == nil {
		result := VPNQuitResult{Success: false, Message: "No active VPN connection"}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	// Cancel the connection
	if conn.CancelFunc != nil {
		conn.CancelFunc()
	}

	// Remove from connections map
	s.connMutex.Lock()
	delete(s.connections, sessionID)
	s.connMutex.Unlock()

	result := VPNQuitResult{Success: true, Message: "VPN connection stopped successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// handleStatus handles status request
func (s *Server) handleStatus(req *Request) *Response {
	s.connMutex.RLock()
	var connStatuses []ConnectionStatus
	for _, conn := range s.connections {
		connStatuses = append(connStatuses, ConnectionStatus{
			SessionID: conn.SessionID,
			Server:    conn.Server,
			Connected: true,
			Uptime:    int64(time.Since(conn.StartTime).Seconds()),
		})
	}
	s.connMutex.RUnlock()

	result := StatusResult{
		Running:     true,
		Connections: connStatuses,
	}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// handlePing handles ping request
func (s *Server) handlePing(req *Request) *Response {
	result := PingResult{Message: "pong"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

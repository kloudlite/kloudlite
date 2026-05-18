package daemon

import (
	"encoding/json"
)

// handleHostsAdd handles adding a host entry
func (s *Server) handleHostsAdd(req *Request) *Response {
	var params HostsAddParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	if err := s.hostsManager.Add(params.Hostname, params.IP, params.Comment); err != nil {
		result := HostsAddResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsAddResult{Success: true, Message: "Host entry added successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// handleHostsRemove handles removing a host entry
func (s *Server) handleHostsRemove(req *Request) *Response {
	var params HostsRemoveParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	if err := s.hostsManager.Remove(params.Hostname); err != nil {
		result := HostsRemoveResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsRemoveResult{Success: true, Message: "Host entry removed successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// handleHostsList handles listing all host entries
func (s *Server) handleHostsList(req *Request) *Response {
	entries, err := s.hostsManager.List()
	if err != nil {
		return NewErrorResponse(req.ID, ErrCodeInternal, "Failed to list hosts", err.Error())
	}

	// Convert to protocol entries
	var protoEntries []HostsEntry
	for _, entry := range entries {
		protoEntries = append(protoEntries, HostsEntry{
			IP:       entry.IP,
			Hostname: entry.Hostname,
			Comment:  entry.Comment,
		})
	}

	result := HostsListResult{Entries: protoEntries}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// handleHostsSync handles synchronizing hosts
func (s *Server) handleHostsSync(req *Request) *Response {
	if err := s.hostsManager.Sync(); err != nil {
		result := HostsSyncResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsSyncResult{Success: true, Message: "Hosts synchronized successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// handleHostsClean handles cleaning up hosts
func (s *Server) handleHostsClean(req *Request) *Response {
	if err := s.hostsManager.Clean(); err != nil {
		result := HostsCleanResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsCleanResult{Success: true, Message: "Hosts cleaned successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// handleHostsFlush handles flushing DNS cache
func (s *Server) handleHostsFlush(req *Request) *Response {
	if err := s.hostsManager.Flush(); err != nil {
		result := HostsFlushResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsFlushResult{Success: true, Message: "DNS cache flushed successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

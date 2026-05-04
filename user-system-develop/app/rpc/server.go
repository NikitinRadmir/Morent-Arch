package rpc

import (
	"encoding/json"
	"net/http"
)

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return e.Message
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type RPCServer struct {
	methods *RPCMethods
}

func NewRPCServer(methods *RPCMethods) *RPCServer {
	return &RPCServer{
		methods: methods,
	}
}

func (s *RPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, nil, -32700, "Parse error", nil)
		return
	}

	if req.JSONRPC != "2.0" {
		s.sendError(w, req.ID, -32600, "Invalid Request", nil)
		return
	}

	var result interface{}
	var err error

	switch req.Method {
	case "company.getHierarchy":
		result, err = s.methods.GetCompanyHierarchy(req.Params)
	case "user.getByRole":
		result, err = s.methods.GetUsersByRole(req.Params)
	case "role.getUsers":
		result, err = s.methods.GetRoleUsers(req.Params)
	case "permission.checkUserAccess":
		result, err = s.methods.CheckUserAccess(req.Params)
	default:
		s.sendError(w, req.ID, -32601, "Method not found", nil)
		return
	}

	if err != nil {
		// Теперь это работает, потому что RPCError implements error
		if rpcErr, ok := err.(*RPCError); ok {
			s.sendError(w, req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
		} else {
			s.sendError(w, req.ID, -32603, err.Error(), nil)
		}
		return
	}

	s.sendResponse(w, req.ID, result)
}

func (s *RPCServer) sendResponse(w http.ResponseWriter, id interface{}, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	})
}

func (s *RPCServer) sendError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	})
}

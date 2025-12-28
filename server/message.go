package server

type Message struct {
	JSONRPC string `json:"jsonrpc"`
}

type RequestMessage struct {
	Message
	ID     *int   `json:"id"`
	Method string `json:"method"`
	Params *any   `json:"params,omitempty"`
}

type ResponseMessage struct {
	Message
	ID     *int           `json:"id"`
	Result *any           `json:"result,omitempty"`
	Error  *ResponseError `json:"error,omitempty"`
}

type NotificationMessage struct {
	Message
	Method string `json:"method"`
	Params *any   `json:"params,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *any   `json:"data,omitempty"`
}

type ErrorCodes int

const (
	ParseError           ErrorCodes = -32700
	InvalidRequest       ErrorCodes = -32600
	MethodNotFound       ErrorCodes = -32601
	InvalidParams        ErrorCodes = -32602
	InternalError        ErrorCodes = -32603
	ServerErrorStart     ErrorCodes = -32099
	ServerErrorEnd       ErrorCodes = -32000
	ServerNotInitialized ErrorCodes = -32002
	UnknownErrorCode     ErrorCodes = -32001

	RequestCancelled ErrorCodes = -32800
	ContentModified  ErrorCodes = -32801
)

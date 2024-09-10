package contexter

import "context"

type ctxKey string

var (
	// RequestID is the context key for the request ID.
	RequestID = ctxKey("request_id")
)

// GetRequestID	returns the request id from the context.
func GetRequestID(ctx context.Context) string {
	requestID := ctx.Value(RequestID).(string)
	if requestID == "" {
		requestID = "unknown"
	}
	return requestID
}

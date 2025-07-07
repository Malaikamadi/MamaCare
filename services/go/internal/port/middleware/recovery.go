package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/middleware"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// Recovery middleware recovers from panics and logs the error
func Recovery(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					// Get request ID from context for correlation
					reqID := middleware.GetReqID(r.Context())
					
					// Create error message
					err := errorx.New(
						errorx.InternalServerError,
						fmt.Sprintf("panic: %v", rvr),
					)
					
					// Log the panic with stack trace
					log.Error().
						Str("request_id", reqID).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Str("panic", fmt.Sprintf("%v", rvr)).
						Str("stack", string(debug.Stack())).
						Msg("Panic recovered")
					
					// Return 500 Internal Server Error
					http.Error(
						w, 
						errorx.NewErrorResponse(err, reqID).String(),
						http.StatusInternalServerError,
					)
				}
			}()
			
			next.ServeHTTP(w, r)
		}
		
		return http.HandlerFunc(fn)
	}
}

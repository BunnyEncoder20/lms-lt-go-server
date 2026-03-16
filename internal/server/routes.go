package server

import (
	"log"
	"net/http"
	"os"

	"go-server/internal/models"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes here
	mux.HandleFunc("/", s.HelloWorldHandler)
	mux.HandleFunc("/dbhealth", s.healthHandler)

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			// 204 No Content: Modern browsers have a security feature: before they send a "risky" request (like a POST with JSON or a DELETE), they send a "test" request first to see if the server allows it.
			// This test req comes with the OPTIONS HTTP method.
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler (here, our router/mux from above)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Greeting the World")
	resp := models.JSONResponse{
		Success: true,
		Message: "Hello World",
		Data:    []int{1, 2, 3},
	}

	// Writing the reponse
	s.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	dbHealthMpp := s.db.Health()

	resp := models.JSONResponse{
		Success: true,
		Message: dbHealthMpp["message"],
		Data:    dbHealthMpp,
	}

	s.WriteJSON(w, http.StatusOK, resp)
}

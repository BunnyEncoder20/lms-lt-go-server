package server

import (
	"log"
	"net/http"
	"os"

	"go-server/internal/auth"
	"go-server/internal/middleware"
	"go-server/internal/models"
)

// Middleware is a custom type that makes our function signature cleaner
type Middleware func(http.Handler) http.Handler

// applyMiddleware is a helper function to apply multiple middlewares to a handler in a clean way
// It iterates backwards so that first middleware in the list is the outermost one (the first one to execute when a request comes in)
func applyMiddleware(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Module Init
	jwtSecret := os.Getenv("JWT_SECRET")
	authService := auth.NewService(s.db, jwtSecret)
	authHandler := auth.NewHandler(authService)

	// Middleware Stacks
	// grouping middlewares into slices makes applying them to routes incredibly easy and clean.
	// We can have different stacks for different types of routes (e.g., public vs protected)
	adminOnlyMiddlewares := []Middleware{
		middleware.RequireAuth,
		middleware.RequireRoles(string(models.RoleAdmin)),
	}

	managerOrAdminMiddlewares := []Middleware{
		middleware.RequireAuth,
		middleware.RequireRoles(string(models.RoleAdmin), string(models.RoleManager)),
	}

	// Register routes

	// --- Public Routes
	mux.HandleFunc("GET /", s.HelloWorldHandler)
	mux.HandleFunc("GET /dbhealth", s.healthHandler)
	mux.HandleFunc("POST /login", authHandler.HandleLogin)

	// --- Protected Routes
	mux.Handle("GET /admin/user/list", applyMiddleware(http.HandlerFunc(s.HelloWorldHandler), adminOnlyMiddlewares...))
	mux.Handle("GET /manager/user/list", applyMiddleware(http.HandlerFunc(s.HelloWorldHandler), managerOrAdminMiddlewares...))

	// Can add more modules routes here

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
		Data:    []string{"Welcome to the Go Server!", "This is a sample response.", "No you cannot build this thing even if you tried"},
	}

	// Writing the reponse
	models.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	dbHealthMpp := s.db.Health()

	resp := models.JSONResponse{
		Success: true,
		Message: dbHealthMpp["message"],
		Data:    dbHealthMpp,
	}

	models.WriteJSON(w, http.StatusOK, resp)
}

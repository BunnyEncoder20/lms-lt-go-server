package server

import (
	"log"
	"net/http"
	"os"

	"go-server/internal/admin"
	"go-server/internal/auth"
	"go-server/internal/middleware"
	"go-server/internal/models"
	"go-server/internal/nominations"
	"go-server/internal/trainings"
	"go-server/internal/users"
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
	jwtSecret := os.Getenv("JWT_SECRET")

	// Module Init
	authService := auth.NewService(s.db, jwtSecret, s.log)
	authHandler := auth.NewHandler(authService, s.log)

	usersService := users.NewService(s.db, s.log)
	usersHandler := users.NewHandler(usersService, s.log)

	trainingsService := trainings.NewService(s.db)
	trainingsHandler := trainings.NewHandler(trainingsService, s.log)

	nominationsService := nominations.NewService(s.db)
	nominationsHandler := nominations.NewHandler(nominationsService, s.log)

	adminService := admin.NewService(s.db)
	adminHandler := admin.NewHandler(adminService, s.log)

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
	mux.Handle("GET /me", middleware.RequireAuth(http.HandlerFunc(authHandler.HandleMe)))

	// User Management
	mux.Handle("GET /my-team", applyMiddleware(http.HandlerFunc(usersHandler.HandleGetMyTeam), managerOrAdminMiddlewares...))
	mux.Handle("GET /admin/users", applyMiddleware(http.HandlerFunc(usersHandler.HandleListUsers), adminOnlyMiddlewares...))
	mux.Handle("GET /admin/users/{id}", applyMiddleware(http.HandlerFunc(usersHandler.HandleGetUser), adminOnlyMiddlewares...))
	mux.Handle("POST /admin/users", applyMiddleware(http.HandlerFunc(usersHandler.HandleCreateUser), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/users/{id}/status", applyMiddleware(http.HandlerFunc(usersHandler.HandleUpdateUserStatus), adminOnlyMiddlewares...))
	mux.Handle("DELETE /admin/users/{id}", applyMiddleware(http.HandlerFunc(usersHandler.HandleSoftDeleteUser), adminOnlyMiddlewares...))
	mux.Handle("DELETE /admin/users/permanent/{id}", applyMiddleware(http.HandlerFunc(usersHandler.HandleDeleteUser), adminOnlyMiddlewares...))

	// Training Management
	// -- Public/Employee routes (accessible to authenticated users)
	mux.Handle("GET /trainings", middleware.RequireAuth(http.HandlerFunc(trainingsHandler.HandleListTraining)))
	mux.Handle("GET /trainings/{id}", middleware.RequireAuth(http.HandlerFunc(trainingsHandler.HandleGetTraining)))
	mux.Handle("GET /trainings/category/{category}", middleware.RequireAuth(http.HandlerFunc(trainingsHandler.HandleGetTrainingCategory)))
	mux.Handle("GET /trainings/upcoming", middleware.RequireAuth(http.HandlerFunc(trainingsHandler.HandleGetUpcomingTraining)))
	mux.Handle("GET /my-trainings", middleware.RequireAuth(http.HandlerFunc(trainingsHandler.HandleGetEmployeeTraining)))
	// -- Admin only routes
	mux.Handle("POST /admin/trainings", applyMiddleware(http.HandlerFunc(trainingsHandler.HandleCreateTraining), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/trainings/{id}", applyMiddleware(http.HandlerFunc(trainingsHandler.HandleUpdateTraining), adminOnlyMiddlewares...))
	mux.Handle("DELETE /admin/trainings/{id}", applyMiddleware(http.HandlerFunc(trainingsHandler.HandleDeleteTraining), adminOnlyMiddlewares...))

	// Nominations Management
	// -- Employee/Common routes (accessible to authenticated users)
	mux.Handle("GET /nominations/my", middleware.RequireAuth(http.HandlerFunc(nominationsHandler.HandleGetMyNominations)))
	mux.Handle("POST /nominations/self", middleware.RequireAuth(http.HandlerFunc(nominationsHandler.HandleSelfNomination)))
	mux.Handle("POST /nominations/{id}/respond", middleware.RequireAuth(http.HandlerFunc(nominationsHandler.HandleRespondToNomination)))
	mux.Handle("GET /dashboard/employee", middleware.RequireAuth(http.HandlerFunc(nominationsHandler.HandleGetEmployeeDashboard)))
	mux.Handle("GET /courses/published", middleware.RequireAuth(http.HandlerFunc(nominationsHandler.HandleGetAllPublishedCourses)))
	// -- Manager routes
	mux.Handle("POST /nominations", applyMiddleware(http.HandlerFunc(nominationsHandler.HandleNominateEmployees), managerOrAdminMiddlewares...))
	mux.Handle("GET /nominations/team", applyMiddleware(http.HandlerFunc(nominationsHandler.HandleGetTeamNominations), managerOrAdminMiddlewares...))
	mux.Handle("POST /nominations/{id}/manager-respond", applyMiddleware(http.HandlerFunc(nominationsHandler.HandleRespondToSelfNomination), managerOrAdminMiddlewares...))
	mux.Handle("GET /dashboard/manager", applyMiddleware(http.HandlerFunc(nominationsHandler.HandleGetManagerDashboard), managerOrAdminMiddlewares...))
	// -- Admin only routes
	mux.Handle("GET /admin/nominations", applyMiddleware(http.HandlerFunc(nominationsHandler.HandleGetAllNominations), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/nominations/{id}/status", applyMiddleware(http.HandlerFunc(nominationsHandler.HandleUpdateNominationStatus), adminOnlyMiddlewares...))

	// Admin Endpoints
	mux.Handle("GET /admin/kpis", applyMiddleware(http.HandlerFunc(adminHandler.HandleGetKpis), adminOnlyMiddlewares...))
	mux.Handle("GET /admin/monthly-stats", applyMiddleware(http.HandlerFunc(adminHandler.HandleGetMonthlyStats), adminOnlyMiddlewares...))
	mux.Handle("GET /admin/category-distribution", applyMiddleware(http.HandlerFunc(adminHandler.HandleGetCategoryDistribution), adminOnlyMiddlewares...))
	mux.Handle("GET /admin/cluster-stats", applyMiddleware(http.HandlerFunc(adminHandler.HandleGetClusterStats), adminOnlyMiddlewares...))
	mux.Handle("POST /admin/import-history", applyMiddleware(http.HandlerFunc(adminHandler.HandleImportHistory), adminOnlyMiddlewares...))

	// Global Middlewares - these apply to all routes and are added at the end to wrap everything
	globalMiddlewares := []Middleware{
		s.corsMiddleware,                // CORS middlware should be first to handle preflight requests and set headers
		middleware.RequestLogger(s.log), // Request logger for caputring all the requests and check route timings
	}
	return applyMiddleware(mux, globalMiddlewares...)
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

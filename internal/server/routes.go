package server

import (
	"log"
	"net/http"
	"os"
	"strings"

	"go-server/internal/admin"
	"go-server/internal/auth"
	"go-server/internal/courses"
	"go-server/internal/learning"
	"go-server/internal/middleware"
	"go-server/internal/models"
	"go-server/internal/nominations"
	"go-server/internal/notifications"
	"go-server/internal/trainings"
	"go-server/internal/users"
	"go-server/internal/utils"
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

	notificationsService := notifications.NewGatewayService(s.db, s.log)
	notificationsHandler := notifications.NewHandler(notificationsService, s.log)
	learningService := learning.NewService(s.db, notificationsService)
	learningHandler := learning.NewHandler(learningService, s.log)

	coursesService := courses.NewService(s.db)
	coursesHandler := courses.NewHandler(coursesService, s.log)

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
	mux.HandleFunc("POST /auth/login", authHandler.HandleLogin)
	mux.HandleFunc("POST /auth/refresh", authHandler.HandleRefresh)
	mux.HandleFunc("GET /uploads/{filename}", coursesHandler.HandleServeUpload)
	mux.Handle("GET /socket.io/", notifications.AsHTTPHandler(notificationsService))
	mux.Handle("POST /socket.io/", notifications.AsHTTPHandler(notificationsService))

	// --- Protected Routes
	mux.Handle("GET /users/me", middleware.RequireAuth(http.HandlerFunc(usersHandler.HandleGetMe)))
	mux.Handle("GET /notifications/feed", applyMiddleware(http.HandlerFunc(notificationsHandler.HandleGetHrFeed), adminOnlyMiddlewares...))

	// User Management
	mux.Handle("GET /my-team", applyMiddleware(http.HandlerFunc(usersHandler.HandleGetMyTeam), managerOrAdminMiddlewares...))
	mux.Handle("GET /admin/users", applyMiddleware(http.HandlerFunc(adminHandler.HandleGetUsers), adminOnlyMiddlewares...))
	mux.Handle("GET /admin/users/{id}", applyMiddleware(http.HandlerFunc(adminHandler.HandleGetUser), adminOnlyMiddlewares...))
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
	mux.Handle("GET /courses/published", middleware.RequireAuth(http.HandlerFunc(coursesHandler.HandleListPublishedCourses)))
	mux.Handle("GET /courses/published/{id}", middleware.RequireAuth(http.HandlerFunc(coursesHandler.HandleGetPublishedCourse)))
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

	// Courses Endpoints
	mux.Handle("GET /admin/courses/stats", applyMiddleware(http.HandlerFunc(coursesHandler.HandleGetDashboardStats), adminOnlyMiddlewares...))
	mux.Handle("GET /admin/courses", applyMiddleware(http.HandlerFunc(coursesHandler.HandleListCourses), adminOnlyMiddlewares...))
	mux.Handle("GET /admin/courses/{id}", applyMiddleware(http.HandlerFunc(coursesHandler.HandleGetCourse), adminOnlyMiddlewares...))
	mux.Handle("POST /admin/courses", applyMiddleware(http.HandlerFunc(coursesHandler.HandleCreateCourse), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/courses/{id}", applyMiddleware(http.HandlerFunc(coursesHandler.HandleUpdateCourse), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/courses/{id}/publish", applyMiddleware(http.HandlerFunc(coursesHandler.HandlePublishCourse), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/courses/{id}/archive", applyMiddleware(http.HandlerFunc(coursesHandler.HandleArchiveCourse), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/courses/{id}/restore", applyMiddleware(http.HandlerFunc(coursesHandler.HandleRestoreCourse), adminOnlyMiddlewares...))
	mux.Handle("DELETE /admin/courses/{id}", applyMiddleware(http.HandlerFunc(coursesHandler.HandleDeleteCourse), adminOnlyMiddlewares...))
	mux.Handle("POST /admin/courses/{courseId}/modules", applyMiddleware(http.HandlerFunc(coursesHandler.HandleCreateModule), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/modules/{id}", applyMiddleware(http.HandlerFunc(coursesHandler.HandleUpdateModule), adminOnlyMiddlewares...))
	mux.Handle("DELETE /admin/modules/{id}", applyMiddleware(http.HandlerFunc(coursesHandler.HandleDeleteModule), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/modules/reorder", applyMiddleware(http.HandlerFunc(coursesHandler.HandleReorderModules), adminOnlyMiddlewares...))
	mux.Handle("POST /admin/modules/{moduleId}/lessons", applyMiddleware(http.HandlerFunc(coursesHandler.HandleCreateLesson), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/lessons/{id}", applyMiddleware(http.HandlerFunc(coursesHandler.HandleUpdateLesson), adminOnlyMiddlewares...))
	mux.Handle("DELETE /admin/lessons/{id}", applyMiddleware(http.HandlerFunc(coursesHandler.HandleDeleteLesson), adminOnlyMiddlewares...))
	mux.Handle("PATCH /admin/lessons/reorder", applyMiddleware(http.HandlerFunc(coursesHandler.HandleReorderLessons), adminOnlyMiddlewares...))
	mux.Handle("POST /admin/upload", applyMiddleware(http.HandlerFunc(coursesHandler.HandleUploadFile), adminOnlyMiddlewares...))

	// Learning Endpoints
	mux.Handle("GET /learning/admin/courses", applyMiddleware(http.HandlerFunc(learningHandler.HandleGetPublishedCourses), adminOnlyMiddlewares...))
	mux.Handle("POST /learning/admin/assignments/bulk", applyMiddleware(http.HandlerFunc(learningHandler.HandleBulkAssign), adminOnlyMiddlewares...))
	mux.Handle("GET /learning/admin/assignments", applyMiddleware(http.HandlerFunc(learningHandler.HandleGetAllAssignments), adminOnlyMiddlewares...))
	mux.Handle("GET /learning/my", middleware.RequireAuth(http.HandlerFunc(learningHandler.HandleGetMyAssignments)))
	mux.Handle("GET /learning/courses/{courseId}", middleware.RequireAuth(http.HandlerFunc(learningHandler.HandleGetCoursePlayer)))
	mux.Handle("GET /learning/courses/{courseId}/navigation", middleware.RequireAuth(http.HandlerFunc(learningHandler.HandleGetCourseNavigation)))
	mux.Handle("GET /learning/lessons/{lessonId}/next", middleware.RequireAuth(http.HandlerFunc(learningHandler.HandleGetNextLesson)))
	mux.Handle("GET /learning/lessons/{lessonId}/previous", middleware.RequireAuth(http.HandlerFunc(learningHandler.HandleGetPreviousLesson)))
	mux.Handle("PATCH /learning/progress/{lessonId}", middleware.RequireAuth(http.HandlerFunc(learningHandler.HandleUpdateProgress)))
	mux.Handle("POST /learning/progress/heartbeat/{lessonId}", middleware.RequireAuth(http.HandlerFunc(learningHandler.HandleHeartbeatProgress)))
	mux.Handle("POST /learning/progress/complete/{lessonId}", middleware.RequireAuth(http.HandlerFunc(learningHandler.HandleCompleteLesson)))

	// Global Middlewares - these apply to all routes and are added at the end to wrap everything
	globalMiddlewares := []Middleware{
		s.corsMiddleware,                // CORS middlware should be first to handle preflight requests and set headers
		middleware.RequestLogger(s.log), // Request logger for caputring all the requests and check route timings
	}
	return applyMiddleware(mux, globalMiddlewares...)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOriginCfg := os.Getenv("ALLOWED_ORIGIN")
		isAllowedOrigin := isOriginAllowed(origin, allowedOriginCfg)

		// Set CORS headers
		if isAllowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Add("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			if origin != "" && !isAllowedOrigin {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			// 204 No Content: Modern browsers have a security feature: before they send a "risky" request (like a POST with JSON or a DELETE), they send a "test" request first to see if the server allows it.
			// This test req comes with the OPTIONS HTTP method.
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler (here, our router/mux from above)
		next.ServeHTTP(w, r)
	})
}

func isOriginAllowed(requestOrigin, allowedOriginCfg string) bool {
	if requestOrigin == "" {
		return false
	}

	req := normalizeOrigin(requestOrigin)
	for allowed := range strings.SplitSeq(allowedOriginCfg, ",") {
		normalized := normalizeOrigin(allowed)
		if normalized == "" {
			continue
		}
		if normalized == "*" || normalized == req {
			return true
		}
	}

	return false
}

func normalizeOrigin(origin string) string {
	trimmed := strings.TrimSpace(origin)
	return strings.TrimRight(trimmed, "/")
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Greeting the World")
	resp := models.JSONResponse{
		Success: true,
		Message: "Hello World",
		Data:    []string{"Welcome to the Go Server!", "This is a sample response.", "No you cannot build this thing even if you tried"},
	}

	// Writing the reponse
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	dbHealthMpp := s.db.Health()

	if dbHealthMpp["status"] == "down" {
		utils.WriteJSON(w, http.StatusServiceUnavailable, models.JSONResponse{
			Success: false,
			Message: dbHealthMpp["Database is unavailable"],
			Data:    dbHealthMpp,
		})
	}

	utils.WriteJSON(w, http.StatusOK, models.JSONResponse{
		Success: true,
		Message: dbHealthMpp["message"],
		Data:    dbHealthMpp,
	})
}

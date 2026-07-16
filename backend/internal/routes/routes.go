package routes

import (
	"net/http"
	"time"

	"tst-isp/internal/handlers"
	"tst-isp/internal/middleware"
)

func Register() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", handlers.Health)
	mux.Handle("/api/auth/login", middleware.RateLimit(5, time.Minute)(middleware.ValidateLogin(http.HandlerFunc(handlers.Login))))
	mux.Handle("/api/auth/forgot-password", middleware.RateLimit(5, time.Hour)(http.HandlerFunc(handlers.ForgotPassword)))
	mux.Handle("/api/auth/reset-password", middleware.RateLimit(5, time.Hour)(http.HandlerFunc(handlers.ResetPassword)))
	mux.Handle("/api/auth/admin/register", middleware.RateLimit(3, time.Hour)(http.HandlerFunc(handlers.AdminRegister)))
	mux.HandleFunc("/api/mpesa/stkpush", handlers.MpesaSTKPush)
	mux.HandleFunc("/api/mpesa/callback", handlers.MpesaCallback)
	mux.HandleFunc("/api/subisp/register", handlers.SubIspRegister)

	mux.Handle("/api/admin/users", middleware.RequireRole("super_admin")(http.HandlerFunc(handlers.AdminUsers)))
	mux.Handle("/api/admin/admins", middleware.RequireRole("super_admin")(http.HandlerFunc(handlers.AdminCreateAdmin)))
	mux.Handle("/api/admin/revenue", middleware.RequireRole("super_admin")(http.HandlerFunc(handlers.AdminRevenue)))
	mux.Handle("/api/admin/routers", middleware.RequireRole("super_admin")(http.HandlerFunc(handlers.AdminRouters)))
	mux.Handle("/api/admin/routers/test", middleware.RequireRole("super_admin")(http.HandlerFunc(handlers.AdminRouters)))
	mux.Handle("/api/admin/usage", middleware.RequireRole("super_admin")(http.HandlerFunc(handlers.AdminUsage)))
	mux.Handle("/api/admin/subisps", middleware.RequireRole("super_admin")(http.HandlerFunc(handlers.AdminSubIsps)))
	mux.Handle("/api/admin/subisps/update", middleware.RequireRole("super_admin")(http.HandlerFunc(handlers.AdminUpdateSubIsp)))

	mux.Handle("/api/subisp/me", middleware.RequireRole("sub_isp")(http.HandlerFunc(handlers.SubIspMe)))
	mux.Handle("/api/subisp/routers", middleware.RequireRole("sub_isp")(http.HandlerFunc(handlers.SubIspRouters)))
	mux.Handle("/api/subisp/usage", middleware.RequireRole("sub_isp")(http.HandlerFunc(handlers.SubIspUsage)))

	return middleware.CORS(mux)
}

// internal/wire.go
// +build wireinject

package internal

// import (
//     "github.com/google/wire"
//     "github.com/go-chi/chi/v5"
//     "github.com/go-chi/chi/v5/middleware"
//     "getnoti.com/config"
//     notificationroutes "getnoti.com/internal/notifications/infra/http"
 
//     "getnoti.com/pkg/db"
//     "getnoti.com/pkg/logger"
// )

// func InitializeRouter(cfg *config.Config) (*chi.Mux, error) {
//     wire.Build(
//         db.NewDatabaseFactory,
//         logger.New,
//         ProvideNotificationRouter,
//         ProvideTemplateRouter,
//         NewMainRouter,
//     )
//     return &chi.Mux{}, nil
// }





// func NewMainRouter(
//     cfg *config.Config,
//     database db.Database,
//     log *logger.Logger,
//     notificationRouter *chi.Mux,
//     templateRouter *chi.Mux,
// ) *chi.Mux {
//     router := chi.NewRouter()

//     // Use common middleware
//     router.Use(middleware.RequestID)
//     router.Use(middleware.RealIP)
//     router.Use(middleware.Logger)
//     router.Use(middleware.Recoverer)

//     // Mount domain routers
//     router.Mount("/notifications", notificationRouter)
//     router.Mount("/templates", templateRouter)

//     return router
// }

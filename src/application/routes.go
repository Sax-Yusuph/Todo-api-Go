package application

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sax-yusuph/todo/handlers"
)

func (a *App) loadRoutes() {

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	c := cors.AllowAll()
	router.Use(c.Handler)

	auth := handlers.Auth{
		Repo:      a.rdb,
		JWTSecret: a.Config.JWTSecret,
	}

	todo := handlers.Todo{
		Client: a.rdb,
	}

	router.Post("/register", auth.Register)
	router.Post("/login", auth.Login)
	router.Post("/logout", auth.Logout)

	router.With(auth.Middleware).Post("/users/{userID}/todos", todo.Create)
	router.With(auth.Middleware).Get("/users/{userID}/todos", todo.List)
	router.With(auth.Middleware).Get("/users/{userID}/todos/{todoID}", todo.GetById)
	router.With(auth.Middleware).Put("/users/{userID}/todos/{todoID}", todo.UpdateById)
	router.With(auth.Middleware).Delete("/users/{userID}/todos/{todoID}", todo.DeleteById)

	a.Router = router
}

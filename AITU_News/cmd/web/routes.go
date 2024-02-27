package main

import (
	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	dynamicMiddleware := alice.New(app.session.Enable)

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/article/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createArticleForm))
	mux.Post("/article/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createArticle))
	mux.Get("/article/:id", dynamicMiddleware.ThenFunc(app.showArticle))
	mux.Get("/category/:category", dynamicMiddleware.ThenFunc(app.showCategoryArticles))
	mux.Get("/contacts", dynamicMiddleware.ThenFunc(app.contacts))

	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser))
	mux.Get("/admin/approve", dynamicMiddleware.Append(app.requireAdminAuthentication).ThenFunc(app.adminApproval))

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}

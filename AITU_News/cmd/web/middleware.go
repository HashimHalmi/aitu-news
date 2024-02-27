package main

import (
	"fmt"
	"net/http"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(rw, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)

	})
}

// isAdminAuthenticated checks if the user is authenticated as an admin.
func (app *application) isAdminAuthenticated(r *http.Request) bool {
	// Get the authenticated user's role from the session
	role := app.session.GetString(r, "authenticatedUserRole")
	if role != "Admin" {
		return false
	}
	return true
}

func (app *application) requireAdminAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the application instance is nil
		if app == nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Check if the user is authenticated
		userID := app.session.GetInt(r, "authenticatedUserID")
		if userID == 0 {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Fetch the user details from the database using the user ID
		user, err := app.users.Get(userID)
		if err != nil {
			// Handle error
			app.serverError(w, err)
			return
		}

		// Check if the user is an admin
		if user.Role != "Admin" {
			// If not an admin, redirect to a different page (e.g., home page)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// If user is authenticated and is an admin, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

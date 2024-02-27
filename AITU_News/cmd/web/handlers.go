package main

import (
	"errors"
	"fmt"
	"hamedfrogh.net/aitunews/pkg/forms"
	"hamedfrogh.net/aitunews/pkg/models"
	"net/http"
	"strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s, err := app.articles.Latest(ctx)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Retrieve the list of categories
	categories, err := app.articles.GetCategories()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Articles:   s,
		Categories: categories,
	})
}

func (app *application) showArticle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	s, err := app.articles.Get(id)

	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, r, "show.page.tmpl", &templateData{
		Article: s,
	})
}

func (app *application) createArticleForm(w http.ResponseWriter, r *http.Request) {

	// Check if the user is authenticated
	if !app.isAuthenticated(r) {
		// If not authenticated, redirect to the login page
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Get the authenticated user's role
	userID := app.session.GetInt(r, "authenticatedUserID")

	// Check if the user's role is teacher and if they are approved
	approved, err := app.users.IsApproved(userID)
	if err != nil {
		// Handle error fetching user approval status
		app.serverError(w, err)
		return
	}

	// If the user is not approved, redirect to the home page

	// Fetch user's role from the databaseI
	role, err := app.users.GetRoleByID(userID)
	if err != nil {
		// Handle error fetching user role
		app.serverError(w, err)
		return
	}

	// Check if the user's role is student
	if role == "Student" {
		// If user's role is student, redirect to the home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if role == "Teacher" && !approved {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Debugging: Print the retrieved role
	fmt.Println("User Role:", role)

	// Check if the user's role is admin or teacher
	if role != "Admin" && role != "Teacher" {
		// If user's role is not admin or teacher, display an error message or redirect to another page
		app.clientError(w, http.StatusForbidden)
		return
	}

	app.render(w, r, "create.page.tmpl", &templateData{
		Form: forms.New(nil),
	})

}

func (app *application) createArticle(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires", "category")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	id, err := app.articles.Insert(form.Get("title"), form.Get("content"), form.Get("expires"), form.Get("category"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Article successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/article/%d", id), http.StatusSeeOther)
}

func (app *application) showCategoryArticles(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get(":category")

	articles, err := app.articles.GetByCategory(category)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "category.page.tmpl", &templateData{
		Category: category,
		Articles: articles,
	})
}

func (app *application) contacts(w http.ResponseWriter, r *http.Request) {

	app.render(w, r, "contacts.page.tmpl", &templateData{})
}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	// Parse the form data.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// Validate the form contents using the form helper we made earlier.
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password", "role")
	form.MaxLength("name", 255)
	form.MaxLength("email", 255)
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 4)

	// If there are any errors, redisplay the signup form.
	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"), form.Get("role"))
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.Errors.Add("email", "Address is already in use")
			app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.session.Put(r, "flash", "Your signup was successful. Please log in.")
	// Otherwise send a placeholder response (for now!).
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("generic", "Email or Password is incorrect")
			app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.session.Put(r, "authenticatedUserID", id)
	http.Redirect(w, r, "/article/create", http.StatusSeeOther)

}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, "authenticatedUserID")
	app.session.Put(r, "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) adminApproval(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authenticated as admin
	if !app.isAdminAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Fetch a list of teacher users pending approval
	pendingTeachers, err := app.users.GetPendingTeachers()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Process approval requests
	if r.Method == http.MethodPost {
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		// Retrieve the list of selected teacher user IDs from the form
		approvedTeacherIDs := r.PostForm["approved"]

		// Set the approval status for each selected teacher user
		for _, teacherIDStr := range approvedTeacherIDs {
			teacherID, err := strconv.Atoi(teacherIDStr)
			if err != nil {
				// Handle error
				continue
			}

			// Set the user as approved
			err = app.users.SetApprovalStatus(teacherID, true)
			if err != nil {
				// Handle error
				continue
			}
		}

		// Redirect to the admin approval page to refresh the list
		http.Redirect(w, r, "/admin/approve", http.StatusSeeOther)
		return
	}

	// Render the admin approval interface template
	app.render(w, r, "admin_approval.page.tmpl", &templateData{
		PendingTeachers: pendingTeachers,
	})
}

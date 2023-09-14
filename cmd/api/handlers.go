package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/1jack80/guardian"
	"github.com/1jack80/todo-api/models"
	"github.com/go-sql-driver/mysql"
)

func (a *api) signupHandler(w http.ResponseWriter, r *http.Request) {
	user := a.models.User.New()

	err := readJsonFromReq(r, &user)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, "internal server error")
		a.errLog.Println("unable to process json from request Body")
		return
	}
	err = a.models.User.Create(user)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateUsername) {
			jsonResponse(w, http.StatusBadRequest, "username already exists")
			return
		}
		jsonResponse(w, http.StatusInternalServerError, "internal server error")
		a.errLog.Println(err)
		return
	}

	jsonResponse(w, (http.StatusOK), "user created successfully")
}

func (a *api) loginHandler(w http.ResponseWriter, r *http.Request) {
	userCreds := a.models.User.New()
	readJsonFromReq(r, &userCreds)

	/*
			get usr details from request
		 	validate user details
				if invlid:
					return err
				else if valid:
					add create new session
					respond with cookie
	*/

	user, err := a.models.User.GetByUsername(userCreds.Username)
	if err != nil {
		a.errLog.Println(err)

		var mySQLError *mysql.MySQLError
		if errors.Is(err, sql.ErrNoRows) {
			jsonResponse(w, http.StatusUnauthorized, "user not found")
			return
		}
		if errors.As(err, &mySQLError) || strings.Contains(err.Error(), "sql") {
			jsonResponse(w, http.StatusInternalServerError, "internal server error")
			return
		}
		jsonResponse(w, http.StatusUnauthorized, "something went wrong")
		return
	}

	if !comparePasswordAndHash(userCreds.Password, user.Password) {
		jsonResponse(w, http.StatusUnauthorized, "incorrect password")
		return
	}

	newSession := a.sessions.CreateSession()
	newSession.Data["username"] = user.Username
	newSession.Data["userID"] = user.ID

	a.sessions.UpdateSession(newSession.ID, newSession)

	cookie, err := a.sessions.CreateCookie(newSession.ID)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, "Something went wrong")
	}
	http.SetCookie(w, &cookie)
	jsonResponse(w, http.StatusOK, "Login successful")
}

func (a *api) logoutHandler(w http.ResponseWriter, r *http.Request) {
	ctxVal := r.Context().Value(a.sessions.ContextKey())

	session, ok := ctxVal.(guardian.Session)
	if !ok {
		jsonResponse(w, http.StatusInternalServerError, "")
		return
	}

	err := a.sessions.InvalidateSession(session.ID)
	if !ok {
		a.errLog.Println(err)
		jsonResponse(w, http.StatusInternalServerError, "")
		return
	}

	cookie, err := a.sessions.CreateCookie(session.ID)
	if err != nil {
		a.errLog.Println(err)
		jsonResponse(w, http.StatusInternalServerError, "")
		return
	}
	cookie.Value = ""
	http.SetCookie(w, &cookie)
	jsonResponse(w, http.StatusOK, "logout successful")

}
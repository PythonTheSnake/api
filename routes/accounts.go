package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/zenazn/goji/web"

	"github.com/lavab/api/db"
	"github.com/lavab/api/dbutils"
	"github.com/lavab/api/env"
	"github.com/lavab/api/models"
	"github.com/lavab/api/utils"
)

// Accounts list
type AccountsListResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func AccountsList(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, 501, &AccountsListResponse{
		Success: false,
		Message: "Method not implemented",
	})
}

// Account registration
type AccountsCreateRequest struct {
	Username string `json:"username" schema:"username"`
	Password string `json:"password" schema:"password"`
}

type AccountsCreateResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	User    *models.User `json:"data,omitempty"`
}

func AccountsCreate(w http.ResponseWriter, r *http.Request) {
	// Decode the request
	var input AccountsCreateRequest
	err := utils.ParseRequest(r, input)
	if err != nil {
		env.G.Log.WithFields(logrus.Fields{
			"error": err,
		}).Warning("Unable to decode a request")

		utils.JSONResponse(w, 409, &AccountsCreateResponse{
			Success: false,
			Message: "Invalid input format",
		})
		return
	}

	// Ensure that the user with requested username doesn't exist
	if _, ok := dbutils.FindUserByName(username); ok {
		utils.JSONResponse(w, 409, &AccountsCreateResponse{
			Success: false,
			Message: "Username already exists",
		})
		return
	}

	// Try to hash the password
	hash, err := utils.BcryptHash(password)
	if err != nil {
		utils.JSONResponse(w, 500, &AccountsCreateResponse{
			Success: false,
			Message: "Internal server error - AC/CR/01",
		})

		env.G.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Unable to hash a password")
		return
	}

	// TODO: sanitize user name (i.e. remove caps, periods)

	// Create a new user object
	user := &models.User{
		Resource: base.MakeResource(utils.UUID(), username),
		Password: string(hash),
	}

	// Try to save it in the database
	if err := db.Insert("users", user); err != nil {
		utils.JSONResponse(w, 500, &AccountsCreateResponse{
			Success: false,
			Message: "Internal server error - AC/CR/02",
		})

		env.G.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Could not insert an user to the database")
		return
	}

	utils.JSONResponse(w, 201, &AccountsCreateResponse{
		Success: true,
		Message: "A new account was successfully created",
		User:    user,
	})
}

// AccountsGet returns the information about the specified account
type AccountsGetResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	User    *models.User `json:"user,omitempty"`
}

func AccountsGet(c *web.C, w http.ResponseWriter, r *http.Request) {
	// Get the account ID from the request
	id, ok := c.URLParams["id"]
	if !ok {
		utils.JSONResponse(409, &AccountsGetResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	// Right now we only support "me" as the ID
	if id != "me" {
		utils.JSONResponse(501, &AccountsGetResponse{
			Success: false,
			Message: `Only the "me" user is implemented`,
		})
		return
	}

	// Fetch the current session from the database
	session := models.CurrentSession(r)

	// Fetch the user object from the database
	user, ok := dbutils.GetUser(session.UserID)
	if !ok {
		// The session refers to a non-existing user
		env.G.Log.WithFields(logrus.Fields{
			"id": session.ID,
		}).Warning("Valid session referred to a removed account")

		// Try to remove the orphaned session
		if err := db.Delete("sessions", session.ID); err != nil {
			env.G.Log.WithFields(logrus.Fields{
				"id":    session.ID,
				"error": err,
			}).Error("Unable to remove an orphaned session")
		} else {
			env.G.Log.WithFields(logrus.Fields{
				"id": session.ID,
			}).Info("Removed an orphaned session")
		}

		utils.JSONResponse(410, &AccountsGetResponse{
			Success: false,
			Message: "Account disabled",
		})
		return
	}

	// Return the user struct
	utils.JSONResponse(200, &AccountsGetResponse{
		Success: true,
		User:    user,
	})
}

// AccountsUpdate TODO
type AccountsUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func AccountsUpdate(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(501, &AccountsUpdateResponse{
		Success: false,
		Message: `Sorry, not implemented yet`,
	})
}

// AccountsDelete TODO
type AccountsDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func AccountsDelete(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(501, &AccountsDeleteResponse{
		Success: false,
		Message: `Sorry, not implemented yet`,
	})
}

// AccountsWipeData TODO
type AccountsWipeDataResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func AccountsWipeData(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(501, &AccountsWipeDataResponse{
		Success: false,
		Message: `Sorry, not implemented yet`,
	})
}

// AccountsSessionsList TODO
type AccountsSessionsListResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func AccountsSessionsList(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(501, &AccountsSessionsListResponse{
		Success: false,
		Message: `Sorry, not implemented yet`,
	})
}

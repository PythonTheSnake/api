package routes

import (
	"net/http"

	"github.com/lavab/api/env"
	"github.com/lavab/api/utils"
)

// KeysListResponse contains the result of the KeysList request
type KeysListResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
	Keys    *[]string `json:"keys,omitempty"`
}

// KeysList responds with the list of keys assigned to the spiecified email
func KeysList(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		utils.JSONResponse(w, 409, &KeysListResponse{
			Success: false,
			Message: "Invalid username",
		})
		return
	}

	keys, err := env.Keys.FindByName(user)
	if err != nil {
		utils.JSONResponse(w, 500, &KeysListResponse{
			Success: false,
			Message: "Internal server error (KE/LI/01)",
		})
		return
	}

	keyIDs := []string{}
	for _, key := range keys {
		keyIDs = append(keyIDs, key.ID)
	}

	utils.JSONResponse(w, 200, &KeysListResponse{
		Success: true,
		Keys:    &keyIDs,
	})
}

// KeysCreateResponse contains the result of the KeysCreate request.
type KeysCreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// KeysCreate does *something* - TODO
func KeysCreate(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, 501, &KeysCreateResponse{
		Success: false,
		Message: "Sorry, not implemented yet",
	})
}

// KeysGetResponse contains the result of the KeysGet request.
type KeysGetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// KeysGet does *something* - TODO
func KeysGet(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, 501, &KeysGetResponse{
		Success: false,
		Message: "Sorry, not implemented yet",
	})
}

// KeysVoteResponse contains the result of the KeysVote request.
type KeysVoteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// KeysVote does *something* - TODO
func KeysVote(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, 501, &KeysVoteResponse{
		Success: false,
		Message: "Sorry, not implemented yet",
	})
}

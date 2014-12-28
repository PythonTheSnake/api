package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"github.com/zenazn/goji/web"

	"github.com/lavab/api/env"
	"github.com/lavab/api/models"
	"github.com/lavab/api/utils"
)

var (
	msgpackCodec codec.MsgpackHandle
)

// EmailsListResponse contains the result of the EmailsList request.
type EmailsListResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message,omitempty"`
	Emails  *[]*models.Email `json:"emails,omitempty"`
}

// EmailsList sends a list of the emails in the inbox.
func EmailsList(c web.C, w http.ResponseWriter, r *http.Request) {
	// Fetch the current session from the database
	session := c.Env["token"].(*models.Token)

	// Parse the query
	var (
		query     = r.URL.Query()
		sortRaw   = query.Get("sort")
		offsetRaw = query.Get("offset")
		limitRaw  = query.Get("limit")
		sort      []string
		offset    int
		limit     int
	)

	if offsetRaw != "" {
		o, err := strconv.Atoi(offsetRaw)
		if err != nil {
			env.Log.WithFields(logrus.Fields{
				"error":  err,
				"offset": offset,
			}).Error("Invalid offset")

			utils.JSONResponse(w, 400, &EmailsListResponse{
				Success: false,
				Message: "Invalid offset",
			})
			return
		}
		offset = o
	}

	if limitRaw != "" {
		l, err := strconv.Atoi(limitRaw)
		if err != nil {
			env.Log.WithFields(logrus.Fields{
				"error": err,
				"limit": limit,
			}).Error("Invalid limit")

			utils.JSONResponse(w, 400, &EmailsListResponse{
				Success: false,
				Message: "Invalid limit",
			})
			return
		}
		limit = l
	}

	if sortRaw != "" {
		sort = strings.Split(sortRaw, ",")
	}

	// Get contacts from the database
	emails, err := env.Emails.List(session.Owner, sort, offset, limit)
	if err != nil {
		env.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Unable to fetch emails")

		utils.JSONResponse(w, 500, &EmailsListResponse{
			Success: false,
			Message: "Internal error (code EM/LI/01)",
		})
		return
	}

	if offsetRaw != "" || limitRaw != "" {
		count, err := env.Emails.CountOwnedBy(session.Owner)
		if err != nil {
			env.Log.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unable to count emails")

			utils.JSONResponse(w, 500, &EmailsListResponse{
				Success: false,
				Message: "Internal error (code EM/LI/02)",
			})
			return
		}
		w.Header().Set("X-Total-Count", strconv.Itoa(count))
	}

	utils.JSONResponse(w, 200, &EmailsListResponse{
		Success: true,
		Emails:  &emails,
	})

	// GET parameters:
	//   sort - split by commas, prefixes: - is desc, + is asc
	//   offset, limit - for pagination
	// Pagination ADDS X-Total-Count to the response!
}

type EmailsCreateRequest struct {
	To              []string `json:"to"`
	BCC             []string `json:"bcc"`
	ReplyTo         string   `json:"reply_to"`
	ThreadID        string   `json:"thread_id"`
	Title           string   `json:"title"`
	Body            string   `json:"body"`
	Preview         string   `json:"preview"`
	Attachments     []string `json:"attachments"`
	PGPFingerprints []string `json:"pgp_fingerprints"`
}

// EmailsCreateResponse contains the result of the EmailsCreate request.
type EmailsCreateResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Created []string `json:"created,omitempty"`
}

// EmailsCreate sends a new email
func EmailsCreate(c web.C, w http.ResponseWriter, r *http.Request) {
	// Decode the request
	var input EmailsCreateRequest
	err := utils.ParseRequest(r, &input)
	if err != nil {
		env.Log.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Unable to decode a request")

		utils.JSONResponse(w, 400, &EmailsCreateResponse{
			Success: false,
			Message: "Invalid input format",
		})
		return
	}

	// Fetch the current session from the middleware
	session := c.Env["token"].(*models.Token)

	// Ensure that the input data isn't empty
	if len(input.To) == 0 || input.Title == "" || input.Body == "" {
		utils.JSONResponse(w, 400, &EmailsCreateResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	// Create a new email struct
	email := &models.Email{
		Resource:      models.MakeResource(session.Owner, input.Title),
		AttachmentIDs: input.Attachments,
		Body: models.Encrypted{
			Encoding:        "json",
			PGPFingerprints: input.PGPFingerprints,
			Data:            input.Body,
			Schema:          "email_body",
			VersionMajor:    1,
			VersionMinor:    0,
		},
		Preview: models.Encrypted{
			Encoding:        "json",
			PGPFingerprints: input.PGPFingerprints,
			Data:            input.Preview,
			Schema:          "email_preview",
			VersionMajor:    1,
			VersionMinor:    0,
		},
		ThreadID: input.ThreadID,
		Status:   "queued",
	}

	// Insert the email into the database
	if err := env.Emails.Insert(email); err != nil {
		utils.JSONResponse(w, 500, &EmailsCreateResponse{
			Success: false,
			Message: "internal server error - EM/CR/01",
		})

		env.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Could not insert an email into the database")
		return
	}

	// Add a send request to the queue
	err = env.NATS.Publish("send", email.ID)
	if err != nil {
		utils.JSONResponse(w, 500, &EmailsCreateResponse{
			Success: false,
			Message: "internal server error - EM/CR/03",
		})

		env.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Could not publish an email send request")
		return
	}

	utils.JSONResponse(w, 201, &EmailsCreateResponse{
		Success: true,
		Created: []string{email.ID},
	})
}

// EmailsGetResponse contains the result of the EmailsGet request.
type EmailsGetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

// EmailsGet responds with a single email message
func EmailsGet(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, 200, &EmailsGetResponse{
		Success: true,
		Status:  "sending",
	})
}

// EmailsUpdateResponse contains the result of the EmailsUpdate request.
type EmailsUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// EmailsUpdate does *something* - TODO
func EmailsUpdate(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, 501, &EmailsUpdateResponse{
		Success: false,
		Message: "Sorry, not implemented yet",
	})
}

// EmailsDeleteResponse contains the result of the EmailsDelete request.
type EmailsDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// EmailsDelete remvoes an email from the system
func EmailsDelete(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, 501, &EmailsDeleteResponse{
		Success: false,
		Message: "Sorry, not implemented yet",
	})
}

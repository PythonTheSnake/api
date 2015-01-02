package routes_test

import (
	"testing"
	"time"

	"github.com/franela/goreq"
	"github.com/stretchr/testify/require"

	"github.com/lavab/api/env"
	"github.com/lavab/api/models"
	"github.com/lavab/api/routes"
)

func TestTokensPrepareAccount(t *testing.T) {
	const (
		username = "jeremy"
		password = "potato"
	)

	// Prepare a token
	inviteToken := models.Token{
		Resource: models.MakeResource("", "test invite token"),
		Type:     "invite",
	}
	inviteToken.ExpireSoon()

	err := env.Tokens.Insert(inviteToken)
	require.Nil(t, err)

	// POST /accounts - invited
	result1, err := goreq.Request{
		Method:      "POST",
		Uri:         server.URL + "/accounts",
		ContentType: "application/json",
		Body: routes.AccountsCreateRequest{
			Username: username,
			Password: password,
			Token:    inviteToken.ID,
		},
	}.Do()
	require.Nil(t, err)

	// Unmarshal the response
	var response1 routes.AccountsCreateResponse
	err = result1.Body.FromJsonTo(&response1)
	require.Nil(t, err)

	// Check the result's contents
	require.True(t, response1.Success)
	require.Equal(t, "A new account was successfully created", response1.Message)
	require.NotEmpty(t, response1.Account.ID)

	accountID = response1.Account.ID
}

func TestTokensCreate(t *testing.T) {
	// log in as mr jeremy potato
	const (
		username = "jeremy"
		password = "potato"
	)
	// POST /accounts - classic
	request, err := goreq.Request{
		Method:      "POST",
		Uri:         server.URL + "/tokens",
		ContentType: "application/json",
		Body: routes.TokensCreateRequest{
			Username: username,
			Password: password,
			Type:     "auth",
		},
	}.Do()
	require.Nil(t, err)

	// Unmarshal the response
	var response routes.TokensCreateResponse
	err = request.Body.FromJsonTo(&response)
	require.Nil(t, err)

	// Check the result's contents
	require.True(t, response.Success)
	require.Equal(t, "Authentication successful", response.Message)
	require.NotEmpty(t, response.Token.ID)

	// Populate the global token variable
	authToken = response.Token.ID
}

func TestTokensCreateNonAuth(t *testing.T) {
	request, err := goreq.Request{
		Method:      "POST",
		Uri:         server.URL + "/tokens",
		ContentType: "application/json",
		Body: routes.TokensCreateRequest{
			Type: "not-auth",
		},
	}.Do()
	require.Nil(t, err)

	var response routes.TokensCreateResponse
	err = request.Body.FromJsonTo(&response)
	require.Nil(t, err)

	require.False(t, response.Success)
	require.Equal(t, "Only auth tokens are implemented", response.Message)
}

func TestTokensCreateWrongUsername(t *testing.T) {
	request, err := goreq.Request{
		Method:      "POST",
		Uri:         server.URL + "/tokens",
		ContentType: "application/json",
		Body: routes.TokensCreateRequest{
			Type:     "auth",
			Username: "not-jeremy",
			Password: "potato",
		},
	}.Do()
	require.Nil(t, err)

	var response routes.TokensCreateResponse
	err = request.Body.FromJsonTo(&response)
	require.Nil(t, err)

	require.False(t, response.Success)
	require.Equal(t, "Wrong username or password", response.Message)
}

func TestTokensCreateWrongPassword(t *testing.T) {
	request, err := goreq.Request{
		Method:      "POST",
		Uri:         server.URL + "/tokens",
		ContentType: "application/json",
		Body: routes.TokensCreateRequest{
			Type:     "auth",
			Username: "jeremy",
			Password: "not-potato",
		},
	}.Do()
	require.Nil(t, err)

	var response routes.TokensCreateResponse
	err = request.Body.FromJsonTo(&response)
	require.Nil(t, err)

	require.False(t, response.Success)
	require.Equal(t, "Wrong username or password", response.Message)
}

func TestTokensCreateInvalid(t *testing.T) {
	request, err := goreq.Request{
		Method:      "POST",
		Uri:         server.URL + "/tokens",
		ContentType: "application/json",
		Body:        "123123123###434$#$",
	}.Do()
	require.Nil(t, err)

	// Unmarshal the response
	var response routes.TokensCreateResponse
	err = request.Body.FromJsonTo(&response)
	require.Nil(t, err)

	// Check the result's contents
	require.False(t, response.Success)
	require.Equal(t, "Invalid input format", response.Message)
}

func TestTokensGet(t *testing.T) {
	request := goreq.Request{
		Method: "GET",
		Uri:    server.URL + "/tokens",
	}
	request.AddHeader("Authorization", "Bearer "+authToken)
	result, err := request.Do()
	require.Nil(t, err)

	var response routes.TokensGetResponse
	err = result.Body.FromJsonTo(&response)
	require.Nil(t, err)

	require.True(t, response.Success)
	require.True(t, response.Token.ExpiryDate.After(time.Now().UTC()))
}

func TestTokensDeleteById(t *testing.T) {
	const (
		username = "jeremy"
		password = "potato"
	)

	request1, err := goreq.Request{
		Method:      "POST",
		Uri:         server.URL + "/tokens",
		ContentType: "application/json",
		Body: routes.TokensCreateRequest{
			Username: username,
			Password: password,
			Type:     "auth",
		},
	}.Do()
	require.Nil(t, err)

	// Unmarshal the response
	var response1 routes.TokensCreateResponse
	err = request1.Body.FromJsonTo(&response1)
	require.Nil(t, err)

	// Check the result's contents
	require.True(t, response1.Success)
	require.Equal(t, "Authentication successful", response1.Message)
	require.NotEmpty(t, response1.Token.ID)

	request2 := goreq.Request{
		Method: "DELETE",
		Uri:    server.URL + "/tokens/" + response1.Token.ID,
	}
	request2.AddHeader("Authorization", "Bearer "+authToken)
	result2, err := request2.Do()
	require.Nil(t, err)

	var response2 routes.TokensDeleteResponse
	err = result2.Body.FromJsonTo(&response2)
	require.Nil(t, err)

	require.True(t, response2.Success)
	require.Equal(t, "Successfully logged out", response2.Message)
}

func TestTokensDeleteByInvalidID(t *testing.T) {
	request := goreq.Request{
		Method: "DELETE",
		Uri:    server.URL + "/tokens/123",
	}
	request.AddHeader("Authorization", "Bearer "+authToken)
	result, err := request.Do()
	require.Nil(t, err)

	var response routes.TokensDeleteResponse
	err = result.Body.FromJsonTo(&response)
	require.Nil(t, err)

	require.False(t, response.Success)
	require.Equal(t, "Invalid token ID", response.Message)
}

func TestTokensDeleteCurrent(t *testing.T) {
	request := goreq.Request{
		Method: "DELETE",
		Uri:    server.URL + "/tokens",
	}
	request.AddHeader("Authorization", "Bearer "+authToken)
	result, err := request.Do()
	require.Nil(t, err)

	var response routes.TokensDeleteResponse
	err = result.Body.FromJsonTo(&response)
	require.Nil(t, err)

	require.True(t, response.Success)
	require.Equal(t, "Successfully logged out", response.Message)
}

package routes_test

import (
	"testing"
	"time"

	"github.com/dchest/uniuri"
	"github.com/franela/goreq"
	. "github.com/smartystreets/goconvey/convey"
	//"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"github.com/lavab/api/env"
	"github.com/lavab/api/models"
	"github.com/lavab/api/routes"
)

func TestAccountsRoute(t *testing.T) {
	Convey("When creating a new account", t, func() {
		Convey("Misformatted body should fail", func() {
			result, err := goreq.Request{
				Method:      "POST",
				Uri:         server.URL + "/accounts",
				ContentType: "application/json",
				Body:        "!@#!@#",
			}.Do()
			So(err, ShouldBeNil)

			var response routes.AccountsCreateResponse
			err = result.Body.FromJsonTo(&response)
			So(err, ShouldBeNil)

			So(response.Success, ShouldBeFalse)
			So(response.Message, ShouldEqual, "Invalid input format")
		})

		Convey("Invalid set of data should fail", func() {
			result, err := goreq.Request{
				Method: "POST",
				Uri:    server.URL + "/accounts",
			}.Do()
			So(err, ShouldBeNil)

			var response routes.AccountsCreateResponse
			err = result.Body.FromJsonTo(&response)
			So(err, ShouldBeNil)

			So(response.Success, ShouldBeFalse)
			So(response.Message, ShouldEqual, "Invalid request")
		})

		Convey("Account creation should succeed", func() {
			var (
				username = uniuri.New()
				password = uniuri.New()
				email    = uniuri.New() + "@potato.org"
			)

			passwordHash := sha3.Sum256([]byte(password))
			accountPassword := string(passwordHash[:])
			result, err := goreq.Request{
				Method:      "POST",
				Uri:         server.URL + "/accounts",
				ContentType: "application/json",
				Body: routes.AccountsCreateRequest{
					Username: username,
					AltEmail: email,
				},
			}.Do()
			So(err, ShouldBeNil)

			var response routes.AccountsCreateResponse
			err = result.Body.FromJsonTo(&response)
			So(err, ShouldBeNil)

			So(response.Message, ShouldEqual, "Your account has been added to the beta queue")
			So(response.Success, ShouldBeTrue)
			So(response.Account.ID, ShouldNotBeEmpty)

			account := response.Account

			Convey("Duplicating the username should fail", func() {
				result, err := goreq.Request{
					Method:      "POST",
					Uri:         server.URL + "/accounts",
					ContentType: "application/json",
					Body: routes.AccountsCreateRequest{
						Username: username,
						AltEmail: email,
					},
				}.Do()
				So(err, ShouldBeNil)

				var response routes.AccountsCreateResponse
				err = result.Body.FromJsonTo(&response)
				So(err, ShouldBeNil)

				So(response.Success, ShouldBeFalse)
				So(response.Message, ShouldEqual, "Username already used")
			})

			Convey("Duplicating the email should fail", func() {
				result, err := goreq.Request{
					Method:      "POST",
					Uri:         server.URL + "/accounts",
					ContentType: "application/json",
					Body: routes.AccountsCreateRequest{
						Username: uniuri.New(),
						AltEmail: email,
					},
				}.Do()
				So(err, ShouldBeNil)

				var response routes.AccountsCreateResponse
				err = result.Body.FromJsonTo(&response)
				So(err, ShouldBeNil)

				So(response.Success, ShouldBeFalse)
				So(response.Message, ShouldEqual, "Email already used")
			})

			Convey("Verification with an invalid username should fail", func() {
				result, err := goreq.Request{
					Method:      "POST",
					Uri:         server.URL + "/accounts",
					ContentType: "application/json",
					Body: routes.AccountsCreateRequest{
						Username:   uniuri.New(),
						InviteCode: uniuri.New(),
					},
				}.Do()
				So(err, ShouldBeNil)

				var response routes.AccountsCreateResponse
				err = result.Body.FromJsonTo(&response)
				So(err, ShouldBeNil)

				So(response.Message, ShouldEqual, "Invalid username")
				So(response.Success, ShouldBeFalse)
			})

			Convey("Verification with an invalid code should fail", func() {
				result, err := goreq.Request{
					Method:      "POST",
					Uri:         server.URL + "/accounts",
					ContentType: "application/json",
					Body: routes.AccountsCreateRequest{
						Username:   account.Name,
						InviteCode: uniuri.New(),
					},
				}.Do()
				So(err, ShouldBeNil)

				var response routes.AccountsCreateResponse
				err = result.Body.FromJsonTo(&response)
				So(err, ShouldBeNil)

				So(response.Message, ShouldEqual, "Invalid invitation code")
				So(response.Success, ShouldBeFalse)
			})

			Convey("Verification with a not owned code should fail", func() {
				verificationToken := models.Token{
					Resource: models.MakeResource("top kek", "test verification token"),
					Type:     "verify",
				}
				verificationToken.ExpireSoon()

				err := env.Tokens.Insert(verificationToken)
				So(err, ShouldBeNil)

				result, err := goreq.Request{
					Method:      "POST",
					Uri:         server.URL + "/accounts",
					ContentType: "application/json",
					Body: routes.AccountsCreateRequest{
						Username:   account.Name,
						InviteCode: verificationToken.ID,
					},
				}.Do()
				So(err, ShouldBeNil)

				var response routes.AccountsCreateResponse
				err = result.Body.FromJsonTo(&response)
				So(err, ShouldBeNil)

				So(response.Message, ShouldEqual, "Invalid invitation code")
				So(response.Success, ShouldBeFalse)
			})

			Convey("Verification with a token that is not a verification token should fail", func() {
				verificationToken := models.Token{
					Resource: models.MakeResource(account.ID, "test verification token"),
					Type:     "notverify",
				}
				verificationToken.ExpireSoon()

				err := env.Tokens.Insert(verificationToken)
				So(err, ShouldBeNil)

				result, err := goreq.Request{
					Method:      "POST",
					Uri:         server.URL + "/accounts",
					ContentType: "application/json",
					Body: routes.AccountsCreateRequest{
						Username:   account.Name,
						InviteCode: verificationToken.ID,
					},
				}.Do()
				So(err, ShouldBeNil)

				var response routes.AccountsCreateResponse
				err = result.Body.FromJsonTo(&response)
				So(err, ShouldBeNil)

				So(response.Message, ShouldEqual, "Invalid invitation code")
				So(response.Success, ShouldBeFalse)
			})

			Convey("Verification with an expired invitation code should fail", func() {
				verificationToken := models.Token{
					Resource: models.MakeResource(account.ID, "test verification token"),
					Type:     "verify",
				}
				verificationToken.ExpiryDate = time.Now().Truncate(time.Hour * 24)

				err := env.Tokens.Insert(verificationToken)
				So(err, ShouldBeNil)

				result, err := goreq.Request{
					Method:      "POST",
					Uri:         server.URL + "/accounts",
					ContentType: "application/json",
					Body: routes.AccountsCreateRequest{
						Username:   account.Name,
						InviteCode: verificationToken.ID,
					},
				}.Do()
				So(err, ShouldBeNil)

				var response routes.AccountsCreateResponse
				err = result.Body.FromJsonTo(&response)
				So(err, ShouldBeNil)

				So(response.Message, ShouldEqual, "Expired invitation code")
				So(response.Success, ShouldBeFalse)
			})

			Convey("Verification of the account should succeed", func() {
				verificationToken := models.Token{
					Resource: models.MakeResource(account.ID, "test verification token"),
					Type:     "verify",
				}
				verificationToken.ExpireSoon()

				err := env.Tokens.Insert(verificationToken)
				So(err, ShouldBeNil)

				result, err := goreq.Request{
					Method:      "POST",
					Uri:         server.URL + "/accounts",
					ContentType: "application/json",
					Body: routes.AccountsCreateRequest{
						Username:   username,
						InviteCode: verificationToken.ID,
					},
				}.Do()
				So(err, ShouldBeNil)

				var response routes.AccountsCreateResponse
				err = result.Body.FromJsonTo(&response)
				So(err, ShouldBeNil)

				So(response.Message, ShouldEqual, "Valid token was provided")
				So(response.Success, ShouldBeTrue)

				Convey("Setup with a weak password should fail", func() {
					result, err := goreq.Request{
						Method:      "POST",
						Uri:         server.URL + "/accounts",
						ContentType: "application/json",
						Body: routes.AccountsCreateRequest{
							Username:   account.Name,
							InviteCode: verificationToken.ID,
							Password:   "c0067d4af4e87f00dbac63b6156828237059172d1bbeac67427345d6a9fda484",
						},
					}.Do()

					var response routes.AccountsCreateResponse
					err = result.Body.FromJsonTo(&response)
					So(err, ShouldBeNil)

					So(response.Message, ShouldEqual, "Weak password")
					So(response.Success, ShouldBeFalse)
				})

				Convey("Setup with an invalid username should fail", func() {
					result, err := goreq.Request{
						Method:      "POST",
						Uri:         server.URL + "/accounts",
						ContentType: "application/json",
						Body: routes.AccountsCreateRequest{
							Username:   uniuri.New(),
							InviteCode: verificationToken.ID,
							Password:   accountPassword,
						},
					}.Do()
					So(err, ShouldBeNil)

					var response routes.AccountsCreateResponse
					err = result.Body.FromJsonTo(&response)
					So(err, ShouldBeNil)

					So(response.Message, ShouldEqual, "Invalid username")
					So(response.Success, ShouldBeFalse)
				})

				Convey("Setup with an invalid code should fail", func() {
					result, err := goreq.Request{
						Method:      "POST",
						Uri:         server.URL + "/accounts",
						ContentType: "application/json",
						Body: routes.AccountsCreateRequest{
							Username:   account.Name,
							InviteCode: uniuri.New(),
							Password:   accountPassword,
						},
					}.Do()
					So(err, ShouldBeNil)

					var response routes.AccountsCreateResponse
					err = result.Body.FromJsonTo(&response)
					So(err, ShouldBeNil)

					So(response.Message, ShouldEqual, "Invalid invitation code")
					So(response.Success, ShouldBeFalse)
				})

				Convey("Setup with a code that user does not own should fail", func() {
					verificationToken := models.Token{
						Resource: models.MakeResource(uniuri.New(), "test verification token"),
						Type:     "verify",
					}
					verificationToken.ExpireSoon()

					err := env.Tokens.Insert(verificationToken)
					So(err, ShouldBeNil)

					result, err := goreq.Request{
						Method:      "POST",
						Uri:         server.URL + "/accounts",
						ContentType: "application/json",
						Body: routes.AccountsCreateRequest{
							Username:   account.Name,
							InviteCode: verificationToken.ID,
							Password:   accountPassword,
						},
					}.Do()
					So(err, ShouldBeNil)

					var response routes.AccountsCreateResponse
					err = result.Body.FromJsonTo(&response)
					So(err, ShouldBeNil)

					So(response.Message, ShouldEqual, "Invalid invitation code")
					So(response.Success, ShouldBeFalse)
				})

				Convey("Setup with a token that is not a verification token should fail", func() {
					verificationToken := models.Token{
						Resource: models.MakeResource(account.ID, "test verification token"),
						Type:     "notverify",
					}
					verificationToken.ExpireSoon()

					err := env.Tokens.Insert(verificationToken)
					So(err, ShouldBeNil)

					result, err := goreq.Request{
						Method:      "POST",
						Uri:         server.URL + "/accounts",
						ContentType: "application/json",
						Body: routes.AccountsCreateRequest{
							Username:   account.Name,
							InviteCode: verificationToken.ID,
							Password:   accountPassword,
						},
					}.Do()
					So(err, ShouldBeNil)

					var response routes.AccountsCreateResponse
					err = result.Body.FromJsonTo(&response)
					So(err, ShouldBeNil)

					So(response.Message, ShouldEqual, "Invalid invitation code")
					So(response.Success, ShouldBeFalse)
				})

				Convey("Setup with a token that expired should fail", func() {
					verificationToken := models.Token{
						Resource: models.MakeResource(account.ID, "test verification token"),
						Type:     "verify",
					}
					verificationToken.ExpiryDate = time.Now().Truncate(time.Hour * 24)

					err := env.Tokens.Insert(verificationToken)
					So(err, ShouldBeNil)

					result, err := goreq.Request{
						Method:      "POST",
						Uri:         server.URL + "/accounts",
						ContentType: "application/json",
						Body: routes.AccountsCreateRequest{
							Username:   account.Name,
							InviteCode: verificationToken.ID,
							Password:   accountPassword,
						},
					}.Do()
					So(err, ShouldBeNil)

					// Unmarshal the response
					var response routes.AccountsCreateResponse
					err = result.Body.FromJsonTo(&response)
					So(err, ShouldBeNil)

					// Check the result's contents
					So(response.Message, ShouldEqual, "Expired invitation code")
					So(response.Success, ShouldBeFalse)
				})

				Convey("Setup with proper data should succeed", func() {
					result, err := goreq.Request{
						Method:      "POST",
						Uri:         server.URL + "/accounts",
						ContentType: "application/json",
						Body: routes.AccountsCreateRequest{
							Username:   account.Name,
							InviteCode: verificationToken.ID,
							Password:   accountPassword,
						},
					}.Do()
					So(err, ShouldBeNil)

					var response routes.AccountsCreateResponse
					err = result.Body.FromJsonTo(&response)
					So(err, ShouldBeNil)

					So(response.Message, ShouldEqual, "Your account has been initialized successfully")
					So(response.Success, ShouldBeTrue)

					Convey("After acquiring an authentication token", func() {
						request, err := goreq.Request{
							Method:      "POST",
							Uri:         server.URL + "/tokens",
							ContentType: "application/json",
							Body: routes.TokensCreateRequest{
								Username: account.Name,
								Password: accountPassword,
								Type:     "auth",
							},
						}.Do()
						So(err, ShouldBeNil)

						var response routes.TokensCreateResponse
						err = request.Body.FromJsonTo(&response)
						So(err, ShouldBeNil)

						So(response.Message, ShouldEqual, "Authentication successful")
						So(response.Success, ShouldBeTrue)
						So(response.Token.ID, ShouldNotBeEmpty)

						authToken := response.Token.ID

						Convey("Accounts list query should return a proper response", func() {
							request := goreq.Request{
								Method: "GET",
								Uri:    server.URL + "/accounts",
							}
							request.AddHeader("Authorization", "Bearer "+authToken)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsListResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Success, ShouldBeFalse)
							So(response.Message, ShouldEqual, "Sorry, not implemented yet")
						})

						Convey("Getting own account information should return the account information", func() {
							request := goreq.Request{
								Method: "GET",
								Uri:    server.URL + "/accounts/me",
							}
							request.AddHeader("Authorization", "Bearer "+authToken)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsGetResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Success, ShouldBeTrue)
							So(response.Account.Name, ShouldEqual, "jeremy")
						})

						Convey("Getting any non-me account should return a proper response", func() {
							request := goreq.Request{
								Method: "GET",
								Uri:    server.URL + "/accounts/not-me",
							}
							request.AddHeader("Authorization", "Bearer "+authToken)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsGetResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Success, ShouldBeFalse)
							So(response.Message, ShouldEqual, `Only the "me" user is implemented`)
						})

						Convey("Updating own account should succeed", func() {
							newPasswordHashBytes := sha3.Sum256([]byte("cabbage123"))
							newPasswordHash := string(newPasswordHashBytes[:])

							request := goreq.Request{
								Method:      "PUT",
								Uri:         server.URL + "/accounts/me",
								ContentType: "application/json",
								Body: &routes.AccountsUpdateRequest{
									CurrentPassword: accountPassword,
									NewPassword:     newPasswordHash,
									AltEmail:        "john.cabbage@example.com",
								},
							}
							request.AddHeader("Authorization", "Bearer "+authToken)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsUpdateResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Message, ShouldEqual, "Your account has been successfully updated")
							So(response.Success, ShouldBeTrue)
							So(response.Account.Name, ShouldEqual, "jeremy")
							So(response.Account.AltEmail, ShouldEqual, "john.cabbage@example.com")
						})

						Convey("Updating with an invalid body should fail", func() {
							request := goreq.Request{
								Method:      "PUT",
								Uri:         server.URL + "/accounts/me",
								ContentType: "application/json",
								Body:        "123123123!@#!@#!@#",
							}
							request.AddHeader("Authorization", "Bearer "+authToken)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsUpdateResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Message, ShouldEqual, "Invalid input format")
							So(response.Success, ShouldBeFalse)
						})

						Convey("Trying to update not own account should fail", func() {
							request := goreq.Request{
								Method:      "PUT",
								Uri:         server.URL + "/accounts/not-me",
								ContentType: "application/json",
								Body: &routes.AccountsUpdateRequest{
									CurrentPassword: "potato",
									NewPassword:     "cabbage",
									AltEmail:        "john.cabbage@example.com",
								},
							}
							request.AddHeader("Authorization", "Bearer "+authToken)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsUpdateResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Message, ShouldEqual, `Only the "me" user is implemented`)
							So(response.Success, ShouldBeFalse)
						})

						Convey("Trying to update with an invalid password should fail", func() {
							request := goreq.Request{
								Method:      "PUT",
								Uri:         server.URL + "/accounts/me",
								ContentType: "application/json",
								Body: &routes.AccountsUpdateRequest{
									CurrentPassword: "potato2",
									NewPassword:     "cabbage",
									AltEmail:        "john.cabbage@example.com",
								},
							}
							request.AddHeader("Authorization", "Bearer "+authToken)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsUpdateResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Message, ShouldEqual, "Invalid current password")
							So(response.Success, ShouldBeFalse)
						})

						Convey("Wiping not own account should fail", func() {
							request := goreq.Request{
								Method: "POST",
								Uri:    server.URL + "/accounts/not-me/wipe-data",
							}
							request.AddHeader("Authorization", "Bearer "+authToken)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsWipeDataResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Message, ShouldEqual, `Only the "me" user is implemented`)
							So(response.Success, ShouldBeFalse)
						})

						Convey("Wiping own account should succeed", func() {
							request := goreq.Request{
								Method: "POST",
								Uri:    server.URL + "/accounts/me/wipe-data",
							}
							request.AddHeader("Authorization", "Bearer "+authToken)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsWipeDataResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Message, ShouldEqual, "Your account has been successfully wiped")
							So(response.Success, ShouldBeTrue)
						})

						Convey("Deleting not own account should fail", func() {
							token := models.Token{
								Resource: models.MakeResource(account.ID, "test invite token"),
								Type:     "auth",
							}
							token.ExpireSoon()

							err := env.Tokens.Insert(token)
							So(err, ShouldBeNil)

							request := goreq.Request{
								Method: "DELETE",
								Uri:    server.URL + "/accounts/not-me",
							}
							request.AddHeader("Authorization", "Bearer "+token.ID)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsWipeDataResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Message, ShouldEqual, `Only the "me" user is implemented`)
							So(response.Success, ShouldBeFalse)
						})

						Convey("Deleting own account should succeed", func() {
							token := models.Token{
								Resource: models.MakeResource(account.ID, "test invite token"),
								Type:     "auth",
							}
							token.ExpireSoon()

							err := env.Tokens.Insert(token)
							So(err, ShouldBeNil)

							request := goreq.Request{
								Method: "DELETE",
								Uri:    server.URL + "/accounts/me",
							}
							request.AddHeader("Authorization", "Bearer "+token.ID)
							result, err := request.Do()
							So(err, ShouldBeNil)

							var response routes.AccountsWipeDataResponse
							err = result.Body.FromJsonTo(&response)
							So(err, ShouldBeNil)

							So(response.Message, ShouldEqual, "Your account has been successfully deleted")
							So(response.Success, ShouldBeTrue)
						})
					})
				})
			})
		})
	})
}

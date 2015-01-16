package routes_test

import (
	"testing"

	"github.com/dchest/uniuri"
	"github.com/franela/goreq"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/lavab/api/env"
	"github.com/lavab/api/models"
	"github.com/lavab/api/routes"
)

func TestAttachmentsRoute(t *testing.T) {
	Convey("When uploading a new attachment", t, func() {
		account := &models.Account{
			Resource: models.MakeResource("", "johnorange")
			Status: "complete",
			AltEmail: "john@orange.org",
		}
		err := account.SetPassword("fruityloops")
		So(err, ShouldBeNil)

		err = env.Accounts.Insert(account)
		So(err, ShouldBeNil)

		result, err := goreq.Request{
			Method:      "POST",
			Uri:         server.URL + "/tokens",
			ContentType: "application/json",
			Body:        `{
				"type": "auth",
				"username": "johnorange",
				"password": "fruityloops"
			}`,
		}.Do()
		So(err, ShouldBeNil)

		var response routes.TokensCreateResponse
		err = result.Body.FromJsonTo(&response)
		So(err, ShouldBeNil)

		So(response.Success, ShouldBeTrue)
		authToken := response.Token

		Convey("Misformatted body should fail", func() {
			request := goreq.Request{
				Method:      "POST",
				Uri:         server.URL + "/attachments",
				ContentType: "application/json",
				Body:        "!@#!@#",
			}
			request.AddHeader("Authorization", "Bearer "+authToken.ID)
			result, err := request.Do()
			So(err, ShouldBeNil)

			var response routes.AttachmentsCreateResponse
			err = result.Body.FromJsonTo(&response)
			So(err, ShouldBeNil)

			So(response.Success, ShouldBeFalse)
			So(response.Message, ShouldEqual, "Invalid input format")
		})

		Convey("Invalid set of data should fail", func() {
			request := goreq.Request{
				Method: "POST",
				Uri:    server.URL + "/attachments",
			}
			request.AddHeader("Authorization", "Bearer "+authToken.ID)
			result, err := request.Do()
			So(err, ShouldBeNil)

			var response routes.AccountsCreateResponse
			err = result.Body.FromJsonTo(&response)
			So(err, ShouldBeNil)

			So(response.Success, ShouldBeFalse)
			So(response.Message, ShouldEqual, "Invalid request")
		})

		Convey("Attachment upload should succeed", func() {
			request := goreq.Request{
				Method:      "POST",
				Uri:         server.URL + "/attachments",
				ContentType: "application/json",
				Body: `{
	"data": "` + uniuri.NewLen(64) + `",
	"name": "` + uniuri.New() + `"
	"encoding": "json",
	"version_major": 1,
	"version_minor": 0,
	"pgp_fingerprints": ["` + uniuri.New() + `"]
}`,
			}
			request.AddHeader("Authorization", "Bearer "+authToken.ID)
			result, err := request.Do()
			So(err, ShouldBeNil)

			var response routes.AttachmentsCreateResponse
			err = result.Body.FromJsonTo(&response)
			So(err, ShouldBeNil)

			So(response.Message, ShouldEqual, "A new attachment was successfully created")
			So(response.Success, ShouldBeTrue)
			So(response.Attachment.ID, ShouldNotBeEmpty)

			attachment := response.Attachment

			Convey("Getting that attachment should succeed", func() {
				request := goreq.Request{
					Method: "GET",
					Uri:    server.URL + "/attachments/" + attachment.ID,
				}
			request.AddHeader("Authorization", "Bearer "+authToken.ID)
			result, err := request.Do()
			So(err, ShouldBeNil)

				var response routes.AttachmentsGetResponse
				err = result.Body.FromJsonTo(&response)
				So(err, ShouldBeNil)

				So(response.Attachment.ID, ShouldNotBeNil)
				So(response.Success, ShouldBeTrue)
			})
		})

		Convey("Getting a non-existing attachment should fail", func() {
			request := goreq.Request{
				Method: "GET",
				Uri:    server.URL + "/attachments/doesntexist",
			}
			request.AddHeader("Authorization", "Bearer "+authToken.ID)
			result, err := request.Do()
			So(err, ShouldBeNil)

			var response routes.AttachmentsGetResponse
			err = result.Body.FromJsonTo(&response)
			So(err, ShouldBeNil)

			So(response.Message, ShouldEqual, "Attachment not found")
			So(response.Success, ShouldBeFalse)
		})

		Convey("Getting a non-owned attachment should fail", func() {
			attachment := &models.Attachment{
				Encrypted: models.Encrypted{
					Encoding:        "json",
					Data:            uniuri.NewLen(64),
					Schema:          "attachment",
					VersionMajor:    1,
					VersionMinor:    0,
					PGPFingerprints: []string{uniuri.New()},
				},
				Resource: models.MakeResource("nonowned", "photo.jpg"),
			}

			err := env.Attachments.Insert(attachment)
			So(err, ShouldBeNil)

			request := goreq.Request{
				Method: "GET",
				Uri:    server.URL + "/attachments/" + attachment.ID,
			}
			request.AddHeader("Authorization", "Bearer "+authToken.ID)
			result, err := request.Do()
			So(err, ShouldBeNil)

			var response routes.AttachmentsGetResponse
			err = result.Body.FromJsonTo(&response)
			So(err, ShouldBeNil)

			So(response.Message, ShouldEqual, "Attachment not found")
			So(response.Success, ShouldBeFalse)
		})
	})
}

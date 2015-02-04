package db

import (
	"github.com/lavab/api/models"

	"github.com/dancannon/gorethink"
)

type AttachmentsTable struct {
	RethinkCRUD
	Emails *EmailsTable
}

func (a *AttachmentsTable) GetAttachment(id string) (*models.Attachment, error) {
	var result models.Attachment

	if err := a.FindFetchOne(id, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *AttachmentsTable) GetOwnedBy(id string) ([]*models.Attachment, error) {
	var result []*models.Attachment

	err := a.WhereAndFetch(map[string]interface{}{
		"owner": id,
	}, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *AttachmentsTable) DeleteOwnedBy(id string) error {
	return a.Delete(map[string]interface{}{
		"owner": id,
	})
}

func (a *AttachmentsTable) GetEmailAttachments(id string) ([]*models.Attachment, error) {
	email, err := a.Emails.GetEmail(id)
	if err != nil {
		return nil, err
	}

	query, err := a.Emails.GetTable().Filter(func(row gorethink.Term) gorethink.Term {
		return gorethink.Expr(email.Attachments).Contains(row.Field("id"))
	}).GetAll().Run(a.Emails.GetSession())
	if err != nil {
		return nil, err
	}

	var result []*models.Attachment
	err = query.All(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *AttachmentsTable) CountByEmail(id string) (int, error) {
	query, err := a.GetTable().GetAllByIndex("owner", id).Count().Run(a.GetSession())
	if err != nil {
		return 0, err
	}

	var result int
	err = query.One(&result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (a *AttachmentsTable) CountByThread(id ...interface{}) (int, error) {
	query, err := a.GetTable().Filter(func(row gorethink.Term) gorethink.Term {
		return gorethink.Table("emails").GetAllByIndex("owner", id...).Field("attachments").Contains(row.Field("id"))
	}).Count().Run(a.GetSession())
	if err != nil {
		return 0, err
	}

	var result int
	err = query.One(&result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

package db

import (
	"github.com/dancannon/gorethink"

	"github.com/lavab/api/models"
)

// Emails implements the CRUD interface for tokens
type EmailsTable struct {
	RethinkCRUD
}

// GetEmail returns a token with specified name
func (e *EmailsTable) GetEmail(id string) (*models.Email, error) {
	var result models.Email

	if err := e.FindFetchOne(id, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetOwnedBy returns all emails owned by id
func (e *EmailsTable) GetOwnedBy(id string) ([]*models.Email, error) {
	var result []*models.Email

	err := e.WhereAndFetch(map[string]interface{}{
		"owner": id,
	}, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteOwnedBy deletes all emails owned by id
func (e *EmailsTable) DeleteOwnedBy(id string) error {
	return e.Delete(map[string]interface{}{
		"owner": id,
	})
}

func (e *EmailsTable) CountOwnedBy(id string) (int, error) {
	return e.FindByAndCount("owner", id)
}

func (e *EmailsTable) List(
	owner string,
	sort []string,
	offset int,
	limit int,
) ([]*models.Email, error) {
	// Filter by owner's ID
	term := e.GetTable().Filter(map[string]interface{}{
		"owner": owner,
	})

	// If sort array has contents, parse them and add to the term
	if sort != nil && len(sort) > 0 {
		var conds []interface{}
		for _, cond := range sort {
			if cond[0] == '-' {
				conds = append(conds, gorethink.Desc(cond[1:]))
			} else if cond[0] == '+' || cond[0] == ' ' {
				conds = append(conds, gorethink.Asc(cond[1:]))
			} else {
				conds = append(conds, gorethink.Asc(cond))
			}
		}

		term = term.OrderBy(conds...)
	}

	// Slice the result in 3 cases
	if offset != 0 && limit == 0 {
		term = term.Skip(offset)
	}

	if offset == 0 && limit != 0 {
		term = term.Limit(limit)
	}

	if offset != 0 && limit != 0 {
		term = term.Slice(offset, offset+limit)
	}

	// Run the query
	cursor, err := term.Run(e.GetSession())
	if err != nil {
		return nil, err
	}

	// Fetch the cursor
	var resp []*models.Email
	err = cursor.All(&resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
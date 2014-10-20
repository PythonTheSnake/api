package models

// User holds the data fields for a Lavaboom user
type User struct {
	ID       string `json:"id" gorethink:"id"`
	Name     string `json:"name" gorethink:"name"`
	Password string `json:"-"  gorethink:"password"`
	Type     string `json:"-" gorethink:"type"` // std, beta, admin

	Pgp      PGP          `json:"-"  gorethink:"pgp"`
	Settings SettingsData `json:"-"  gorethink:"settings"`
	Billing  BillingData  `json:"-"  gorethink:"billing"`
}

// PGP TODO
type PGP struct {
	PublicKey string `json:"public_key"  gorethink:"public_key"`
	Finger    string `json:"finger"  gorethink:"finger"`
	ExpDate   string `json:"exp_date"  gorethink:"exp_date" actual_type:"time.Time"`
}

// SettingsData TODO will do later, contains settings stuff
type SettingsData struct {
}

type BillingData struct {
}

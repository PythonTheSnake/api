package env

// Flags contains values of flags which are important in the whole API
type Flags struct {
	BindAddress      string
	APIVersion       string
	LogFormatterType string
	ForceColors      bool

	SessionDuration     int
	ClassicRegistration bool
	UsernameReservation bool

	RethinkDBURL      string
	RethinkDBKey      string
	RethinkDBDatabase string
}

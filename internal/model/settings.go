package model

import "encoding/json"

// Settings struct represents the model used in database, table `settings`.
// Store all data needed to form and send receipts.
type Settings struct {
	Emails      map[string]string `db:"-"`      // map 'Payer's Full Name' -> 'Payer's email'
	EmailsJSON  string            `db:"Emails"` // emails are stored in DB as json string
	SenderEmail string            `db:"SenderEmail"`
}

// BeforeSave fills EmailsJSON field with serialized data to save it in DB
func (s *Settings) BeforeSave() error {
	b, err := json.Marshal(s.Emails)
	if err != nil {
		return err
	}
	s.EmailsJSON = string(b)
	return nil
}

// AfterLoad decodes json string fetched from DB into Emails map
func (s *Settings) AfterLoad() error {
	if s.EmailsJSON == "" {
		s.Emails = make(map[string]string)
		return nil
	}
	return json.Unmarshal([]byte(s.EmailsJSON), &s.Emails)
}

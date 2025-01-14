package model

// File struct represents the model used in database, table `files`. Listed on history page
type File struct {
	FileName string `db:"FileName"`
	FileData []byte `db:"File"`
}

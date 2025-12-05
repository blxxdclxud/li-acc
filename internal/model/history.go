package model

// File struct represents the model used in database, table `files`. Listed on history page
type File struct {
	FileName string `db:"FileName" json:"file_name"`
	FileData []byte `db:"File" json:"file_data"`
	FilePath string `db:"-" json:"file_path"`
}

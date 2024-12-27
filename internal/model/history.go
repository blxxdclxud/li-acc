package model

type File struct {
	FileName string `db:"FileName"`
	FileData []byte `db:"File"`
}

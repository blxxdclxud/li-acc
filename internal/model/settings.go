package model

type Settings struct {
	Id          int    `db:"Id"`
	ReceiptFile []byte `db:"ReceiptFile"`
	Emails      []byte `db:"Emails"`
	QrPattern   string `db:"QrPattern"`
	SenderEmail string `db:"SenderEmail"`
}

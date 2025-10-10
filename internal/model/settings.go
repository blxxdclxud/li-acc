package model

// Settings struct represents the model used in database, table `settings`.
// Store all data needed to form and send receipts.
type Settings struct {
	ReceiptFile []byte `db:"ReceiptFile"`
	Emails      []byte `db:"Emails"`
	QrPattern   string `db:"QrPattern"`
	SenderEmail string `db:"SenderEmail"`
}

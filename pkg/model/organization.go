package model

type Organization struct {
	Name        string `json:"Наименование организации"`
	PersonalAcc string `json:"Расчетный счет"`
	BankName    string `json:"Наименование банка"`
	BIC         string `json:"БИК"`
	CorrespAcc  string `json:"Корреспондентский счет"`
	PayeeINN    string `json:"ИНН"`
	KPP         string `json:"КПП"`
	ExtraParams string `json:"Дополнительные параметры ДШК" includeQr:"false"` // tag `includeQr` means "include to QrCode data"
}

package model

type Payer struct {
	PersAcc  string `json:"Лицевой счет"`
	CHILDFIO string `json:"ФИО обучающегося"`
	Purpose  string `json:"Назначение"`
	CBC      string `json:"КБК"`
	OKTMO    string `json:"ОКТМО"`
	Sum      string `json:"Сумма"`
}

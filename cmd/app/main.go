package main

import (
	"fmt"
	"li-acc/config"
	"li-acc/pkg/logger"
	"li-acc/pkg/sender"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		zap.L().Fatal("filed to load .env file", zap.Error(err))
	}

	err = logger.Init("development")
	if err != nil {
		zap.L().Fatal("cannot initialize logger: ", zap.Error(err))
	}

	//user := os.Getenv("POSTGRES_USER")
	//password := os.Getenv("POSTGRES_PASSWORD")
	//dbName := os.Getenv("POSTGRES_DB")
	//dbHost := os.Getenv("POSTGRES_HOST")
	//dbPort := os.Getenv("POSTGRES_PORT")
	//
	//repo, err := repository.ConnectDB(
	//	fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
	//		user, password, dbHost, dbPort, dbName),
	//)
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//changed, err := db.ApplyMigrations(repo.DB, "internal/db/migrations")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(changed)

	//c, err := converter.NewConverter()
	//err = c.Convert("./tmp/receipt_pattern.xlsx", "./tmp/receipt_pattern.pdf")
	//
	//fmt.Println(err)

	//org, err := xls.ParseSettings("payers.xls")
	//payers, err := xls.ParsePayers("payers.xls")

	//err = xls.FillOrganizationParamsInReceipt("./assets/excel/blank_receipt_pattern.xls", *org)
	//fmt.Println(err)

	//q := qr.NewQrPattern(*org)
	//data := q.GetPayersQrDataString(payers[0])
	//err = q.GenerateQRCode(data)

	//err = pdf.GeneratePersonalReceipt("./tmp/receipt_pattern.pdf", "./TEST.pdf", qr.QRCodePath, payers[0])

	// Just test the code:
	cfg := config.LoadConfig()
	fmt.Println(cfg)
	return

	smtpHost, smtpPort := cfg.SMTP.Host, cfg.SMTP.Port
	smtpEmail, smtpPassword := os.Getenv("SMTP_EMAIL"), os.Getenv("SMTP_PASSWORD")

	s := sender.NewSender(smtpHost, smtpPort, smtpEmail, smtpPassword)
	wg := &sync.WaitGroup{}

	emails := []string{"blabla@gmail.com", "bla@gmail.com", "r.bla@bla.university"}

	status := make(chan sender.EmailStatus, len(emails))

	go func() {
		for st := range status {
			fmt.Println(st.StatusType, st.Status)
		}
	}()

	for _, email := range emails {

		msg := sender.FormMessage("Payment", "", "./TEST.pdf", smtpEmail, email)

		wg.Add(1)
		go s.SendEmail(msg.Msg, status, wg)
	}

	fmt.Println("wait")
	wg.Wait()

	close(status)

	fmt.Println(err)
}

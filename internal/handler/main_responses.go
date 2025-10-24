package handler

type PayersFileUploadResponseSuccess struct {
	Message string `json:"message"`
}

type PayersFileUploadResponse struct {
	Message        string   `json:"message"`                  // summary message for user
	SentAmount     int      `json:"sent_amount,omitempty"`    // number of sent emails
	FailedEmails   []string `json:"failed_emails,omitempty"`  // emails list from EmailSendingError
	MissingPayers  []string `json:"missing_payers,omitempty"` // payers list from EmailMappingError
	PartialSuccess bool     `json:"partial_success"`          // indicates partial failure occurred
}

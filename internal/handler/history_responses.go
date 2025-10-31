package handler

import "li-acc/internal/model"

type FilesHistoryResponse struct {
	Files []model.File `json:"files"`
}

package errs

import "errors"

// IsUserError возвращает true, если ошибка вызвана действиями пользователя.
func IsUserError(err error) bool {
	var ce CodedError
	if errors.As(err, &ce) {
		return ce.Kind() == User || ce.Kind() == Validation
	}
	return false
}

// IsSystemError возвращает true, если ошибка системная.
func IsSystemError(err error) bool {
	var ce CodedError
	if errors.As(err, &ce) {
		return ce.Kind() == System
	}
	return false
}

// IsExternalError возвращает true, если ошибка связана с внешними сервисами.
func IsExternalError(err error) bool {
	var ce CodedError
	if errors.As(err, &ce) {
		return ce.Kind() == External
	}
	return false
}

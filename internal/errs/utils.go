package errs

import "errors"

// IsUserError возвращает true, если ошибка вызвана действиями пользователя.
func IsUserError(err error) bool {
	return hasKind(err, User)
}

// IsSystemError возвращает true, если ошибка системная.
func IsSystemError(err error) bool {
	return hasKind(err, System)
}

// IsValidationError возвращает true, если ошибка связана с внешними сервисами.
func IsValidationError(err error) bool {
	return hasKind(err, Validation)
}

func hasKind(err error, kind Kind) bool {
	// Проходим по всей цепочке ошибок
	for err != nil {
		var ce CodedError
		if errors.As(err, &ce) && ce.Kind() == kind {
			return true
		}
		// Переходим к следующей ошибке в цепочке
		err = errors.Unwrap(err)
	}
	return false
}

package common

import "net/http"

// WriteInternalError не отдаёт клиенту текст внутренней ошибки (SQL, stack и т.д.).
func WriteInternalError(w http.ResponseWriter, publicMessage string) {
	if publicMessage == "" {
		publicMessage = "internal server error"
	}
	http.Error(w, publicMessage, http.StatusInternalServerError)
}

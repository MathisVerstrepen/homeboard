package handlers

import (
	"net/http"

	"diikstra.fr/homeboard/components"
	"github.com/labstack/echo/v4"
)

func ThrowClientError(c echo.Context, err error) {
	Render(c, http.StatusInternalServerError, components.Alert(
		"error", "Une erreur est survenue", err.Error(),
	))
}

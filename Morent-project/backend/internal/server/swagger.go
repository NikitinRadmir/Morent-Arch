package server

import (
	"net/http"
)

// SwaggerJSON возвращает минимальное OpenAPI-описание API в формате JSON.
// Этого достаточно, чтобы подключить внешний Swagger UI.
func SwaggerJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
  "openapi": "3.0.0",
  "info": {
    "title": "Morent API",
    "version": "1.0.0",
    "description": "Minimal OpenAPI spec for Morent car rental backend"
  },
  "paths": {
    "/Cars/GetAll": {
      "get": {
        "summary": "List cars",
        "responses": { "200": { "description": "OK" } }
      }
    },
    "/Rentals": {
      "get": {
        "summary": "List user rentals",
        "responses": { "200": { "description": "OK" } }
      },
      "post": {
        "summary": "Create rental",
        "responses": { "201": { "description": "Created" } }
      }
    },
    "/Comments/Create": {
      "post": {
        "summary": "Create comment",
        "responses": { "201": { "description": "Created" } }
      }
    },
    "/Auth/Register": {
      "post": {
        "summary": "Register user",
        "responses": { "201": { "description": "Created" } }
      }
    },
    "/Auth/Login": {
      "post": {
        "summary": "Login user",
        "responses": { "200": { "description": "OK" } }
      }
    }
  }
}`))
}

// SwaggerUI отдаёт простую HTML‑страницу с подключённым Swagger UI (через CDN),
// который использует /swagger.json как источник схемы.
func SwaggerUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Morent API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/swagger.json',
        dom_id: '#swagger-ui',
      });
    };
  </script>
</body>
</html>`))
}


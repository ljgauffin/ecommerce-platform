package http

import (
	"embed"
	"net/http"
)

var swaggerFS embed.FS

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Enterprise E-Commerce API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; padding: 0; background: #fafafa; font-family: sans-serif; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

type SwaggerHandler struct{}

func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

func (h *SwaggerHandler) UI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(swaggerHTML))
}

func (h *SwaggerHandler) Doc(w http.ResponseWriter, r *http.Request) {
	data, err := swaggerFS.ReadFile("swagger/openapi.json")
	if err != nil {
		http.Error(w, "Failed to load openapi specification", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

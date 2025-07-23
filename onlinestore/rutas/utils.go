package rutas

import (
    
    "encoding/json"
    "net/http"
)

// Claims representa la estructura del token JWT
type Claims struct {
    UserID  int  `json:"user_id"`
    IsAdmin bool `json:"is_admin"`
}

// RespondWithError envía una respuesta de error al cliente
func RespondWithError(w http.ResponseWriter, code int, message string) {
    RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON envía una respuesta JSON al cliente
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, _ := json.Marshal(payload)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}

// ValidateToken valida el token JWT y retorna los claims
func ValidateToken(r *http.Request) (*Claims, error) {
    // Implementa la validación del token según tu sistema de autenticación
    // Este es solo un ejemplo
    return &Claims{
        UserID:  1,
        IsAdmin: true,
    }, nil
}
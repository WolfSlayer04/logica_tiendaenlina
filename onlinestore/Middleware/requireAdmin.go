package middlewares

import (
    "net/http"
)

// Requiere que el usuario sea admin (acceso total)
func RequireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tipoUsuario := r.Header.Get("X-Tipo-Usuario") // O donde guardes el tipo
        if tipoUsuario != "admin" {
            http.Error(w, "No tienes permisos suficientes", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
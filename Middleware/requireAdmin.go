package middlewares

import (
    "net/http"
)

// Requiere que el usuario sea admin (acceso total)
func RequireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // SIEMPRE agrega los headers de CORS
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, X-Tipo-Usuario")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        tipoUsuario := r.Header.Get("X-Tipo-Usuario") // O donde guardes el tipo
        if tipoUsuario != "A" {
            http.Error(w, "No tienes permisos suficientes", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}


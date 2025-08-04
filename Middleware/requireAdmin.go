package middlewares

import (
    "net/http"
)


// RequireAdmin checks that the user is admin using JWT claims from context
func RequireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, X-Tipo-Usuario")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        // Get tipo from context (set by JWT middleware)
        tipoUsuario, ok := r.Context().Value(ContextTipoKey).(string)
        if !ok || tipoUsuario != "A" {
            http.Error(w, "No tienes permisos suficientes", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
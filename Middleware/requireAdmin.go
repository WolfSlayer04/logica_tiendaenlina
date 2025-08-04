package middlewares

import (
    "net/http"
)


// RequireAdminPermisos: solo permite admins con los permisos correctos
func RequireAdminPermisos(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, X-Tipo-Usuario")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        // Leer tipo_usuario y permisos del contexto (puestos por JWT middleware)
        tipoUsuario, tipoOk := r.Context().Value(ContextTipoKey).(string)
        soloLectura, _ := r.Context().Value(ContextSoloLecturaKey).(bool)
        accesoTotal, _ := r.Context().Value(ContextAccesoTotalKey).(bool)

        // Solo admins pueden pasar
        if !tipoOk || tipoUsuario != "A" {
            http.Error(w, "No tienes permisos suficientes", http.StatusForbidden)
            return
        }

        // Si tiene acceso total, puede cualquier método
        if accesoTotal {
            next.ServeHTTP(w, r)
            return
        }

        // Si tiene soloLectura, solo puede GET/OPTIONS/HEAD
        if soloLectura {
            if r.Method == http.MethodGet || r.Method == http.MethodOptions || r.Method == http.MethodHead {
                next.ServeHTTP(w, r)
                return
            }
            http.Error(w, "Tu usuario es solo lectura, no puedes editar.", http.StatusForbidden)
            return
        }

        // Si no tiene ninguno de los dos flags, bloquear (por seguridad)
        http.Error(w, "No tienes permisos suficientes (falta accesoTotal o soloLectura)", http.StatusForbidden)
    })
}
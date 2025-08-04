package middlewares

import (
    "net/http"
    "strings"
    "github.com/golang-jwt/jwt/v5"
    "context"
    "fmt"
)

var jwtKey = []byte("TU_CLAVE_SECRETA")

type Claims struct {
    ID       int               `json:"id"`
    Tipo     string            `json:"tipo"` // 'C' o 'A'
    Correo   string            `json:"correo"`
    Permisos map[string]bool   `json:"permisos"` // ← NUEVO
    jwt.RegisteredClaims
}

// Claves para contexto
type contextKey string

const (
    ContextUserIDKey     contextKey = "userID"
    ContextTipoKey       contextKey = "tipo"
    ContextCorreoKey     contextKey = "correo"
    ContextSoloLecturaKey contextKey = "solo_lectura"
    ContextAccesoTotalKey contextKey = "acceso_total"
)

// Middleware JWT para rutas protegidas
func JWTAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, X-Tipo-Usuario")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        authHeader := r.Header.Get("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            http.Error(w, "Token faltante o inválido", http.StatusUnauthorized)
            return
        }
        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

        claims := &Claims{}
        token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })
        if err != nil || !token.Valid {
            http.Error(w, "Token inválido o expirado", http.StatusUnauthorized)
            return
        }

        if claims.Tipo != "C" && claims.Tipo != "A" {
            http.Error(w, "Tipo de usuario no permitido", http.StatusForbidden)
            return
        }

        // Extraer flags de permisos seguros del JWT
        soloLectura := false
        accesoTotal := false
        if claims.Permisos != nil {
            if v, ok := claims.Permisos["soloLectura"]; ok {
                soloLectura = v
            }
            if v, ok := claims.Permisos["accesoTotal"]; ok {
                accesoTotal = v
            }
        }

        ctx := context.WithValue(r.Context(), ContextUserIDKey, claims.ID)
        ctx = context.WithValue(ctx, ContextTipoKey, claims.Tipo)
        ctx = context.WithValue(ctx, ContextCorreoKey, claims.Correo)
        ctx = context.WithValue(ctx, ContextSoloLecturaKey, soloLectura)
        ctx = context.WithValue(ctx, ContextAccesoTotalKey, accesoTotal)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Recuperar datos del usuario autenticado desde el contexto
func GetUserFromContext(r *http.Request) (id int, tipo string, correo string, soloLectura bool, accesoTotal bool) {
    if val := r.Context().Value(ContextUserIDKey); val != nil {
        id, _ = val.(int)
    }
    if val := r.Context().Value(ContextTipoKey); val != nil {
        tipo, _ = val.(string)
    }
    if val := r.Context().Value(ContextCorreoKey); val != nil {
        correo, _ = val.(string)
    }
    if val := r.Context().Value(ContextSoloLecturaKey); val != nil {
        soloLectura, _ = val.(bool)
    }
    if val := r.Context().Value(ContextAccesoTotalKey); val != nil {
        accesoTotal, _ = val.(bool)
    }
    return
}

// Ejemplo de uso en tu handler
func RutaProtegidaHandler(w http.ResponseWriter, r *http.Request) {
    id, tipo, correo, soloLectura, accesoTotal := GetUserFromContext(r)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf(`{"success":true,"id":%d,"tipo":"%s","correo":"%s","soloLectura":%t,"accesoTotal":%t}`, id, tipo, correo, soloLectura, accesoTotal)))
}
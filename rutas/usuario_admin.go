package rutas

import (
    "encoding/json"
    "net/http"
    "strconv"
    "log"

    "github.com/WolfSlayer04/logica_tiendaenlina/db"
)

// Modelo para admin_usuarios (lectura)
type AdminUsuarioUser struct {
    IDUsuario   int             `json:"idusuario"`
    IDPerfil    int             `json:"idperfil"`
    Permisos    json.RawMessage `json:"permisos"`
    TipoUsuario string          `json:"tipo_usuario"`
    Correo      string          `json:"correo"`
}

// Modelo para crear admin_usuarios (debe coincidir con el frontend)
type AdminUsuarioCreate struct {
    IDPerfil    int    `json:"idperfil"`
    Permisos    string `json:"permisos"`    // Como string, no RawMessage
    TipoUsuario string `json:"tipo_usuario"`
    Correo      string `json:"correo"`
    Clave       string `json:"clave"`
}

// ===================
// Listar administradores
// ===================
func GetAllAdminUsuarios(dbConn *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        rows, err := dbConn.Local.Query("SELECT idusuario, idperfil, permisos, tipo_usuario, correo FROM admin_usuarios")
        if err != nil {
            log.Printf("Error obteniendo administradores: %v", err)
            http.Error(w, "Error obteniendo administradores", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var admins []AdminUsuario
        for rows.Next() {
            var a AdminUsuario
            var permisosStr string
            if err := rows.Scan(&a.IDUsuario, &a.IDPerfil, &permisosStr, &a.TipoUsuario, &a.Correo); err != nil {
                log.Printf("Error escaneando admin: %v", err)
                http.Error(w, "Error escaneando admin", http.StatusInternalServerError)
                return
            }
            a.Permisos = json.RawMessage(permisosStr)
            admins = append(admins, a)
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "data": admins})
    }
}

// ===================
// Crear nuevo admin
// ===================
func CreateAdminUsuario(dbConn *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var nuevo AdminUsuarioCreate
        if err := json.NewDecoder(r.Body).Decode(&nuevo); err != nil {
            log.Printf("Error decodificando JSON: %v", err)
            http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
            return
        }

        // Log para debug
        log.Printf("Datos recibidos: %+v", nuevo)

        // Validación básica
        if nuevo.Correo == "" {
            http.Error(w, "El correo es obligatorio", http.StatusBadRequest)
            return
        }
        if nuevo.Clave == "" {
            http.Error(w, "La contraseña es obligatoria", http.StatusBadRequest)
            return
        }
        if nuevo.TipoUsuario == "" {
            http.Error(w, "El tipo de usuario es obligatorio", http.StatusBadRequest)
            return
        }
        if nuevo.Permisos == "" {
            http.Error(w, "Los permisos son obligatorios", http.StatusBadRequest)
            return
        }

        // Validar que permisos sea JSON válido
        var permisosTest interface{}
        if err := json.Unmarshal([]byte(nuevo.Permisos), &permisosTest); err != nil {
            http.Error(w, "Los permisos deben ser JSON válido", http.StatusBadRequest)
            return
        }

        // Si idperfil no se envía, usar valor por defecto
        if nuevo.IDPerfil == 0 {
            nuevo.IDPerfil = 1
        }

        // Validar correo único
        var existe int
        err := dbConn.Local.QueryRow("SELECT COUNT(*) FROM admin_usuarios WHERE correo = ?", nuevo.Correo).Scan(&existe)
        if err != nil {
            log.Printf("Error validando correo: %v", err)
            http.Error(w, "Error validando correo", http.StatusInternalServerError)
            return
        }
        if existe > 0 {
            http.Error(w, "El correo ya está registrado", http.StatusBadRequest)
            return
        }

        // Insertar en base de datos
        res, err := dbConn.Local.Exec(
            "INSERT INTO admin_usuarios (idperfil, permisos, tipo_usuario, correo, clave) VALUES (?, ?, ?, ?, ?)",
            nuevo.IDPerfil, nuevo.Permisos, nuevo.TipoUsuario, nuevo.Correo, nuevo.Clave,
        )
        if err != nil {
            log.Printf("Error creando admin: %v", err)
            http.Error(w, "Error al crear admin: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        id, _ := res.LastInsertId()
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "ok": true, 
            "idusuario": id,
            "message": "Administrador creado exitosamente",
        })
    }
}

// ===================
// Editar admin
// ===================
func UpdateAdminUsuario(dbConn *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idStr := r.URL.Query().Get("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            http.Error(w, "ID inválido", http.StatusBadRequest)
            return
        }
        
        var upd AdminUsuarioCreate
        if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
            log.Printf("Error decodificando JSON para actualización: %v", err)
            http.Error(w, "JSON inválido", http.StatusBadRequest)
            return
        }

        // Validar que permisos sea JSON válido si se proporciona
        if upd.Permisos != "" {
            var permisosTest interface{}
            if err := json.Unmarshal([]byte(upd.Permisos), &permisosTest); err != nil {
                http.Error(w, "Los permisos deben ser JSON válido", http.StatusBadRequest)
                return
            }
        }

        _, err = dbConn.Local.Exec(
            "UPDATE admin_usuarios SET idperfil=?, permisos=?, tipo_usuario=?, correo=? WHERE idusuario=?",
            upd.IDPerfil, upd.Permisos, upd.TipoUsuario, upd.Correo, id,
        )
        if err != nil {
            log.Printf("Error actualizando admin: %v", err)
            http.Error(w, "Error actualizando admin", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "ok": true,
            "message": "Administrador actualizado exitosamente",
        })
    }
}

// ===================
// Eliminar admin
// ===================
func DeleteAdminUsuario(dbConn *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idStr := r.URL.Query().Get("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            http.Error(w, "ID inválido", http.StatusBadRequest)
            return
        }

        // No permitir borrar al admin principal
        if id == 1 {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok": false,
                "message": "No se puede eliminar el administrador principal.",
            })
            return
        }
        
        _, err = dbConn.Local.Exec("DELETE FROM admin_usuarios WHERE idusuario=?", id)
        if err != nil {
            log.Printf("Error eliminando admin: %v", err)
            http.Error(w, "Error eliminando admin", http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "ok": true,
            "message": "Administrador eliminado exitosamente",
        })
    }
}
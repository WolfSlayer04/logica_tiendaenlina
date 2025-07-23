package rutas

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "github.com/WolfSlayer04/logica_tiendaenlina/db"
    "golang.org/x/crypto/bcrypt" // Agregamos bcrypt
)

type LoginRequest struct {
    Correo string `json:"correo"`
    Clave  string `json:"clave"`
}

func writeSuccessResponse1(w http.ResponseWriter, message string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": message,
        "data":    data,
    })
}

func writeErrorResponse1(w http.ResponseWriter, status int, message string, details string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    resp := map[string]interface{}{
        "success": false,
        "message": message,
    }
    if details != "" {
        resp["details"] = details
    }
    json.NewEncoder(w).Encode(resp)
}

type AdminUsuario struct {
    IDUsuario   int             `json:"id_usuario"`
    IDPerfil    int             `json:"id_perfil"`
    Permisos    json.RawMessage `json:"permisos"`
    TipoUsuario string          `json:"tipo_usuario"`
    Correo      string          `json:"correo"`
    Clave       string          `json:"-"`
}

// Función auxiliar para verificar contraseñas
func verificarContraseña(claveHash string, claveIntento string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(claveHash), []byte(claveIntento))
    return err == nil
}

func LoginUsuario(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req LoginRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            writeErrorResponse1(w, http.StatusBadRequest, "Datos inválidos", err.Error())
            return
        }

        // 1. Intentar login como usuario normal
        var u db.Usuario
        query := `SELECT id_usuario, id_empresa, tipo_usuario, nombre_completo, correo, telefono, clave, estatus FROM usuarios WHERE correo = ? LIMIT 1`
        err := dbc.Local.QueryRow(query, req.Correo).Scan(
            &u.IDUsuario, &u.IDEmpresa, &u.TipoUsuario, &u.NombreCompleto, &u.Correo, &u.Telefono, &u.Clave, &u.Estatus,
        )
        if err == sql.ErrNoRows {
            // 2. Intentar login como admin
            var admin AdminUsuario
            adminQuery := `SELECT idusuario, idperfil, permisos, tipo_usuario, correo, clave FROM admin_usuarios WHERE correo = ? LIMIT 1`
            adminErr := dbc.Local.QueryRow(adminQuery, req.Correo).Scan(
                &admin.IDUsuario, &admin.IDPerfil, &admin.Permisos, &admin.TipoUsuario, &admin.Correo, &admin.Clave,
            )
            if adminErr == sql.ErrNoRows {
                writeErrorResponse1(w, http.StatusUnauthorized, "Usuario o contraseña incorrectos", "")
                return
            } else if adminErr != nil {
                writeErrorResponse1(w, http.StatusInternalServerError, "Error de base de datos (admin)", adminErr.Error())
                return
            }

            // Validar clave admin (sin encriptar)
            if admin.Clave != req.Clave {
                writeErrorResponse1(w, http.StatusUnauthorized, "Usuario o contraseña incorrectos", "")
                return
            }
            
            admin.Clave = ""
            writeSuccessResponse1(w, "Login exitoso (admin)", map[string]interface{}{
                "admin": admin,
            })
            return
        } else if err != nil {
            writeErrorResponse1(w, http.StatusInternalServerError, "Error de base de datos", err.Error())
            return
        }

        // Validación de contraseña para usuario normal usando bcrypt
        if !verificarContraseña(u.Clave, req.Clave) {
            writeErrorResponse1(w, http.StatusUnauthorized, "Usuario o contraseña incorrectos", "")
            return
        }

        if u.Estatus != "activo" {
            writeErrorResponse1(w, http.StatusUnauthorized, "Usuario inactivo o suspendido", "")
            return
        }
        u.Clave = "" // no enviar clave al frontend

        // Consultar datos de tienda asociada al usuario (si existe), incluyendo geolocalización
        var tienda struct {
            IDTienda     int     `json:"id_tienda"`
            NombreTienda string  `json:"nombre_tienda"`
            Direccion    string  `json:"direccion"`
            Colonia      string  `json:"colonia"`
            CodigoPostal string  `json:"codigo_postal"`
            Ciudad       string  `json:"ciudad"`
            Estado       string  `json:"estado"`
            Pais         string  `json:"pais"`
            Latitud      float64 `json:"latitud"`
            Longitud     float64 `json:"longitud"`
            LatitudUbic  float64 `json:"latitud_ubic"`
            LongitudUbic float64 `json:"longitud_ubic"`
        }
        var lat, lon, latPoint, lonPoint sql.NullFloat64
        err = dbc.Local.QueryRow(`
            SELECT id_tienda, nombre_tienda, direccion, colonia, codigo_postal, ciudad, estado, pais,
                   latitud, longitud, IFNULL(ST_Y(ubicacion), 0) as latitud_ubic, IFNULL(ST_X(ubicacion), 0) as longitud_ubic
            FROM tiendas
            WHERE id_usuario = ?
            LIMIT 1
        `, u.IDUsuario).Scan(
            &tienda.IDTienda, &tienda.NombreTienda, &tienda.Direccion, &tienda.Colonia,
            &tienda.CodigoPostal, &tienda.Ciudad, &tienda.Estado, &tienda.Pais,
            &lat, &lon, &latPoint, &lonPoint,
        )
        if err != nil && err != sql.ErrNoRows {
            writeErrorResponse1(w, http.StatusInternalServerError, "Error al consultar la tienda revisa de nuevo", err.Error())
            return
        }
        tienda.Latitud = db.NullToFloat(lat)
        tienda.Longitud = db.NullToFloat(lon)
        tienda.LatitudUbic = db.NullToFloat(latPoint)
        tienda.LongitudUbic = db.NullToFloat(lonPoint)

        // Devuelve ambos objetos (usuario y tienda) al frontend
        writeSuccessResponse1(w, "Login exitoso", map[string]interface{}{
            "usuario": u,
            "tienda":  tienda,
        })
    }
}
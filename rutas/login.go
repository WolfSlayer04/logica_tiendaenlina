package rutas

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "github.com/WolfSlayer04/logica_tiendaenlina/db"
    "golang.org/x/crypto/bcrypt"
    "github.com/golang-jwt/jwt/v5"
    "crypto/rand"
    "encoding/base64"
    "time"
)

var jwtKey = []byte("TU_CLAVE_SECRETA") // Reemplaza por una clave segura

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

// Verifica contraseña (hash y texto plano para compatibilidad)
func verificarContraseña(claveAlmacenada, claveIntento string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(claveAlmacenada), []byte(claveIntento))
    if err == nil {
        return true
    }
    return claveAlmacenada == claveIntento
}

// Genera Access y Refresh Token usando zona horaria de Yucatán
func generarTokens(id int, tipo string, correo string) (string, string, error) {
    loc, err := time.LoadLocation("America/Merida")
    if err != nil {
        loc = time.UTC
    }
    now := time.Now().In(loc)
    midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)

    claims := jwt.MapClaims{
        "id":    id,
        "tipo":  tipo,
        "correo": correo,
        "exp":   midnight.Unix(),
    }
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    accessString, err := accessToken.SignedString(jwtKey)
    if err != nil {
        return "", "", err
    }
    refreshBytes := make([]byte, 32)
    _, err = rand.Read(refreshBytes)
    if err != nil {
        return "", "", err
    }
    refreshString := base64.URLEncoding.EncodeToString(refreshBytes)
    return accessString, refreshString, nil
}

// Guarda el refresh token, asegurando que las fechas sean hora local en string
func guardarRefreshToken(dbc *db.DBConnection, userID int, tipoUsuario string, refreshToken, userAgent, ip string) error {
    loc, err := time.LoadLocation("America/Merida")
    if err != nil {
        loc = time.UTC
    }
    now := time.Now().In(loc)
    midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
    nowStr := now.Format("2006-01-02 15:04:05")
    midnightStr := midnight.Format("2006-01-02 15:04:05")

    tipo := tipoUsuario
    if tipo != "A" && tipo != "C" {
        if tipo == "admin" {
            tipo = "A"
        } else {
            tipo = "C"
        }
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    _, err = dbc.Local.Exec(`
        INSERT INTO refresh_tokens (usuario_id, tipo_usuario, token_hash, user_agent, ip_address, expiracion, ultimo_uso, estado)
        VALUES (?, ?, ?, ?, ?, ?, ?, 'activo')
    `, userID, tipo, string(hash), userAgent, ip, midnightStr, nowStr)
    return err
}

// Corregido: lee ultimo_uso como string y lo convierte a time.Time
func sesionActivaReciente(dbc *db.DBConnection, userID int) (bool, error) {
    var ultimoUsoStr sql.NullString
    err := dbc.Local.QueryRow(`
        SELECT ultimo_uso FROM refresh_tokens
        WHERE usuario_id = ? AND estado = 'activo'
        ORDER BY expiracion DESC LIMIT 1
    `, userID).Scan(&ultimoUsoStr)
    if err == sql.ErrNoRows {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    if !ultimoUsoStr.Valid || ultimoUsoStr.String == "" {
        return false, nil
    }
    // Convierte string a time.Time
    ultimoUso, err := time.Parse("2006-01-02 15:04:05", ultimoUsoStr.String)
    if err != nil {
        return false, err
    }
    if time.Since(ultimoUso) < 15*time.Minute {
        return true, nil
    }
    return false, nil
}

func LoginUsuario(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req LoginRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            writeErrorResponse1(w, http.StatusBadRequest, "Datos inválidos", err.Error())
            return
        }

        var u db.Usuario
        query := `SELECT id_usuario, id_empresa, tipo_usuario, nombre_completo, correo, telefono, clave, estatus FROM usuarios WHERE correo = ? LIMIT 1`
        err := dbc.Local.QueryRow(query, req.Correo).Scan(
            &u.IDUsuario, &u.IDEmpresa, &u.TipoUsuario, &u.NombreCompleto, &u.Correo, &u.Telefono, &u.Clave, &u.Estatus,
        )
        var tipoUsuario string
        var admin AdminUsuario

        if err == sql.ErrNoRows {
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
            if !verificarContraseña(admin.Clave, req.Clave) {
                writeErrorResponse1(w, http.StatusUnauthorized, "Usuario o contraseña incorrectos", "")
                return
            }
            activo, err := sesionActivaReciente(dbc, admin.IDUsuario)
            if err != nil {
                writeErrorResponse1(w, http.StatusInternalServerError, "Error verificando sesión activa", err.Error())
                return
            }
            if activo {
                writeErrorResponse1(w, http.StatusConflict, "Sesión iniciada en otro equipo, espere o cierre sesión", "")
                return
            }
            admin.TipoUsuario = "A"
            accessToken, refreshToken, err := generarTokens(admin.IDUsuario, admin.TipoUsuario, admin.Correo)
            if err != nil {
                writeErrorResponse1(w, http.StatusInternalServerError, "Error generando tokens", err.Error())
                return
            }
            userAgent := r.Header.Get("User-Agent")
            ip := r.RemoteAddr
            err = guardarRefreshToken(dbc, admin.IDUsuario, admin.TipoUsuario, refreshToken, userAgent, ip)
            if err != nil {
                writeErrorResponse1(w, http.StatusInternalServerError, "Error guardando refresh token", err.Error())
                return
            }
            admin.Clave = ""
            writeSuccessResponse1(w, "Login exitoso (admin)", map[string]interface{}{
                "admin": admin,
                "access_token": accessToken,
            })
            return
        } else if err != nil {
            writeErrorResponse1(w, http.StatusInternalServerError, "Error de base de datos", err.Error())
            return
        }

        if !verificarContraseña(u.Clave, req.Clave) {
            writeErrorResponse1(w, http.StatusUnauthorized, "Usuario o contraseña incorrectos", "")
            return
        }

        if u.Estatus != "activo" {
            writeErrorResponse1(w, http.StatusUnauthorized, "Usuario inactivo o suspendido", "")
            return
        }
        u.Clave = ""
        tipoUsuario = "C"

        activo, err := sesionActivaReciente(dbc, u.IDUsuario)
        if err != nil {
            writeErrorResponse1(w, http.StatusInternalServerError, "Error verificando sesión activa", err.Error())
            return
        }
        if activo {
            writeErrorResponse1(w, http.StatusConflict, "Sesión iniciada en otro equipo, espere o cierre sesión", "")
            return
        }

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

        accessToken, refreshToken, err := generarTokens(u.IDUsuario, tipoUsuario, u.Correo)
        if err != nil {
            writeErrorResponse1(w, http.StatusInternalServerError, "Error generando tokens", err.Error())
            return
        }
        userAgent := r.Header.Get("User-Agent")
        ip := r.RemoteAddr
        err = guardarRefreshToken(dbc, u.IDUsuario, tipoUsuario, refreshToken, userAgent, ip)
        if err != nil {
            writeErrorResponse1(w, http.StatusInternalServerError, "Error guardando refresh token", err.Error())
            return
        }

        writeSuccessResponse1(w, "Login exitoso", map[string]interface{}{
            "usuario": u,
            "tienda":  tienda,
            "access_token": accessToken,
        })
    }
}
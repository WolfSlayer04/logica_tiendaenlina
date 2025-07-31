package rutas

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/WolfSlayer04/logica_tiendaenlina/db"
	"golang.org/x/crypto/bcrypt"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Genera Access y Refresh Token usando zona horaria de Yucatán

// Valida refresh token
func validarRefreshToken(dbc *db.DBConnection, refreshToken string) (tokenID int, userID int, tipoUsuario, correo string, err error) {
	var (
		id            int
		tokenId       int
		tipo          string
		correoUsuario string
		tokenHash     string
		expiracionStr sql.NullString
	)
	query := `SELECT id, usuario_id, tipo_usuario, token_hash, expiracion FROM refresh_tokens WHERE estado = 'activo'`
	rows, err := dbc.Local.Query(query)
	if err != nil {
		return 0, 0, "", "", err
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		err = rows.Scan(&tokenId, &id, &tipo, &tokenHash, &expiracionStr)
		if err != nil {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(tokenHash), []byte(refreshToken)) == nil {
			found = true
			break
		}
	}
	if !found {
		return 0, 0, "", "", sql.ErrNoRows
	}
	if !expiracionStr.Valid || expiracionStr.String == "" {
		return 0, 0, "", "", sql.ErrNoRows
	}
	expiracion, err := time.Parse("2006-01-02 15:04:05", expiracionStr.String)
	if err != nil {
		return 0, 0, "", "", err
	}
	if time.Now().After(expiracion) {
		return 0, 0, "", "", sql.ErrNoRows
	}
	if tipo == "A" {
		err = dbc.Local.QueryRow(`SELECT correo FROM admin_usuarios WHERE idusuario = ?`, id).Scan(&correoUsuario)
	} else {
		err = dbc.Local.QueryRow(`SELECT correo FROM usuarios WHERE id_usuario = ?`, id).Scan(&correoUsuario)
	}
	if err != nil {
		return 0, 0, "", "", err
	}
	return tokenId, id, tipo, correoUsuario, nil
}

func RefreshTokenEndpoint(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse1(w, http.StatusBadRequest, "Datos inválidos", err.Error())
			return
		}
		if req.RefreshToken == "" {
			writeErrorResponse1(w, http.StatusBadRequest, "Falta refresh_token", "")
			return
		}

		tokenID, userID, tipoUsuario, correo, err := validarRefreshToken(dbc, req.RefreshToken)
		if err != nil {
			writeErrorResponse1(w, http.StatusUnauthorized, "Refresh token inválido o expirado", err.Error())
			return
		}

		loc, err := time.LoadLocation("America/Merida")
		if err != nil {
			loc = time.UTC
		}

		// Verifica sesión activa reciente antes de renovar (lee como string y convierte)
		var ultimoUsoStr sql.NullString
		err = dbc.Local.QueryRow(`
            SELECT ultimo_uso FROM refresh_tokens
            WHERE id = ? AND estado = 'activo'
            LIMIT 1
        `, tokenID).Scan(&ultimoUsoStr)
		if err == nil && ultimoUsoStr.Valid && ultimoUsoStr.String != "" {
			ultimoUso, err := time.Parse("2006-01-02 15:04:05", ultimoUsoStr.String)
			if err == nil && time.Since(ultimoUso) < 15*time.Minute {
				writeErrorResponse1(w, http.StatusConflict, "Sesión iniciada en otro equipo, espere o cierre sesión", "")
				return
			}
		}

		now := time.Now().In(loc)
		midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
		nowStr := now.Format("2006-01-02 15:04:05")
		midnightStr := midnight.Format("2006-01-02 15:04:05")

		accessToken, newRefreshToken, err := generarTokens(userID, tipoUsuario, correo)
		if err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error generando tokens", err.Error())
			return
		}
		userAgent := r.Header.Get("User-Agent")
		ip := r.RemoteAddr
		hash, err := bcrypt.GenerateFromPassword([]byte(newRefreshToken), bcrypt.DefaultCost)
		if err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error generando hash de refresh token", err.Error())
			return
		}

		_, err = dbc.Local.Exec(`UPDATE refresh_tokens SET estado = 'revocado' WHERE id = ?`, tokenID)
		if err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error revocando refresh token anterior", err.Error())
			return
		}
		_, err = dbc.Local.Exec(`
            INSERT INTO refresh_tokens (usuario_id, tipo_usuario, token_hash, user_agent, ip_address, expiracion, ultimo_uso, estado)
            VALUES (?, ?, ?, ?, ?, ?, ?, 'activo')`,
			userID, tipoUsuario, string(hash), userAgent, ip, midnightStr, nowStr)
		if err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error guardando nuevo refresh token", err.Error())
			return
		}

		writeSuccessResponse1(w, "Tokens renovados correctamente", map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
		})
	}
}

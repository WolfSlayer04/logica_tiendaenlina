package rutas

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/WolfSlayer04/logica_tiendaenlina/db"
)

// ConfigEntrega representa la estructura JSON de configuración de entregas
type ConfigEntrega struct {
	DiasHabiles        []string `json:"dias_habiles"`
	TiempoProcesamiento int     `json:"tiempo_procesamiento"`
	ReglasFindeSemana  struct {
		ProcesarSabado       bool `json:"procesar_sabado"`
		ProcesarDomingo      bool `json:"procesar_domingo"`
		DiasAdicionalesSabado int  `json:"dias_adicionales_sabado"`
		DiasAdicionalesDomingo int `json:"dias_adicionales_domingo"`
	} `json:"reglas_fin_semana"`
	HorariosEntrega []struct {
		Etiqueta string `json:"etiqueta"`
		Inicio   string `json:"inicio"`
		Fin      string `json:"fin"`
	} `json:"horarios_entrega"`
	DiasFeriados []struct {
		Fecha          string `json:"fecha"`
		DiasAdicionales int   `json:"dias_adicionales"`
	} `json:"dias_feriados"`
}

// GetConfigEntrega obtiene la configuración de entregas del administrador
func GetConfigEntrega(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// En Postman: Para probar temporalmente, envía el ID admin como parámetro
		idAdminStr := r.URL.Query().Get("id_admin")
		if idAdminStr == "" {
			writeErrorResponse1(w, http.StatusBadRequest, "Falta el ID del administrador", "Para pruebas, usa ?id_admin=X")
			return
		}
		
		var idAdmin int
		fmt.Sscanf(idAdminStr, "%d", &idAdmin)
		
		// Verificar que el usuario existe y es admin
		var tipoUsuario string
		query := `SELECT tipo_usuario FROM admin_usuarios WHERE idusuario = ?`
		err := dbConn.Local.QueryRow(query, idAdmin).Scan(&tipoUsuario)
		if err == sql.ErrNoRows {
			writeErrorResponse1(w, http.StatusUnauthorized, "Administrador no encontrado", "")
			return
		} else if err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error de base de datos", err.Error())
			return
		}
		
		if tipoUsuario != "Admin" {
			writeErrorResponse1(w, http.StatusForbidden, "El usuario no tiene permisos de administrador", "")
			return
		}

		// Consultar la configuración de entregas
		query = `SELECT config_entrega FROM admin_usuarios WHERE idusuario = ?`
		var configJSON []byte
		err = dbConn.Local.QueryRow(query, idAdmin).Scan(&configJSON)
		if err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error al obtener configuración de entregas", err.Error())
			return
		}

		// Si no hay configuración, devolver un objeto vacío
		if len(configJSON) == 0 {
			writeSuccessResponse1(w, "No hay configuración de entregas", map[string]interface{}{
				"config": ConfigEntrega{},
			})
			return
		}

		// Devolver la configuración
		var config map[string]interface{}
		if err := json.Unmarshal(configJSON, &config); err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error al procesar configuración", err.Error())
			return
		}
		
		writeSuccessResponse1(w, "Configuración de entregas obtenida", map[string]interface{}{
			"config": config,
		})
	}
}

// UpdateConfigEntrega actualiza la configuración de entregas del administrador
func UpdateConfigEntrega(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// En Postman: Para probar temporalmente, envía el ID admin como parámetro
		idAdminStr := r.URL.Query().Get("id_admin")
		if idAdminStr == "" {
			writeErrorResponse1(w, http.StatusBadRequest, "Falta el ID del administrador", "Para pruebas, usa ?id_admin=X")
			return
		}
		
		var idAdmin int
		fmt.Sscanf(idAdminStr, "%d", &idAdmin)
		
		// Verificar que el usuario existe y es admin
		var tipoUsuario string
		query := `SELECT tipo_usuario FROM admin_usuarios WHERE idusuario = ?`
		err := dbConn.Local.QueryRow(query, idAdmin).Scan(&tipoUsuario)
		if err == sql.ErrNoRows {
			writeErrorResponse1(w, http.StatusUnauthorized, "Administrador no encontrado", "")
			return
		} else if err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error de base de datos", err.Error())
			return
		}
		
		if tipoUsuario != "Admin" {
			writeErrorResponse1(w, http.StatusForbidden, "El usuario no tiene permisos de administrador", "")
			return
		}

		// Decodificar la configuración de entregas
		var config ConfigEntrega
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			writeErrorResponse1(w, http.StatusBadRequest, "Error al decodificar la configuración", err.Error())
			return
		}

		// Validar la configuración
		if len(config.DiasHabiles) == 0 {
			writeErrorResponse1(w, http.StatusBadRequest, "Debe especificar al menos un día hábil", "")
			return
		}

		// Convertir a JSON
		configJSON, err := json.Marshal(config)
		if err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error al serializar la configuración", err.Error())
			return
		}

		// Actualizar en la base de datos
		query = `UPDATE admin_usuarios SET config_entrega = ? WHERE idusuario = ?`
		_, err = dbConn.Local.Exec(query, configJSON, idAdmin)
		if err != nil {
			writeErrorResponse1(w, http.StatusInternalServerError, "Error al actualizar la configuración", err.Error())
			return
		}

		// Responder con éxito
		writeSuccessResponse1(w, "Configuración de entregas actualizada correctamente", nil)
	}
}
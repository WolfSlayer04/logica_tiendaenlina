package rutas

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/WolfSlayer04/logica_tiendaenlina/db"
	"github.com/gorilla/mux"
)

type EmpresaConfigVisual struct {
	ID        int       `json:"id"`
	IDEmpresa int       `json:"idempresa"`
	Config    string    `json:"config"`
	UpdatedAt time.Time `json:"updated_at"`
}

func getUniqueIDEmpresa(dbConn *db.DBConnection) (int, error) {
	var idempresa int
	err := dbConn.Local.QueryRow("SELECT idempresa FROM adm_empresas LIMIT 1").Scan(&idempresa)
	return idempresa, err
}

func AdminGetAllPersonalizaciones(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idempresa, err := getUniqueIDEmpresa(dbConn)
		if err != nil {
			http.Error(w, "No se pudo obtener idempresa: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rows, err := dbConn.Local.Query("SELECT id, idempresa, config, updated_at FROM empresa_config_visual WHERE idempresa = ?", idempresa)
		if err != nil {
			http.Error(w, "Error de base de datos: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var configs []EmpresaConfigVisual
		for rows.Next() {
			var ecv EmpresaConfigVisual
			var updatedAtRaw interface{}
			if err := rows.Scan(&ecv.ID, &ecv.IDEmpresa, &ecv.Config, &updatedAtRaw); err == nil {
				switch v := updatedAtRaw.(type) {
				case time.Time:
					ecv.UpdatedAt = v
				case []uint8:
					t, err := time.Parse("2006-01-02 15:04:05", string(v))
					if err == nil {
						ecv.UpdatedAt = t
					} else {
						ecv.UpdatedAt = time.Time{}
					}
				default:
					ecv.UpdatedAt = time.Time{}
				}
				configs = append(configs, ecv)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configs)
	}
}

func AdminGetPersonalizacionByID(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var ecv EmpresaConfigVisual
		var updatedAtRaw interface{}
		err := dbConn.Local.QueryRow("SELECT id, idempresa, config, updated_at FROM empresa_config_visual WHERE id = ?", id).
			Scan(&ecv.ID, &ecv.IDEmpresa, &ecv.Config, &updatedAtRaw)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "No hay configuración visual con ese id", http.StatusNotFound)
				return
			}
			http.Error(w, "Error de base de datos: "+err.Error(), http.StatusInternalServerError)
			return
		}
		switch v := updatedAtRaw.(type) {
		case time.Time:
			ecv.UpdatedAt = v
		case []uint8:
			t, err := time.Parse("2006-01-02 15:04:05", string(v))
			if err == nil {
				ecv.UpdatedAt = t
			} else {
				ecv.UpdatedAt = time.Time{}
			}
		default:
			ecv.UpdatedAt = time.Time{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ecv)
	}
}

func AdminCreatePersonalizacion(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idempresa, err := getUniqueIDEmpresa(dbConn)
		if err != nil {
			http.Error(w, "No se pudo obtener idempresa: "+err.Error(), http.StatusInternalServerError)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "No se pudo leer el body", http.StatusBadRequest)
			return
		}
		var configData map[string]interface{}
		if err := json.Unmarshal(body, &configData); err != nil {
			http.Error(w, "Body no es JSON válido: "+err.Error(), http.StatusBadRequest)
			return
		}
		configStr := string(body)
		res, err := dbConn.Local.Exec(`
			INSERT INTO empresa_config_visual (idempresa, config, updated_at)
			VALUES (?, ?, NOW())
		`, idempresa, configStr)
		if err != nil {
			http.Error(w, "Error al crear registro: "+err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":         true,
			"idempresa":  idempresa,
			"id":         id,
			"mensaje":    "Registro creado correctamente",
		})
	}
}

func AdminUpdatePersonalizacionByID(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "No se pudo leer el body", http.StatusBadRequest)
			return
		}
		var configData map[string]interface{}
		if err := json.Unmarshal(body, &configData); err != nil {
			http.Error(w, "Body no es JSON válido: "+err.Error(), http.StatusBadRequest)
			return
		}
		configStr := string(body)
		res, err := dbConn.Local.Exec(`
			UPDATE empresa_config_visual
			SET config = ?, updated_at = NOW()
			WHERE id = ?
		`, configStr, id)
		if err != nil {
			http.Error(w, "Error al actualizar registro: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      true,
			"id":      id,
			"rows":    rows,
			"mensaje": "Registro actualizado correctamente",
		})
	}
}

func AdminDeletePersonalizacionByID(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		res, err := dbConn.Local.Exec("DELETE FROM empresa_config_visual WHERE id = ?", id)
		if err != nil {
			http.Error(w, "Error al borrar registro: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      true,
			"id":      id,
			"rows":    rows,
			"mensaje": "Registro borrado correctamente",
		})
	}
}
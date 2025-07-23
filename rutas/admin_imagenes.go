package rutas

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/WolfSlayer04/logica_tiendaenlina/db"
)

// getIdentificadorFromRequest extrae el identificador de logo/logotipo del formulario o query.
func getIdentificadorFromRequest(r *http.Request) (string, error) {
	identificador := r.FormValue("identificador")
	if identificador != "" {
		return identificador, nil
	}
	identificador = r.URL.Query().Get("identificador")
	if identificador != "" {
		return identificador, nil
	}
	return "", http.ErrMissingFile
}

// POST /api/empresa/logo  (requiere "identificador")
func EmpresaUploadLogo(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idempresa, err := getUniqueIDEmpresa(dbConn)
		if err != nil {
			log.Println("[EmpresaUploadLogo] Error obteniendo idempresa:", err)
			http.Error(w, "No se pudo obtener idempresa: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = r.ParseMultipartForm(5 << 20)
		if err != nil {
			log.Println("[EmpresaUploadLogo] Error al leer el formulario:", err)
			http.Error(w, "Error al leer el formulario: "+err.Error(), http.StatusBadRequest)
			return
		}
		identificador, err := getIdentificadorFromRequest(r)
		if err != nil {
			log.Println("[EmpresaUploadLogo] identificador requerido")
			http.Error(w, "Falta el identificador (logo, logotipo, etc)", http.StatusBadRequest)
			return
		}
		// El campo del archivo debe coincidir con el identificador ("logo" o "logotipo")
		file, header, err := r.FormFile(identificador)
		if err != nil {
			log.Printf("[EmpresaUploadLogo] Archivo de %s no recibido: %v", identificador, err)
			http.Error(w, "Archivo de imagen no recibido: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Println("[EmpresaUploadLogo] Error leyendo archivo:", err)
			http.Error(w, "Error leyendo archivo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		mimeType := header.Header.Get("Content-Type")
		if mimeType != "image/png" && mimeType != "image/jpeg" {
			log.Println("[EmpresaUploadLogo] Formato de imagen no permitido:", mimeType)
			http.Error(w, "Formato de imagen no permitido (solo PNG o JPG)", http.StatusBadRequest)
			return
		}
		_, err = dbConn.Local.Exec("DELETE FROM empresa_logos WHERE idempresa = ? AND identificador = ?", idempresa, identificador)
		if err != nil {
			log.Println("[EmpresaUploadLogo] Error borrando imagen anterior:", err)
		}
		_, err = dbConn.Local.Exec(
			"INSERT INTO empresa_logos (idempresa, identificador, imagen, mime_type, updated_at) VALUES (?, ?, ?, ?, NOW())",
			idempresa, identificador, fileBytes, mimeType,
		)
		if err != nil {
			log.Println("[EmpresaUploadLogo] Error guardando imagen nueva:", err)
			http.Error(w, "Error guardando imagen: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}
}

// PUT /api/empresa/logo  (requiere "identificador")
func EmpresaUpdateLogo(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idempresa, err := getUniqueIDEmpresa(dbConn)
		if err != nil {
			log.Println("[EmpresaUpdateLogo] Error obteniendo idempresa:", err)
			http.Error(w, "No se pudo obtener idempresa: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = r.ParseMultipartForm(5 << 20)
		if err != nil {
			log.Println("[EmpresaUpdateLogo] Error al leer el formulario:", err)
			http.Error(w, "Error al leer el formulario: "+err.Error(), http.StatusBadRequest)
			return
		}
		identificador, err := getIdentificadorFromRequest(r)
		if err != nil {
			log.Println("[EmpresaUpdateLogo] identificador requerido")
			http.Error(w, "Falta el identificador (logo, logotipo, etc)", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile(identificador)
		if err != nil {
			log.Printf("[EmpresaUpdateLogo] Archivo de %s no recibido: %v", identificador, err)
			http.Error(w, "Archivo de imagen no recibido: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Println("[EmpresaUpdateLogo] Error leyendo archivo:", err)
			http.Error(w, "Error leyendo archivo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		mimeType := header.Header.Get("Content-Type")
		if mimeType != "image/png" && mimeType != "image/jpeg" {
			log.Println("[EmpresaUpdateLogo] Formato de imagen no permitido:", mimeType)
			http.Error(w, "Formato de imagen no permitido (solo PNG o JPG)", http.StatusBadRequest)
			return
		}
		res, err := dbConn.Local.Exec(
			"UPDATE empresa_logos SET imagen = ?, mime_type = ?, updated_at = NOW() WHERE idempresa = ? AND identificador = ?",
			fileBytes, mimeType, idempresa, identificador,
		)
		if err != nil {
			log.Println("[EmpresaUpdateLogo] Error actualizando imagen:", err)
			http.Error(w, "Error al actualizar registro: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			log.Printf("[EmpresaUpdateLogo] No se encontró registro para actualizar (%s), intenta con POST", identificador)
			http.Error(w, "No se pudo actualizar, intenta con POST", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}
}

// DELETE /api/empresa/logo?identificador=logo
func EmpresaDeleteLogo(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idempresa, err := getUniqueIDEmpresa(dbConn)
		if err != nil {
			log.Println("[EmpresaDeleteLogo] Error obteniendo idempresa:", err)
			http.Error(w, "No se pudo obtener idempresa: "+err.Error(), http.StatusInternalServerError)
			return
		}
		identificador, err := getIdentificadorFromRequest(r)
		if err != nil {
			log.Println("[EmpresaDeleteLogo] identificador requerido")
			http.Error(w, "Falta el identificador (logo, logotipo, etc)", http.StatusBadRequest)
			return
		}
		_, err = dbConn.Local.Exec("DELETE FROM empresa_logos WHERE idempresa = ? AND identificador = ?", idempresa, identificador)
		if err != nil {
			log.Println("[EmpresaDeleteLogo] Error borrando imagen:", err)
			http.Error(w, "Error borrando imagen: "+err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}
}

// GET /api/empresa/logo?identificador=logo
func EmpresaGetLogo(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idempresa, err := getUniqueIDEmpresa(dbConn)
		if err != nil {
			log.Println("[EmpresaGetLogo] Error obteniendo idempresa:", err)
			http.Error(w, "No se pudo obtener idempresa: "+err.Error(), http.StatusInternalServerError)
			return
		}
		identificador, err := getIdentificadorFromRequest(r)
		if err != nil {
			log.Println("[EmpresaGetLogo] identificador requerido")
			http.Error(w, "Falta el identificador (logo, logotipo, etc)", http.StatusBadRequest)
			return
		}
		row := dbConn.Local.QueryRow("SELECT imagen, mime_type, updated_at FROM empresa_logos WHERE idempresa = ? AND identificador = ?", idempresa, identificador)
		var img []byte
		var mime string
		var updatedAtBytes []byte
		err = row.Scan(&img, &mime, &updatedAtBytes)
		if err == sql.ErrNoRows {
			log.Printf("[EmpresaGetLogo] No hay imagen para esta empresa con identificador: %s", identificador)
			http.Error(w, "No hay imagen para esta empresa con ese identificador", http.StatusNotFound)
			return
		} else if err != nil {
			log.Println("[EmpresaGetLogo] Error consultando imagen:", err)
			http.Error(w, "Error consultando imagen: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var updatedAt time.Time
		layouts := []string{
			"2006-01-02 15:04:05",
			time.RFC3339,
			time.RFC1123,
			time.RFC1123Z,
			time.RFC822,
			time.RFC822Z,
		}
		parsed := false
		for _, layout := range layouts {
			t, err := time.Parse(layout, string(updatedAtBytes))
			if err == nil {
				updatedAt = t
				parsed = true
				break
			}
		}
		if !parsed {
			updatedAt = time.Now()
		}
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Last-Modified", updatedAt.UTC().Format(http.TimeFormat))
		_, err = w.Write(img)
		if err != nil {
			log.Println("[EmpresaGetLogo] Error enviando imagen:", err)
		}
	}
}
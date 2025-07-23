package rutas

import (
	"encoding/json"
	"net/http"
	"github.com/WolfSlayer04/logica_tiendaenlina/db"
	
)

// ---------- RESPUESTA PERFIL (usuario + tiendas asociadas + pedidos) ----------

func GetPerfilUsuario(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idUsuario := r.URL.Query().Get("id_usuario")
		if idUsuario == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Falta el id_usuario", "")
			return
		}

		// Info usuario
		var u db.Usuario
		err := dbc.Local.QueryRow(`
			SELECT id_usuario, id_empresa, tipo_usuario, nombre_completo, correo, telefono, estatus
			FROM usuarios WHERE id_usuario = ? LIMIT 1
		`, idUsuario).Scan(&u.IDUsuario, &u.IDEmpresa, &u.TipoUsuario, &u.NombreCompleto, &u.Correo, &u.Telefono, &u.Estatus)
		if err != nil {
			writeErrorResponse(w, http.StatusNotFound, "Usuario no encontrado", err.Error())
			return
		}

		// Tiendas asociadas
		rows, err := dbc.Local.Query(`
			SELECT id_tienda, nombre_tienda, razon_social, rfc, direccion, colonia, codigo_postal, ciudad, estado, pais, tipo_tienda, estatus
			FROM tiendas WHERE id_usuario = ?
		`, idUsuario)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener tiendas", err.Error())
			return
		}
		defer rows.Close()
		var tiendas []map[string]interface{}
		for rows.Next() {
			var t struct {
				IDTienda     int
				NombreTienda string
				RazonSocial  string
				RFC          string
				Direccion    string
				Colonia      string
				CodigoPostal string
				Ciudad       string
				Estado       string
				Pais         string
				TipoTienda   string
				Estatus      string
			}
			err := rows.Scan(&t.IDTienda, &t.NombreTienda, &t.RazonSocial, &t.RFC, &t.Direccion, &t.Colonia, &t.CodigoPostal, &t.Ciudad, &t.Estado, &t.Pais, &t.TipoTienda, &t.Estatus)
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al leer tienda", err.Error())
				return
			}
			tiendas = append(tiendas, map[string]interface{}{
				"id_tienda":      t.IDTienda,
				"nombre_tienda":  t.NombreTienda,
				"razon_social":   t.RazonSocial,
				"rfc":            t.RFC,
				"direccion":      t.Direccion,
				"colonia":        t.Colonia,
				"codigo_postal":  t.CodigoPostal,
				"ciudad":         t.Ciudad,
				"estado":         t.Estado,
				"pais":           t.Pais,
				"tipo_tienda":    t.TipoTienda,
				"estatus":        t.Estatus,
			})
		}

		// Contar pedidos realizados por el usuario
		var totalPedidos int
		err = dbc.Local.QueryRow("SELECT COUNT(*) FROM pedidos WHERE id_usuario = ?", idUsuario).Scan(&totalPedidos)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al contar pedidos", err.Error())
			return
		}

		writeSuccessResponse(w, "Perfil obtenido correctamente", map[string]interface{}{
			"usuario":       u,
			"tiendas":       tiendas,
			"total_pedidos": totalPedidos,
		})
	}
}

// ---------- EDITAR USUARIO (nombre, teléfono, contraseña) ----------

func EditarUsuario(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			IDUsuario      int    `json:"id_usuario"`
			NombreCompleto string `json:"nombre_completo"`
			Telefono       string `json:"telefono"`
			ClaveNueva     string `json:"clave_nueva"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Datos inválidos", err.Error())
			return
		}
		query := "UPDATE usuarios SET nombre_completo=?, telefono=?"
		args := []interface{}{req.NombreCompleto, req.Telefono}
		if req.ClaveNueva != "" {
			query += ", clave=?"
			args = append(args, req.ClaveNueva)
		}
		query += " WHERE id_usuario=?"
		args = append(args, req.IDUsuario)
		_, err := dbc.Local.Exec(query, args...)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo actualizar usuario", err.Error())
			return
		}
		writeSuccessResponse(w, "Usuario actualizado", nil)
	}
}

// ---------- ELIMINAR TIENDA (solo si el usuario queda con al menos 1 activa) ----------

func EliminarTienda(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idTienda := r.URL.Query().Get("id_tienda")
		idUsuario := r.URL.Query().Get("id_usuario")
		if idTienda == "" || idUsuario == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Faltan parámetros", "")
			return
		}
		var count int
		err := dbc.Local.QueryRow("SELECT COUNT(*) FROM tiendas WHERE id_usuario=? AND estatus='activo'", idUsuario).Scan(&count)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al contar tiendas", err.Error())
			return
		}
		if count <= 1 {
			writeErrorResponse(w, http.StatusBadRequest, "Debes tener al menos una tienda activa", "")
			return
		}
		_, err = dbc.Local.Exec("UPDATE tiendas SET estatus='eliminado' WHERE id_tienda=? AND id_usuario=?", idTienda, idUsuario)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo eliminar tienda", err.Error())
			return
		}
		writeSuccessResponse(w, "Tienda eliminada", nil)
	}
}
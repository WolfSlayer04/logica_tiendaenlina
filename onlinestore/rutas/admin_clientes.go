package rutas

import (
	"database/sql"
	
	"net/http"
	"onlinestore/db"
)

// --- Estructuras locales para respuesta compuesta ---

type usuarioTiendaRow struct {
	// Usuario
	IDUsuario      int            `json:"id_usuario"`
	IDEmpresa      int            `json:"id_empresa"`
	TipoUsuario    string         `json:"tipo_usuario"`
	NombreCompleto string         `json:"nombre_completo"`
	Correo         string         `json:"correo"`
	Telefono       string         `json:"telefono"`
	Clave          string         `json:"clave"`
	FechaRegistro  sql.NullString `json:"fecha_registro"`
	UltimoAcceso   sql.NullString `json:"ultimo_acceso"`
	Estatus        string         `json:"estatus"`
	RequiereClave  bool           `json:"requiere_cambiar_clave"`

	// Tienda (puede ser nula)
	IDTienda            sql.NullInt64   `json:"id_tienda"`
	TiendaIDUsuario     sql.NullInt64   `json:"tienda_id_usuario"`
	TiendaIDEmpresa     sql.NullInt64   `json:"tienda_id_empresa"`
	NombreTienda        sql.NullString  `json:"nombre_tienda"`
	RazonSocial         sql.NullString  `json:"razon_social"`
	RFC                 sql.NullString  `json:"rfc"`
	Direccion           sql.NullString  `json:"direccion"`
	Colonia             sql.NullString  `json:"colonia"`
	CodigoPostal        sql.NullString  `json:"codigo_postal"`
	Ciudad              sql.NullString  `json:"ciudad"`
	Estado              sql.NullString  `json:"estado"`
	Pais                sql.NullString  `json:"pais"`
	Latitud             sql.NullFloat64 `json:"latitud"`
	Longitud            sql.NullFloat64 `json:"longitud"`
	TipoTienda          sql.NullString  `json:"tipo_tienda"`
	HorarioApertura     sql.NullString  `json:"horario_apertura"`
	HorarioCierre       sql.NullString  `json:"horario_cierre"`
	DiasOperacion       sql.NullString  `json:"dias_operacion"`
	FechaRegTienda      sql.NullString  `json:"fecha_registro_tienda"`
	UltimaActualizacion sql.NullString  `json:"ultima_actualizacion"`
	EstatusTienda       sql.NullString  `json:"estatus_tienda"`
}

// --- Respuesta para admin clientes (usuario + tienda) ---

type AdminClienteResponse struct {
	Usuario map[string]interface{} `json:"usuario"`
	Tienda  map[string]interface{} `json:"tienda,omitempty"`
}

// --- Handler para obtener todos los usuarios con su tienda asociada ---

func GetAllUsuariosConTienda(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
			SELECT
				u.id_usuario, u.id_empresa, u.tipo_usuario, u.nombre_completo, u.correo, u.telefono, u.clave, 
				CAST(u.fecha_registro AS CHAR), CAST(u.ultimo_acceso AS CHAR), u.estatus, u.requiere_cambiar_clave,
				t.id_tienda, t.id_usuario, t.id_empresa, t.nombre_tienda, t.razon_social, t.rfc, t.direccion, t.colonia, t.codigo_postal, t.ciudad, t.estado, t.pais,
				t.latitud, t.longitud, t.tipo_tienda, t.horario_apertura, t.horario_cierre, t.dias_operacion, 
				CAST(t.fecha_registro AS CHAR), CAST(t.ultima_actualizacion AS CHAR), t.estatus
			FROM usuarios u
			LEFT JOIN tiendas t ON u.id_usuario = t.id_usuario
		`
		rows, err := dbc.Local.Query(query)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error consultando usuarios", err.Error())
			return
		}
		defer rows.Close()

		var result []AdminClienteResponse

		for rows.Next() {
			var row usuarioTiendaRow
			err := rows.Scan(
				&row.IDUsuario, &row.IDEmpresa, &row.TipoUsuario, &row.NombreCompleto, &row.Correo, &row.Telefono, &row.Clave,
				&row.FechaRegistro, &row.UltimoAcceso, &row.Estatus, &row.RequiereClave,
				&row.IDTienda, &row.TiendaIDUsuario, &row.TiendaIDEmpresa, &row.NombreTienda, &row.RazonSocial, &row.RFC, &row.Direccion, &row.Colonia, &row.CodigoPostal, &row.Ciudad, &row.Estado, &row.Pais,
				&row.Latitud, &row.Longitud, &row.TipoTienda, &row.HorarioApertura, &row.HorarioCierre, &row.DiasOperacion,
				&row.FechaRegTienda, &row.UltimaActualizacion, &row.EstatusTienda,
			)
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Error leyendo fila", err.Error())
				return
			}

			user := map[string]interface{}{
				"id_usuario":             row.IDUsuario,
				"id_empresa":             row.IDEmpresa,
				"tipo_usuario":           row.TipoUsuario,
				"nombre_completo":        row.NombreCompleto,
				"correo":                 row.Correo,
				"telefono":               row.Telefono,
				"clave":                  row.Clave,
				"fecha_registro":         nullStringToString(row.FechaRegistro),
				"ultimo_acceso":          nullStringToString(row.UltimoAcceso),
				"estatus":                row.Estatus,
				"requiere_cambiar_clave": row.RequiereClave,
			}
			var tienda map[string]interface{}
			if row.IDTienda.Valid {
				tienda = map[string]interface{}{
					"id_tienda":             row.IDTienda.Int64,
					"id_usuario":            nullIntToInt(row.TiendaIDUsuario),
					"id_empresa":            nullIntToInt(row.TiendaIDEmpresa),
					"nombre_tienda":         nullStringToString(row.NombreTienda),
					"razon_social":          nullStringToString(row.RazonSocial),
					"rfc":                   nullStringToString(row.RFC),
					"direccion":             nullStringToString(row.Direccion),
					"colonia":               nullStringToString(row.Colonia),
					"codigo_postal":         nullStringToString(row.CodigoPostal),
					"ciudad":                nullStringToString(row.Ciudad),
					"estado":                nullStringToString(row.Estado),
					"pais":                  nullStringToString(row.Pais),
					"latitud":               nullFloatToFloat(row.Latitud),
					"longitud":              nullFloatToFloat(row.Longitud),
					"tipo_tienda":           nullStringToString(row.TipoTienda),
					"horario_apertura":      nullStringToString(row.HorarioApertura),
					"horario_cierre":        nullStringToString(row.HorarioCierre),
					"dias_operacion":        nullStringToString(row.DiasOperacion),
					"fecha_registro":        nullStringToString(row.FechaRegTienda),
					"ultima_actualizacion":  nullStringToString(row.UltimaActualizacion),
					"estatus":               nullStringToString(row.EstatusTienda),
				}
			}

			resp := AdminClienteResponse{
				Usuario: user,
			}
			if tienda != nil && tienda["id_tienda"].(int64) != 0 {
				resp.Tienda = tienda
			}
			result = append(result, resp)
		}

		writeSuccessResponse(w, "Usuarios y tiendas obtenidos correctamente", result)
	}
}

// --- Utilidades para convertir tipos sql.NullXXX a Go nativos y string ---

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullIntToInt(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0
}

func nullFloatToFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

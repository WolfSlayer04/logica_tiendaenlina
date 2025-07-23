package rutas

import (
	"database/sql"
	"net/http"
	"github.com/WolfSlayer04/logica_tiendaenlina/db"
)

// ---------------------------
// ESTRUCTURA DE RESPUESTA INDICADORES
// ---------------------------
type IndicadorResponse struct {
	ID          int     `json:"id"`
	Fecha       string  `json:"fecha"`
	NumPedidos  int     `json:"num_pedidos"`
	TotPedidos  float64 `json:"tot_pedidos"`
	NumClientes int     `json:"num_clientes"`
}

// ---------------------------
// ENDPOINT: TODOS LOS INDICADORES DIARIOS
// ---------------------------
func GetIndicadoresDiarioAll(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dbc.Local.Query(`
			SELECT id, fecha, num_pedidos, tot_pedidos, num_clientes
			FROM ind_diario
			ORDER BY fecha DESC
		`)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener indicadores diarios", err.Error())
			return
		}
		defer rows.Close()

		var datos []IndicadorResponse
		for rows.Next() {
			var i IndicadorResponse
			var fecha string
			if err := rows.Scan(&i.ID, &fecha, &i.NumPedidos, &i.TotPedidos, &i.NumClientes); err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al leer indicadores diarios", err.Error())
				return
			}
			i.Fecha = fecha
			datos = append(datos, i)
		}
		writeSuccessResponse(w, "Indicadores diarios obtenidos correctamente", datos)
	}
}

// ---------------------------
// ENDPOINT: INDICADOR DIARIO POR FECHA
// ---------------------------
func GetIndicadorDiarioByFecha(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fecha := r.URL.Query().Get("fecha")
		if fecha == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Falta el parámetro fecha (YYYY-MM-DD)", "")
			return
		}
		var i IndicadorResponse
		var fechaVal string
		row := dbc.Local.QueryRow(`
			SELECT id, fecha, num_pedidos, tot_pedidos, num_clientes
			FROM ind_diario
			WHERE fecha = ?
			LIMIT 1
		`, fecha)
		err := row.Scan(&i.ID, &fechaVal, &i.NumPedidos, &i.TotPedidos, &i.NumClientes)
		if err != nil {
			if err == sql.ErrNoRows {
				writeErrorResponse(w, http.StatusNotFound, "No hay indicadores para la fecha", "")
			} else {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al buscar el indicador diario", err.Error())
			}
			return
		}
		i.Fecha = fechaVal
		writeSuccessResponse(w, "Indicador diario obtenido correctamente", i)
	}
}

// ---------------------------
// ENDPOINT: TODOS LOS INDICADORES MENSUALES
// ---------------------------
func GetIndicadoresMensualAll(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dbc.Local.Query(`
			SELECT id, fecha, num_pedidos, tot_pedidos, num_clientes
			FROM ind_mensual
			ORDER BY fecha DESC
		`)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener indicadores mensuales", err.Error())
			return
		}
		defer rows.Close()

		var datos []IndicadorResponse
		for rows.Next() {
			var i IndicadorResponse
			var fecha string
			if err := rows.Scan(&i.ID, &fecha, &i.NumPedidos, &i.TotPedidos, &i.NumClientes); err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al leer indicadores mensuales", err.Error())
				return
			}
			i.Fecha = fecha
			datos = append(datos, i)
		}
		writeSuccessResponse(w, "Indicadores mensuales obtenidos correctamente", datos)
	}
}

// ---------------------------
// ENDPOINT: INDICADOR MENSUAL POR FECHA (YYYY-MM o YYYY-MM-01)
// ---------------------------
func GetIndicadorMensualByFecha(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fecha := r.URL.Query().Get("fecha")
		if fecha == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Falta el parámetro fecha (YYYY-MM)", "")
			return
		}
		// Normaliza a primer día del mes si solo es YYYY-MM
		if len(fecha) == 7 {
			fecha = fecha + "-01"
		}
		var i IndicadorResponse
		var fechaVal string
		row := dbc.Local.QueryRow(`
			SELECT id, fecha, num_pedidos, tot_pedidos, num_clientes
			FROM ind_mensual
			WHERE fecha = ?
			LIMIT 1
		`, fecha)
		err := row.Scan(&i.ID, &fechaVal, &i.NumPedidos, &i.TotPedidos, &i.NumClientes)
		if err != nil {
			if err == sql.ErrNoRows {
				writeErrorResponse(w, http.StatusNotFound, "No hay indicadores para el mes", "")
			} else {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al buscar el indicador mensual", err.Error())
			}
			return
		}
		i.Fecha = fechaVal
		writeSuccessResponse(w, "Indicador mensual obtenido correctamente", i)
	}
}

// ---------------------------
// FUNCION: Inicializa los acumulados históricos si ya existen pedidos
// ---------------------------
func InicializaIndicadoresHistoricos(dbc *db.DBConnection) error {
	// DIARIO
	diarioQuery := `
INSERT INTO ind_diario (fecha, num_pedidos, tot_pedidos, num_clientes)
SELECT
	DATE(fecha_creacion) AS fecha,
	COUNT(*) AS num_pedidos,
	IFNULL(SUM(total), 0) AS tot_pedidos,
	COUNT(DISTINCT id_usuario) AS num_clientes
FROM pedidos
GROUP BY DATE(fecha_creacion)
ON DUPLICATE KEY UPDATE
	num_pedidos = VALUES(num_pedidos),
	tot_pedidos = VALUES(tot_pedidos),
	num_clientes = VALUES(num_clientes)
`
	// Se elimina el print de debug: fmt.Println("DEBUG QUERY DIARIO:\n", diarioQuery)
	_, err := dbc.Local.Exec(diarioQuery)
	if err != nil {
		return err
	}

	// MENSUAL
	mensualQuery := `
INSERT INTO ind_mensual (fecha, num_pedidos, tot_pedidos, num_clientes)
SELECT
	DATE_FORMAT(fecha_creacion, '%Y-%m-01') AS fecha,
	COUNT(*) AS num_pedidos,
	IFNULL(SUM(total), 0) AS tot_pedidos,
	COUNT(DISTINCT id_usuario) AS num_clientes
FROM pedidos
GROUP BY DATE_FORMAT(fecha_creacion, '%Y-%m-01')
ON DUPLICATE KEY UPDATE
	num_pedidos = VALUES(num_pedidos),
	tot_pedidos = VALUES(tot_pedidos),
	num_clientes = VALUES(num_clientes)
`
	// Se elimina el print de debug: fmt.Println("DEBUG QUERY MENSUAL:\n", mensualQuery)
	_, err = dbc.Local.Exec(mensualQuery)
	return err
}

// ---------------------------
// FUNCION: Acumula indicadores al crear un pedido
// ---------------------------
func AcumulaIndicadores(tx *sql.Tx, total float64, usuarioID int, fechaPedido string) error {
	// --- DIARIO ---
	dia := fechaPedido[:10] // "YYYY-MM-DD"
	var existe int
	err := tx.QueryRow(`SELECT COUNT(*) FROM ind_diario WHERE fecha = ?`, dia).Scan(&existe)
	if err != nil {
		return err
	}
	if existe > 0 {
		_, err = tx.Exec(`
			UPDATE ind_diario
			SET num_pedidos = num_pedidos + 1,
				tot_pedidos = tot_pedidos + ?,
				num_clientes = (SELECT COUNT(DISTINCT id_usuario) FROM pedidos WHERE DATE(fecha_creacion) = ?)
			WHERE fecha = ?`,
			total, dia, dia)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec(`
			INSERT INTO ind_diario (fecha, num_pedidos, tot_pedidos, num_clientes)
			VALUES (?, 1, ?, 1)`,
			dia, total)
		if err != nil {
			return err
		}
	}

	// --- MENSUAL ---
	mes := dia[:7] + "-01"
	err = tx.QueryRow(`SELECT COUNT(*) FROM ind_mensual WHERE fecha = ?`, mes).Scan(&existe)
	if err != nil {
		return err
	}
	if existe > 0 {
		_, err = tx.Exec(`
			UPDATE ind_mensual
			SET num_pedidos = num_pedidos + 1,
				tot_pedidos = tot_pedidos + ?,
				num_clientes = (SELECT COUNT(DISTINCT id_usuario) FROM pedidos WHERE fecha_creacion >= ? AND fecha_creacion < DATE_ADD(?, INTERVAL 1 MONTH))
			WHERE fecha = ?`,
			total, mes, mes, mes)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec(`
			INSERT INTO ind_mensual (fecha, num_pedidos, tot_pedidos, num_clientes)
			VALUES (?, 1, ?, 1)`,
			mes, total)
		if err != nil {
			return err
		}
	}
	return nil
}
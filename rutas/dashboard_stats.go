package rutas

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/WolfSlayer04/logica_tiendaenlina/db"
	"time"
)

type ProductosVendidosMes struct {
	Anio  int `json:"anio"`
	Mes   int `json:"mes"`
	Total int `json:"total"`
}

type DashboardStatsResponse struct {
	ProductosTotal             int                  `json:"productos_total"`
	ProductosActivos           int                  `json:"productos_activos"`
	PedidosTotal               int                  `json:"pedidos_total"`
	PedidosPendientes          int                  `json:"pedidos_pendientes"`
	PedidosPendientesMes       int                  `json:"pedidos_pendientes_mes"`
	UsuariosTotal              int                  `json:"usuarios_total"`
	UsuariosActivos            int                  `json:"usuarios_activos"`
	ProductosVendidosMes       int                  `json:"productos_vendidos_mes"`
	ProductosVendidosHistorico []ProductosVendidosMes `json:"productos_vendidos_historico"`
}

func GetDashboardStats(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var stats DashboardStatsResponse

		// Productos
		err := dbConn.Local.QueryRow("SELECT COUNT(*) FROM crm_productos").Scan(&stats.ProductosTotal)
		if err != nil {
			log.Println("Error al contar productos:", err)
			http.Error(w, "Error al contar productos", http.StatusInternalServerError)
			return
		}
		err = dbConn.Local.QueryRow("SELECT COUNT(*) FROM crm_productos WHERE estatus = 'S'").Scan(&stats.ProductosActivos)
		if err != nil {
			log.Println("Error al contar productos activos:", err)
			http.Error(w, "Error al contar productos activos", http.StatusInternalServerError)
			return
		}

		// Pedidos
		err = dbConn.Local.QueryRow("SELECT COUNT(*) FROM pedidos").Scan(&stats.PedidosTotal)
		if err != nil {
			log.Println("Error al contar pedidos:", err)
			http.Error(w, "Error al contar pedidos", http.StatusInternalServerError)
			return
		}
		err = dbConn.Local.QueryRow("SELECT COUNT(*) FROM pedidos WHERE estatus = 'pendiente'").Scan(&stats.PedidosPendientes)
		if err != nil {
			log.Println("Error al contar pedidos pendientes:", err)
			http.Error(w, "Error al contar pedidos pendientes", http.StatusInternalServerError)
			return
		}
		// Pedidos pendientes del mes actual (usa fecha_creacion en pedidos)
		err = dbConn.Local.QueryRow(`
			SELECT COUNT(*)
			FROM pedidos
			WHERE estatus = 'pendiente'
			  AND YEAR(fecha_creacion) = YEAR(NOW())
			  AND MONTH(fecha_creacion) = MONTH(NOW())
		`).Scan(&stats.PedidosPendientesMes)
		if err != nil {
			log.Println("Error al contar pedidos pendientes del mes:", err)
			http.Error(w, "Error al contar pedidos pendientes del mes", http.StatusInternalServerError)
			return
		}

		// Usuarios
		err = dbConn.Local.QueryRow("SELECT COUNT(*) FROM usuarios").Scan(&stats.UsuariosTotal)
		if err != nil {
			log.Println("Error al contar usuarios:", err)
			http.Error(w, "Error al contar usuarios", http.StatusInternalServerError)
			return
		}
		err = dbConn.Local.QueryRow("SELECT COUNT(*) FROM usuarios WHERE estatus = 'activo'").Scan(&stats.UsuariosActivos)
		if err != nil {
			log.Println("Error al contar usuarios activos:", err)
			http.Error(w, "Error al contar usuarios activos", http.StatusInternalServerError)
			return
		}

		// Productos vendidos por mes (histórico) usando fecha_creacion en pedidos
		rows, err := dbConn.Local.Query(`
			SELECT 
				YEAR(p.fecha_creacion) AS anio,
				MONTH(p.fecha_creacion) AS mes,
				COALESCE(SUM(dp.cantidad), 0) AS total
			FROM detalle_pedidos dp
			JOIN pedidos p ON p.id_pedido = dp.id_pedido
			WHERE p.estatus IN ('completado', 'solicitado', 'pendiente','enviado', 'procesando')
			GROUP BY anio, mes
			ORDER BY anio, mes
		`)
		if err != nil {
			log.Println("Error al consultar productos vendidos por mes:", err)
			http.Error(w, "Error al consultar productos vendidos por mes", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Inicializar el slice para que nunca sea null en JSON
		stats.ProductosVendidosHistorico = make([]ProductosVendidosMes, 0)

		for rows.Next() {
			var anio, mes int
			var totalFloat float64
			if err := rows.Scan(&anio, &mes, &totalFloat); err != nil {
				log.Println("Error al escanear fila de productos vendidos por mes:", err)
				http.Error(w, "Error al leer los datos", http.StatusInternalServerError)
				return
			}
			row := ProductosVendidosMes{
				Anio:  anio,
				Mes:   mes,
				Total: int(totalFloat),
			}
			stats.ProductosVendidosHistorico = append(stats.ProductosVendidosHistorico, row)
		}
		if err := rows.Err(); err != nil {
			log.Println("Error en rows de productos vendidos por mes:", err)
			http.Error(w, "Error al leer los datos", http.StatusInternalServerError)
			return
		}

		// Determinar productos vendidos en el mes actual a partir del histórico
		stats.ProductosVendidosMes = 0
		now := time.Now()
		for _, row := range stats.ProductosVendidosHistorico {
			if row.Anio == now.Year() && row.Mes == int(now.Month()) {
				stats.ProductosVendidosMes = row.Total
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
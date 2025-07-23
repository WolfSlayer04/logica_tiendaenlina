package rutas

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/WolfSlayer04/logica_tiendaenlina/db"

	"github.com/gorilla/mux"
)

// ----------- ESTRUCTURAS AUXILIARES -----------

type SucursalRequest struct {
	IDSucursal int `json:"id_sucursal"`
}

type EstatusRequest struct {
	Estatus string `json:"estatus"`
}

type DescuentoRequest struct {
	Descuento float64 `json:"descuento"`
}

type DetallePedidoUpdate struct {
	IDDetalle        int64   `json:"id_detalle"`
	Cantidad         float64 `json:"cantidad"`
	PrecioUnitario   float64 `json:"precio_unitario"`
	ImporteDescuento float64 `json:"importe_descuento"`
	Comentarios      string  `json:"comentarios"`
}

type DetallesUpdateRequest struct {
	Detalles []DetallePedidoUpdate `json:"detalles"`
}



// ----------- ASIGNAR/CAMBIAR SUCURSAL -----------

func AdminActualizarSucursalPedido(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPedidoStr := mux.Vars(r)["id_pedido"]
		idPedido, err := strconv.ParseInt(idPedidoStr, 10, 64)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "ID de pedido inválido", err.Error())
			return
		}
		var req SucursalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "JSON inválido", err.Error())
			return
		}
		if req.IDSucursal <= 0 {
			writeErrorResponse(w, http.StatusBadRequest, "ID de sucursal requerido", "")
			return
		}
		res, err := dbConn.Local.Exec("UPDATE pedidos SET id_sucursal = ? WHERE id_pedido = ?", req.IDSucursal, idPedido)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo actualizar la sucursal", err.Error())
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			writeErrorResponse(w, http.StatusNotFound, "Pedido no encontrado", "")
			return
		}
		writeSuccessResponse(w, "Sucursal asignada correctamente", nil)
	}
}

// ----------- ACTUALIZAR ESTATUS DEL PEDIDO -----------

func AdminActualizarEstatusPedido(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPedidoStr := mux.Vars(r)["id_pedido"]
		idPedido, err := strconv.ParseInt(idPedidoStr, 10, 64)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "ID de pedido inválido", err.Error())
			return
		}
		var req EstatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "JSON inválido", err.Error())
			return
		}
		if req.Estatus == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Estatus requerido", "")
			return
		}
		res, err := dbConn.Local.Exec("UPDATE pedidos SET estatus = ? WHERE id_pedido = ?", req.Estatus, idPedido)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo actualizar el estatus", err.Error())
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			writeErrorResponse(w, http.StatusNotFound, "Pedido no encontrado", "")
			return
		}
		writeSuccessResponse(w, "Estatus actualizado correctamente", nil)
	}
}

// ----------- APLICAR DESCUENTO GLOBAL AL PEDIDO -----------

func AdminAplicarDescuentoPedido(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPedidoStr := mux.Vars(r)["id_pedido"]
		idPedido, err := strconv.ParseInt(idPedidoStr, 10, 64)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "ID de pedido inválido", err.Error())
			return
		}
		var req DescuentoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "JSON inválido", err.Error())
			return
		}
		if req.Descuento < 0 {
			writeErrorResponse(w, http.StatusBadRequest, "El descuento no puede ser negativo", "")
			return
		}
		var subtotal, iva, ieps float64
		err = dbConn.Local.QueryRow("SELECT subtotal, iva, ieps FROM pedidos WHERE id_pedido = ?", idPedido).Scan(&subtotal, &iva, &ieps)
		if err == sql.ErrNoRows {
			writeErrorResponse(w, http.StatusNotFound, "Pedido no encontrado", "")
			return
		} else if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error consultando pedido", err.Error())
			return
		}
		total := subtotal - req.Descuento + iva + ieps
		if total < 0 {
			total = 0
		}
		_, err = dbConn.Local.Exec("UPDATE pedidos SET descuento = ?, total = ? WHERE id_pedido = ?", req.Descuento, total, idPedido)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo aplicar el descuento", err.Error())
			return
		}
		writeSuccessResponse(w, "Descuento aplicado correctamente", map[string]interface{}{
			"total_final": total,
		})
	}
}

// ----------- ACTUALIZAR DETALLES DE PRODUCTOS DEL PEDIDO -----------

func AdminActualizarDetallesPedido(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPedidoStr := mux.Vars(r)["id_pedido"]
		idPedido, err := strconv.ParseInt(idPedidoStr, 10, 64)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "ID de pedido inválido", err.Error())
			return
		}
		var req DetallesUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "JSON inválido", err.Error())
			return
		}
		if len(req.Detalles) == 0 {
			writeErrorResponse(w, http.StatusBadRequest, "Se requiere al menos un producto para actualizar", "")
			return
		}
		tx, err := dbConn.Local.Begin()
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo iniciar la transacción", err.Error())
			return
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				writeErrorResponse(w, http.StatusInternalServerError, "Error inesperado", fmt.Sprintf("%v", r))
			}
		}()
		var subtotal, totalDescuento, totalIVA, totalIEPS, total float64

		for _, d := range req.Detalles {
			if d.Cantidad <= 0 || d.PrecioUnitario < 0 {
				tx.Rollback()
				writeErrorResponse(w, http.StatusBadRequest, "Cantidad o precio inválido en algún producto", "")
				return
			}
			subt := d.PrecioUnitario * d.Cantidad
			iva := 0.0
			ieps := 0.0
			tot := subt - d.ImporteDescuento + iva + ieps

			_, err := tx.Exec(`
				UPDATE detalle_pedidos
				SET cantidad = ?, precio_unitario = ?, importe_descuento = ?, subtotal = ?, importe_iva = ?, importe_ieps = ?, total = ?, comentarios = ?
				WHERE id_detalle = ? AND id_pedido = ?
			`, d.Cantidad, d.PrecioUnitario, d.ImporteDescuento, subt, iva, ieps, tot, d.Comentarios, d.IDDetalle, idPedido)
			if err != nil {
				tx.Rollback()
				writeErrorResponse(w, http.StatusInternalServerError, "Error al actualizar detalle", err.Error())
				return
			}
			subtotal += subt
			totalDescuento += d.ImporteDescuento
			totalIVA += iva
			totalIEPS += ieps
			total += tot
		}
		_, err = tx.Exec(`
			UPDATE pedidos SET subtotal = ?, descuento = ?, iva = ?, ieps = ?, total = ?
			WHERE id_pedido = ?
		`, subtotal, totalDescuento, totalIVA, totalIEPS, total, idPedido)
		if err != nil {
			tx.Rollback()
			writeErrorResponse(w, http.StatusInternalServerError, "Error al actualizar totales del pedido", err.Error())
			return
		}
		if err := tx.Commit(); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al confirmar cambios", err.Error())
			return
		}
		writeSuccessResponse(w, "Detalles del pedido actualizados correctamente", map[string]interface{}{
			"subtotal":  subtotal,
			"descuento": totalDescuento,
			"iva":       totalIVA,
			"ieps":      totalIEPS,
			"total":     total,
		})
	}
}

// ----------- OBTENER LISTA DE PEDIDOS CON DETALLES Y NOMBRES -----------

func AdminGetPedidosConDetalles(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dbConn.Local.Query(`
			SELECT 
				p.id_pedido, p.clave_unica, p.id_usuario, p.id_tienda, p.id_sucursal, p.fecha_creacion, p.fecha_entrega,
				p.subtotal, p.descuento, p.iva, p.ieps, p.total, p.id_metodo_pago, p.referencia_pago, p.direccion_entrega,
				p.colonia_entrega, p.cp_entrega, p.ciudad_entrega, p.estado_entrega, p.latitud_entrega, p.longitud_entrega,
				p.estatus, p.comentarios, p.origen_pedido, p.id_lista_precio,
				u.nombre_completo as nombre_usuario,
				s.sucursal as nombre_sucursal
			FROM pedidos p
			LEFT JOIN usuarios u ON u.id_usuario = p.id_usuario
			LEFT JOIN adm_sucursales s ON s.idsucursal = p.id_sucursal
			ORDER BY p.id_pedido DESC
		`)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al consultar pedidos", err.Error())
			return
		}
		defer rows.Close()

		var pedidos []map[string]interface{}
		for rows.Next() {
			var p struct {
				IDPedido         int64
				ClaveUnica       string
				IDUsuario        int
				IDTienda         int
				IDSucursal       int
				FechaCreacion    string
				FechaEntrega     sql.NullString
				Subtotal         float64
				Descuento        float64
				IVA              float64
				IEPS             float64
				Total            float64
				IDMetodoPago     int
				ReferenciaPago   sql.NullString
				DireccionEntrega string
				ColoniaEntrega   string
				CPEntrega        string
				CiudadEntrega    string
				EstadoEntrega    string
				LatitudEntrega   sql.NullFloat64
				LongitudEntrega  sql.NullFloat64
				Estatus          string
				Comentarios      sql.NullString
				OrigenPedido     string
				IDListaPrecio    int
				NombreUsuario    sql.NullString
				NombreSucursal   sql.NullString
			}
			
			// CORRECCIÓN: Usar sql.NullString para las columnas que pueden ser NULL
			err := rows.Scan(
				&p.IDPedido,
				&p.ClaveUnica,
				&p.IDUsuario,
				&p.IDTienda,
				&p.IDSucursal,
				&p.FechaCreacion,
				&p.FechaEntrega,
				&p.Subtotal,
				&p.Descuento,
				&p.IVA,
				&p.IEPS,
				&p.Total,
				&p.IDMetodoPago,
				&p.ReferenciaPago,
				&p.DireccionEntrega,
				&p.ColoniaEntrega,
				&p.CPEntrega,
				&p.CiudadEntrega,
				&p.EstadoEntrega,
				&p.LatitudEntrega,
				&p.LongitudEntrega,
				&p.Estatus,
				&p.Comentarios,
				&p.OrigenPedido,
				&p.IDListaPrecio,
				&p.NombreUsuario,
				&p.NombreSucursal,
			)
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al leer pedido", err.Error())
				return
			}

			// CORRECCIÓN: Usar las funciones auxiliares para convertir NULL a string
			pedidoMap := map[string]interface{}{
				"id_pedido":         p.IDPedido,
				"clave_unica":       p.ClaveUnica,
				"id_usuario":        p.IDUsuario,
				"id_tienda":         p.IDTienda,
				"id_sucursal":       p.IDSucursal,
				"fecha_creacion":    p.FechaCreacion,
				"fecha_entrega":     NullToStr(p.FechaEntrega),
				"subtotal":          p.Subtotal,
				"descuento":         p.Descuento,
				"iva":               p.IVA,
				"ieps":              p.IEPS,
				"total":             p.Total,
				"id_metodo_pago":    p.IDMetodoPago,
				"referencia_pago":   NullToStr(p.ReferenciaPago),
				"direccion_entrega": p.DireccionEntrega,
				"colonia_entrega":   p.ColoniaEntrega,
				"cp_entrega":        p.CPEntrega,
				"ciudad_entrega":    p.CiudadEntrega,
				"estado_entrega":    p.EstadoEntrega,
				"latitud_entrega":   NullFloatToPtr(p.LatitudEntrega),
				"longitud_entrega":  NullFloatToPtr(p.LongitudEntrega),
				"estatus":           p.Estatus,
				"comentarios":       NullToStr(p.Comentarios),
				"origen_pedido":     p.OrigenPedido,
				"id_lista_precio":   p.IDListaPrecio,
				"nombre_usuario":    NullToStr(p.NombreUsuario),    // CORRECCIÓN: Usar NullToStr
				"nombre_sucursal":   NullToStr(p.NombreSucursal),   // CORRECCIÓN: Usar NullToStr
			}

			detRows, err := dbConn.Local.Query(`
				SELECT id_detalle, id_pedido, id_producto, clave_producto, descripcion, unidad, cantidad, precio_unitario,
				       porcentaje_descuento, importe_descuento, subtotal, importe_iva, importe_ieps, total,
				       latitud_entrega, longitud_entrega, estatus, comentarios, fecha_registro
				FROM detalle_pedidos
				WHERE id_pedido = ?
			`, p.IDPedido)
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al consultar detalles", err.Error())
				return
			}
			var detalles []map[string]interface{}
			for detRows.Next() {
				var d struct {
					IDDetalle           int64
					IDPedido            int64
					IDProducto          int64
					ClaveProducto       string
					Descripcion         string
					Unidad              string
					Cantidad            float64
					PrecioUnitario      float64
					PorcentajeDescuento float64
					ImporteDescuento    float64
					Subtotal            float64
					ImporteIVA          float64
					ImporteIEPS         float64
					Total               float64
					LatitudEntrega      sql.NullFloat64
					LongitudEntrega     sql.NullFloat64
					Estatus             string
					Comentarios         sql.NullString
					FechaRegistro       sql.NullString
				}
				_ = detRows.Scan(
					&d.IDDetalle,
					&d.IDPedido,
					&d.IDProducto,
					&d.ClaveProducto,
					&d.Descripcion,
					&d.Unidad,
					&d.Cantidad,
					&d.PrecioUnitario,
					&d.PorcentajeDescuento,
					&d.ImporteDescuento,
					&d.Subtotal,
					&d.ImporteIVA,
					&d.ImporteIEPS,
					&d.Total,
					&d.LatitudEntrega,
					&d.LongitudEntrega,
					&d.Estatus,
					&d.Comentarios,
					&d.FechaRegistro,
				)
				fechaRegistroStr := ""
				if d.FechaRegistro.Valid {
					fechaRegistroStr = d.FechaRegistro.String
				}
				detalles = append(detalles, map[string]interface{}{
					"id_detalle":           d.IDDetalle,
					"id_pedido":            d.IDPedido,
					"id_producto":          d.IDProducto,
					"clave_producto":       d.ClaveProducto,
					"descripcion":          d.Descripcion,
					"unidad":               d.Unidad,
					"cantidad":             d.Cantidad,
					"precio_unitario":      d.PrecioUnitario,
					"porcentaje_descuento": d.PorcentajeDescuento,
					"importe_descuento":    d.ImporteDescuento,
					"subtotal":             d.Subtotal,
					"importe_iva":          d.ImporteIVA,
					"importe_ieps":         d.ImporteIEPS,
					"total":                d.Total,
					"latitud_entrega":      NullFloatToPtr(d.LatitudEntrega),
					"longitud_entrega":     NullFloatToPtr(d.LongitudEntrega),
					"estatus":              d.Estatus,
					"comentarios":          NullToStr(d.Comentarios),
					"fecha_registro":       fechaRegistroStr,
				})
			}
			detRows.Close()

			pedidos = append(pedidos, map[string]interface{}{
				"pedido":   pedidoMap,
				"detalles": detalles,
			})
		}
		writeSuccessResponse(w, "Pedidos obtenidos correctamente", pedidos)
	}
}

// ----------- OBTENER UN PEDIDO POR ID (INCLUYE DETALLES, USUARIO Y SUCURSAL) -----------

func AdminGetPedidoByID(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPedidoStr := mux.Vars(r)["id_pedido"]
		idPedido, err := strconv.ParseInt(idPedidoStr, 10, 64)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "ID de pedido inválido", err.Error())
			return
		}
		var pedido map[string]interface{}
		row := dbConn.Local.QueryRow(`
			SELECT 
				p.id_pedido, p.clave_unica, p.id_usuario, p.id_tienda, p.id_sucursal, p.fecha_creacion, p.fecha_entrega,
				p.subtotal, p.descuento, p.iva, p.ieps, p.total, p.id_metodo_pago, p.referencia_pago, p.direccion_entrega,
				p.colonia_entrega, p.cp_entrega, p.ciudad_entrega, p.estado_entrega, p.latitud_entrega, p.longitud_entrega,
				p.estatus, p.comentarios, p.origen_pedido, p.id_lista_precio,
				u.nombre_completo as nombre_usuario,
				s.sucursal as nombre_sucursal
			FROM pedidos p
			LEFT JOIN usuarios u ON u.id_usuario = p.id_usuario
			LEFT JOIN adm_sucursales s ON s.idsucursal = p.id_sucursal
			WHERE p.id_pedido = ?
		`, idPedido)
		var p struct {
			IDPedido         int64
			ClaveUnica       string
			IDUsuario        int
			IDTienda         int
			IDSucursal       int
			FechaCreacion    string
			FechaEntrega     sql.NullString
			Subtotal         float64
			Descuento        float64
			IVA              float64
			IEPS             float64
			Total            float64
			IDMetodoPago     int
			ReferenciaPago   sql.NullString
			DireccionEntrega string
			ColoniaEntrega   string
			CPEntrega        string
			CiudadEntrega    string
			EstadoEntrega    string
			LatitudEntrega   sql.NullFloat64
			LongitudEntrega  sql.NullFloat64
			Estatus          string
			Comentarios      sql.NullString
			OrigenPedido     string
			IDListaPrecio    int
			NombreUsuario    sql.NullString
			NombreSucursal   sql.NullString
		}
		err = row.Scan(
			&p.IDPedido,
			&p.ClaveUnica,
			&p.IDUsuario,
			&p.IDTienda,
			&p.IDSucursal,
			&p.FechaCreacion,
			&p.FechaEntrega,
			&p.Subtotal,
			&p.Descuento,
			&p.IVA,
			&p.IEPS,
			&p.Total,
			&p.IDMetodoPago,
			&p.ReferenciaPago,
			&p.DireccionEntrega,
			&p.ColoniaEntrega,
			&p.CPEntrega,
			&p.CiudadEntrega,
			&p.EstadoEntrega,
			&p.LatitudEntrega,
			&p.LongitudEntrega,
			&p.Estatus,
			&p.Comentarios,
			&p.OrigenPedido,
			&p.IDListaPrecio,
			&p.NombreUsuario,
			&p.NombreSucursal,
		)
		if err == sql.ErrNoRows {
			writeErrorResponse(w, http.StatusNotFound, "Pedido no encontrado", "")
			return
		} else if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al consultar el pedido", err.Error())
			return
		}
		pedido = map[string]interface{}{
			"id_pedido":         p.IDPedido,
			"clave_unica":       p.ClaveUnica,
			"id_usuario":        p.IDUsuario,
			"id_tienda":         p.IDTienda,
			"id_sucursal":       p.IDSucursal,
			"fecha_creacion":    p.FechaCreacion,
			"fecha_entrega":     NullToStr(p.FechaEntrega),
			"subtotal":          p.Subtotal,
			"descuento":         p.Descuento,
			"iva":               p.IVA,
			"ieps":              p.IEPS,
			"total":             p.Total,
			"id_metodo_pago":    p.IDMetodoPago,
			"referencia_pago":   NullToStr(p.ReferenciaPago),
			"direccion_entrega": p.DireccionEntrega,
			"colonia_entrega":   p.ColoniaEntrega,
			"cp_entrega":        p.CPEntrega,
			"ciudad_entrega":    p.CiudadEntrega,
			"estado_entrega":    p.EstadoEntrega,
			"latitud_entrega":   NullFloatToPtr(p.LatitudEntrega),
			"longitud_entrega":  NullFloatToPtr(p.LongitudEntrega),
			"estatus":           p.Estatus,
			"comentarios":       NullToStr(p.Comentarios),
			"origen_pedido":     p.OrigenPedido,
			"id_lista_precio":   p.IDListaPrecio,
			"nombre_usuario":    NullToStr(p.NombreUsuario),
			"nombre_sucursal":   NullToStr(p.NombreSucursal),
		}
		rows, err := dbConn.Local.Query(`
			SELECT id_detalle, id_pedido, id_producto, clave_producto, descripcion, unidad, cantidad, precio_unitario,
			       porcentaje_descuento, importe_descuento, subtotal, importe_iva, importe_ieps, total,
			       latitud_entrega, longitud_entrega, estatus, comentarios, fecha_registro
			FROM detalle_pedidos
			WHERE id_pedido = ?
		`, idPedido)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al consultar detalles", err.Error())
			return
		}
		defer rows.Close()
		var detalles []map[string]interface{}
		for rows.Next() {
			var d struct {
				IDDetalle           int64
				IDPedido            int64
				IDProducto          int64
				ClaveProducto       string
				Descripcion         string
				Unidad              string
				Cantidad            float64
				PrecioUnitario      float64
				PorcentajeDescuento float64
				ImporteDescuento    float64
				Subtotal            float64
				ImporteIVA          float64
				ImporteIEPS         float64
				Total               float64
				LatitudEntrega      sql.NullFloat64
				LongitudEntrega     sql.NullFloat64
				Estatus             string
				Comentarios         sql.NullString
				FechaRegistro       sql.NullString
			}
			_ = rows.Scan(
				&d.IDDetalle,
				&d.IDPedido,
				&d.IDProducto,
				&d.ClaveProducto,
				&d.Descripcion,
				&d.Unidad,
				&d.Cantidad,
				&d.PrecioUnitario,
				&d.PorcentajeDescuento,
				&d.ImporteDescuento,
				&d.Subtotal,
				&d.ImporteIVA,
				&d.ImporteIEPS,
				&d.Total,
				&d.LatitudEntrega,
				&d.LongitudEntrega,
				&d.Estatus,
				&d.Comentarios,
				&d.FechaRegistro,
			)
			fechaRegistroStr := ""
			if d.FechaRegistro.Valid {
				fechaRegistroStr = d.FechaRegistro.String
			}
			detalles = append(detalles, map[string]interface{}{
				"id_detalle":           d.IDDetalle,
				"id_pedido":            d.IDPedido,
				"id_producto":          d.IDProducto,
				"clave_producto":       d.ClaveProducto,
				"descripcion":          d.Descripcion,
				"unidad":               d.Unidad,
				"cantidad":             d.Cantidad,
				"precio_unitario":      d.PrecioUnitario,
				"porcentaje_descuento": d.PorcentajeDescuento,
				"importe_descuento":    d.ImporteDescuento,
				"subtotal":             d.Subtotal,
				"importe_iva":          d.ImporteIVA,
				"importe_ieps":         d.ImporteIEPS,
				"total":                d.Total,
				"latitud_entrega":      NullFloatToPtr(d.LatitudEntrega),
				"longitud_entrega":     NullFloatToPtr(d.LongitudEntrega),
				"estatus":              d.Estatus,
				"comentarios":          NullToStr(d.Comentarios),
				"fecha_registro":       fechaRegistroStr,
			})
		}
		writeSuccessResponse(w, "Pedido obtenido correctamente", map[string]interface{}{
			"pedido":   pedido,
			"detalles": detalles,
		})
	}
}
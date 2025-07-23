package rutas

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/WolfSlayer04/logica_tiendaenlina/db"
	"time"
)

// --- Tipos usados en la sincronización ---

type Pedido struct {
	IDPedido            int64           `json:"id_pedido"`
	ClaveUnica          string          `json:"clave_unica"`
	IDUsuario           int             `json:"id_usuario"`
	IDTienda            int             `json:"id_tienda"`
	IDSucursal          int             `json:"id_sucursal"`
	FechaCreacion       string          `json:"fecha_creacion"`
	FechaEntrega        sql.NullTime    `json:"-"`
	Subtotal            float64         `json:"subtotal"`
	Descuento           float64         `json:"descuento"`
	IVA                 float64         `json:"iva"`
	IEPS                float64         `json:"ieps"`
	Total               float64         `json:"total"`
	IDMetodoPago        int             `json:"id_metodo_pago"`
	ReferenciaPago      sql.NullString  `json:"-"`
	DireccionEntrega    string          `json:"direccion_entrega"`
	ColoniaEntrega      string          `json:"colonia_entrega"`
	CPEntrega           string          `json:"cp_entrega"`
	CiudadEntrega       string          `json:"ciudad_entrega"`
	EstadoEntrega       string          `json:"estado_entrega"`
	PaisEntrega         string          `json:"pais_entrega,omitempty"`
	NombreTienda        string          `json:"nombre_tienda,omitempty"`
	LatitudEntrega      sql.NullFloat64 `json:"latitud_entrega"`
	LongitudEntrega     sql.NullFloat64 `json:"longitud_entrega"`
	Estatus             string          `json:"estatus"`
	Comentarios         sql.NullString  `json:"-"`
	OrigenPedido        string          `json:"origen_pedido"`
	IDListaPrecio       int             `json:"id_lista_precio"`
	Sincronizado        bool            `json:"sincronizado"`
	IDPrincipal         sql.NullInt64   `json:"id_principal"`
	IDRemoto            sql.NullInt64   `json:"id_remoto"`
	FechaSincronizacion sql.NullTime    `json:"fecha_sincronizacion"`
}

type DetallePedido struct {
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

type SincronizacionRequest struct {
	IDPedido   int64  `json:"id_pedido"`
	IDSucursal int64  `json:"id_sucursal"`
	ClaveUnica string `json:"clave_unica"`
}

type SincronizacionResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	IDPrincipal       int64  `json:"id_principal,omitempty"`
	IDAutoincremental int64  `json:"id_autoincremental,omitempty"`
}

// --- Funciones auxiliares ---

func NullStringToStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func GenerarClaveCliente(pedidoID int64, claveUnica string) string {
	input := fmt.Sprintf("%d-%s-%d", pedidoID, claveUnica, time.Now().UnixNano())
	hash := fmt.Sprintf("%x", md5.Sum([]byte(input)))
	if len(hash) > 32 {
		return hash[:32]
	}
	return hash
}

func FormatearFechaEntrega(fechaEntrega sql.NullTime) string {
	if !fechaEntrega.Valid {
		return time.Now().AddDate(0, 0, 2).Format("2006-01-02 15:04:05")
	}
	return fechaEntrega.Time.Format("2006-01-02 15:04:05")
}

func ActualizarFechaEntrega(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			IDPedido     int64  `json:"id_pedido"`
			FechaEntrega string `json:"fecha_entrega"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Error en el formato de la solicitud", err.Error())
			return
		}
		if req.IDPedido == 0 || req.FechaEntrega == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Campos requeridos faltantes", "IDPedido y FechaEntrega son obligatorios")
			return
		}
		formatos := []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		}
		var fechaEntrega time.Time
		var errParse error
		for _, formato := range formatos {
			fechaEntrega, errParse = time.Parse(formato, req.FechaEntrega)
			if errParse == nil {
				break
			}
		}
		if errParse != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Formato de fecha inválido", errParse.Error())
			return
		}
		tx, err := dbc.Local.Begin()
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al iniciar transacción", err.Error())
			return
		}
		defer tx.Rollback()
		result, err := tx.Exec(
			"UPDATE pedidos SET fecha_entrega = ? WHERE id_pedido = ?",
			fechaEntrega.Format("2006-01-02 15:04:05"), req.IDPedido)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al actualizar fecha de entrega", err.Error())
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			writeErrorResponse(w, http.StatusNotFound, "Pedido no encontrado", "")
			return
		}
		var sincronizado bool
		var idPrincipal sql.NullInt64
		var idRemoto sql.NullInt64
		err = tx.QueryRow(`
            SELECT COALESCE(sincronizado, false), id_principal, id_remoto 
            FROM pedidos WHERE id_pedido = ?`, 
            req.IDPedido).Scan(&sincronizado, &idPrincipal, &idRemoto)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al verificar sincronización", err.Error())
			return
		}
		if sincronizado && idRemoto.Valid {
			fechaActual := time.Now().Format("2006-01-02 15:04:05")
			_, err = dbc.Remote.Exec(
				"UPDATE crm_pedidos SET fecha_entrega = ?, fecha = ?, mom_entrega = ? WHERE id_pedido = ?",
				fechaEntrega.Format("2006-01-02 15:04:05"), 
				fechaActual,
				fechaActual,
				idRemoto.Int64)
			if err != nil {
				log.Printf("WolfSlayer04 - Error al actualizar fecha en sistema principal: %v", err)
			}
		}
		if err := tx.Commit(); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al confirmar transacción", err.Error())
			return
		}
		writeSuccessResponse(w, "Fecha de entrega actualizada correctamente", map[string]interface{}{
			"id_pedido":           req.IDPedido,
			"fecha_entrega":       fechaEntrega.Format("2006-01-02 15:04:05"),
			"sincronizado":        sincronizado,
			"id_principal":        NullToInt(idPrincipal),
			"id_autoincremental":  NullToInt(idRemoto),
		})
	}
}

func ObtenerPedidoPorIDRemoto(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idRemotoStr := r.URL.Query().Get("id_remoto")
		idPrincipalStr := r.URL.Query().Get("id_principal")
		if idRemotoStr == "" && idPrincipalStr == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Falta el parámetro id_remoto o id_principal", "")
			return
		}
		var query string
		var id int64
		var err error
		if idRemotoStr != "" {
			if _, err = fmt.Sscanf(idRemotoStr, "%d", &id); err != nil || id <= 0 {
				writeErrorResponse(w, http.StatusBadRequest, "id_remoto inválido", "")
				return
			}
			query = "SELECT id_pedido FROM pedidos WHERE id_remoto = ?"
		} else {
			if _, err = fmt.Sscanf(idPrincipalStr, "%d", &id); err != nil || id <= 0 {
				writeErrorResponse(w, http.StatusBadRequest, "id_principal inválido", "")
				return
			}
			query = "SELECT id_pedido FROM pedidos WHERE id_principal = ?"
		}
		var idPedidoLocal int64
		err = dbc.Local.QueryRow(query, id).Scan(&idPedidoLocal)
		if err != nil {
			if err == sql.ErrNoRows {
				writeErrorResponse(w, http.StatusNotFound, "No se encontró pedido con ese ID remoto", "")
			} else {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al buscar pedido", err.Error())
			}
			return
		}
		rows, err := dbc.Local.Query(`
			SELECT p.id_pedido, p.clave_unica, p.fecha_creacion, p.fecha_entrega,
				   p.total, p.estatus, p.id_principal, p.id_remoto, p.sincronizado,
				   p.fecha_sincronizacion
			FROM pedidos p
			WHERE p.id_pedido = ?
		`, idPedidoLocal)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener datos del pedido", err.Error())
			return
		}
		defer rows.Close()
		if !rows.Next() {
			writeErrorResponse(w, http.StatusNotFound, "Pedido no encontrado", "")
			return
		}
		var pedido struct {
			IDPedido            int64           `json:"id_pedido"`
			ClaveUnica          string          `json:"clave_unica"`
			FechaCreacion       string          `json:"fecha_creacion"`
			FechaEntrega        sql.NullString  `json:"fecha_entrega"`
			Total               float64         `json:"total"`
			Estatus             string          `json:"estatus"`
			IDPrincipal         sql.NullInt64   `json:"id_principal"`
			IDRemoto            sql.NullInt64   `json:"id_remoto"`
			Sincronizado        bool            `json:"sincronizado"`
			FechaSincronizacion sql.NullTime    `json:"fecha_sincronizacion"`
		}
		err = rows.Scan(
			&pedido.IDPedido, &pedido.ClaveUnica, &pedido.FechaCreacion, &pedido.FechaEntrega,
			&pedido.Total, &pedido.Estatus, &pedido.IDPrincipal, &pedido.IDRemoto,
			&pedido.Sincronizado, &pedido.FechaSincronizacion)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al leer datos del pedido", err.Error())
			return
		}
		response := map[string]interface{}{
			"id_pedido":            pedido.IDPedido,
			"clave_unica":          pedido.ClaveUnica,
			"fecha_creacion":       pedido.FechaCreacion,
			"fecha_entrega":        NullStringToStr(pedido.FechaEntrega),
			"total":                pedido.Total,
			"estatus":              pedido.Estatus,
			"id_principal":         NullToInt(pedido.IDPrincipal),
			"id_autoincremental":   NullToInt(pedido.IDRemoto),
			"sincronizado":         pedido.Sincronizado,
			"fecha_sincronizacion": NullTimeToStr(pedido.FechaSincronizacion),
		}
		writeSuccessResponse(w, "Pedido encontrado", response)
	}
}

// --- Lógica principal de sincronización ---

func sincronizarPedidoCore(dbc *db.DBConnection, req SincronizacionRequest, usuarioActual string) (*SincronizacionResponse, error) {
	currentTime := time.Now().UTC().Format("2006-01-02 15:04:05")
	txLocal, err := dbc.Local.Begin()
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción local: %w", err)
	}
	defer txLocal.Rollback()
	txRemote, err := dbc.Remote.Begin()
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción remota: %w", err)
	}
	defer txRemote.Rollback()

	var sincronizado bool
	err = txLocal.QueryRow("SELECT COALESCE(sincronizado, false) FROM pedidos WHERE id_pedido = ?", req.IDPedido).Scan(&sincronizado)
	if err != nil {
		return nil, fmt.Errorf("pedido no encontrado: %w", err)
	}
	if sincronizado {
		return nil, fmt.Errorf("el pedido ya está sincronizado")
	}

	var pedido Pedido
	var fechaEntregaStr sql.NullString
	err = txLocal.QueryRow(`
		SELECT id_pedido, clave_unica, id_usuario, id_tienda, id_sucursal, 
			   fecha_creacion, fecha_entrega, subtotal, descuento, iva, 
			   ieps, total, id_metodo_pago, referencia_pago, direccion_entrega,
			   colonia_entrega, cp_entrega, ciudad_entrega, estado_entrega,
			   latitud_entrega, longitud_entrega, estatus, comentarios, 
			   origen_pedido, id_lista_precio
		FROM pedidos WHERE id_pedido = ?`, req.IDPedido).Scan(
		&pedido.IDPedido, &pedido.ClaveUnica, &pedido.IDUsuario, &pedido.IDTienda,
		&pedido.IDSucursal, &pedido.FechaCreacion, &fechaEntregaStr,
		&pedido.Subtotal, &pedido.Descuento, &pedido.IVA, &pedido.IEPS,
		&pedido.Total, &pedido.IDMetodoPago, &pedido.ReferenciaPago,
		&pedido.DireccionEntrega, &pedido.ColoniaEntrega, &pedido.CPEntrega,
		&pedido.CiudadEntrega, &pedido.EstadoEntrega, &pedido.LatitudEntrega,
		&pedido.LongitudEntrega, &pedido.Estatus, &pedido.Comentarios,
		&pedido.OrigenPedido, &pedido.IDListaPrecio)
	if err != nil {
		return nil, fmt.Errorf("error al obtener el pedido: %w", err)
	}
	if fechaEntregaStr.Valid {
		fechaFormatos := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
		}
		var fechaParseada time.Time
		var errParse error
		for _, formato := range fechaFormatos {
			fechaParseada, errParse = time.Parse(formato, fechaEntregaStr.String)
			if errParse == nil {
				pedido.FechaEntrega = sql.NullTime{Time: fechaParseada, Valid: true}
				break
			}
		}
		if errParse != nil {
			pedido.FechaEntrega = sql.NullTime{Time: time.Now().AddDate(0, 0, 2), Valid: true}
		}
	} else {
		pedido.FechaEntrega = sql.NullTime{Time: time.Now().AddDate(0, 0, 2), Valid: true}
	}
	fechaEntregaFormateada := FormatearFechaEntrega(pedido.FechaEntrega)

	var nombreUsuario string
	var idClienteRemoto sql.NullInt64
	err = txLocal.QueryRow(`
		SELECT u.nombre_completo, u.id_remoto FROM usuarios u WHERE u.id_usuario = ?
	`, pedido.IDUsuario).Scan(&nombreUsuario, &idClienteRemoto)
	if err != nil {
		nombreUsuario = "Cliente"
	}

	rows, err := txLocal.Query(`
		SELECT id_detalle, id_pedido, id_producto, clave_producto, descripcion,
			   unidad, cantidad, precio_unitario, porcentaje_descuento,
			   importe_descuento, subtotal, importe_iva, importe_ieps, total,
			   latitud_entrega, longitud_entrega, estatus, comentarios, fecha_registro
		FROM detalle_pedidos WHERE id_pedido = ?`, req.IDPedido)
	if err != nil {
		return nil, fmt.Errorf("error al obtener detalles: %w", err)
	}
	defer rows.Close()
	var detalles []DetallePedido
	for rows.Next() {
		var det DetallePedido
		err := rows.Scan(
			&det.IDDetalle, &det.IDPedido, &det.IDProducto, &det.ClaveProducto,
			&det.Descripcion, &det.Unidad, &det.Cantidad, &det.PrecioUnitario,
			&det.PorcentajeDescuento, &det.ImporteDescuento, &det.Subtotal,
			&det.ImporteIVA, &det.ImporteIEPS, &det.Total, &det.LatitudEntrega,
			&det.LongitudEntrega, &det.Estatus, &det.Comentarios, &det.FechaRegistro)
		if err != nil {
			return nil, fmt.Errorf("error al leer detalle: %w", err)
		}
		detalles = append(detalles, det)
	}

	var valorAnterior int64
	err = txRemote.QueryRow(`SELECT COALESCE(idpedido, 0) FROM crm_indices WHERE idsucursal = ?`, req.IDSucursal).Scan(&valorAnterior)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error al obtener valor anterior: %w", err)
	}

	_, err = txRemote.Exec(`
		INSERT INTO crm_indices (idsucursal, idpedido) VALUES (?, 1)
		ON DUPLICATE KEY UPDATE idpedido = idpedido + 1
	`, req.IDSucursal)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar índices: %w", err)
	}

	var idPrincipal int64
	err = txRemote.QueryRow(`SELECT idpedido FROM crm_indices WHERE idsucursal = ?`, req.IDSucursal).Scan(&idPrincipal)
	if err != nil {
		return nil, fmt.Errorf("error al obtener ID principal: %w", err)
	}

	clavePedidoPrincipal := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d%s%d%s",
		req.IDSucursal, pedido.ClaveUnica, time.Now().Unix(), currentTime))))
	claveCliente := GenerarClaveCliente(pedido.IDPedido, pedido.ClaveUnica)
	idClienteVal := int64(0)
	if idClienteRemoto.Valid {
		idClienteVal = idClienteRemoto.Int64
	}

	result, err := txRemote.Exec(`
		INSERT INTO crm_pedidos (
			idpedido, clave_pedido, estatus, fecha_entrega, mom_creacion,
			fecha, mom_entrega, comentarios, clave_cliente, cliente, 
			persona, monto, iva, ieps, descuento, 
			facturar, idlista, idmetodopago, num_orden, tot_renglones, 
			web_movil, telefonico_presencial, idsucursal, id_cliente
		) VALUES (?, ?, 'N', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'N', ?, ?, ?, ?, ?, ?, ?, ?)`,
		idPrincipal,
		clavePedidoPrincipal,
		fechaEntregaFormateada,
		currentTime,
		currentTime,
		currentTime,
		"Pedido desde la pagina web",
		claveCliente,
		nombreUsuario,
		"Venta Tienda",
		pedido.Total,
		pedido.IVA,
		pedido.IEPS,
		pedido.Descuento,
		pedido.IDListaPrecio,
		pedido.IDMetodoPago,
		pedido.ClaveUnica,
		len(detalles),
		pedido.OrigenPedido == "web" || pedido.OrigenPedido == "movil",
		pedido.OrigenPedido == "telefonico" || pedido.OrigenPedido == "presencial",
		pedido.IDSucursal,
		idClienteVal)
	if err != nil {
		return nil, fmt.Errorf("error al insertar en sistema principal: %w", err)
	}

	var idPedidoAutoincremental int64
	idPedidoAutoincremental, err = result.LastInsertId()
	if err != nil {
		err = txRemote.QueryRow("SELECT LAST_INSERT_ID()").Scan(&idPedidoAutoincremental)
		if err != nil {
			err = txRemote.QueryRow("SELECT id_pedido FROM crm_pedidos WHERE idpedido = ? AND idsucursal = ?",
				idPrincipal, req.IDSucursal).Scan(&idPedidoAutoincremental)
			if err != nil {
				idPedidoAutoincremental = 0
			}
		}
	}

	// Insertar detalles con el mapeo correcto: precio (con IVA/IEPS), precio_o (sin), iva, ieps
	stmt, err := txRemote.Prepare(`
		INSERT INTO crm_pedidos_det (
			id_pedido, idproducto, orden, descripcion,
			precio, cantidad, precio_o, iva, ieps
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, fmt.Errorf("error al preparar inserción de detalles: %w", err)
	}
	defer stmt.Close()
	for i, det := range detalles {
		precioNeto := det.PrecioUnitario
		cantidad := det.Cantidad
		ivaUnitario := 0.0
		iepsUnitario := 0.0
		if cantidad > 0 {
			ivaUnitario = det.ImporteIVA / cantidad
			iepsUnitario = det.ImporteIEPS / cantidad
		}
		precioConIVA := precioNeto + ivaUnitario + iepsUnitario

		_, err = stmt.Exec(
			idPedidoAutoincremental,
			det.IDProducto, i+1, det.Descripcion,
			precioConIVA,         // precio (unitario con IVA/IEPS)
			cantidad,
			precioNeto,           // precio_o (unitario sin impuestos)
			ivaUnitario,          // iva (unitario)
			iepsUnitario)         // ieps (unitario)
		if err != nil {
			return nil, fmt.Errorf("error al insertar detalle en sistema principal: %w", err)
		}
	}

	_, err = txRemote.Exec(`
		INSERT INTO est_ventas_x_producto_dia 
		(idsucursal, idproducto, dia, cantidad) 
		VALUES (?, ?, DATE(?), ?)
		ON DUPLICATE KEY UPDATE 
		cantidad = cantidad + ?`,
		req.IDSucursal, idPedidoAutoincremental, currentTime, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar estadísticas diarias: %w", err)
	}

	_, err = txRemote.Exec(`
		INSERT INTO est_ventas_x_producto_mes
		(idsucursal, idproducto, dia, cantidad)
		VALUES (?, ?, DATE_FORMAT(?, '%Y-%m-01'), ?)
		ON DUPLICATE KEY UPDATE
		cantidad = cantidad + ?`,
		req.IDSucursal, idPedidoAutoincremental, currentTime, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar estadísticas mensuales: %w", err)
	}

	_, err = txLocal.Exec(`
		INSERT INTO log_sincronizacion 
		(id_pedido, id_principal, id_remoto, fecha_sincronizacion, usuario, estado) 
		VALUES (?, ?, ?, ?, ?, 'OK')`,
		req.IDPedido, idPrincipal, idPedidoAutoincremental, currentTime, usuarioActual)
	if err != nil {
		log.Printf("Error al registrar log de sincronización: %v\n", err)
	}

	_, err = txLocal.Exec(`
		UPDATE pedidos 
		SET sincronizado = true,
			id_principal = ?,
			id_remoto = ?,
			fecha_sincronizacion = ?
		WHERE id_pedido = ?`,
		idPrincipal, idPedidoAutoincremental, currentTime, req.IDPedido)
	if err != nil {
		return nil, fmt.Errorf("error al marcar sincronización: %w", err)
	}

	if err := txRemote.Commit(); err != nil {
		return nil, fmt.Errorf("error al confirmar transacción remota: %w", err)
	}
	if err := txLocal.Commit(); err != nil {
		return nil, fmt.Errorf("error al confirmar transacción local: %w", err)
	}

	return &SincronizacionResponse{
		Success:           true,
		IDPrincipal:       idPrincipal,
		IDAutoincremental: idPedidoAutoincremental,
		Message:           fmt.Sprintf("Sincronización completada con éxito el %s", currentTime),
	}, nil
}

// --- Handler HTTP para sincronizar pedido ---
func SincronizarPedido(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErrorResponse(w, http.StatusMethodNotAllowed, "Método no permitido", "")
			return
		}
		var req SincronizacionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Error en el formato de la solicitud", err.Error())
			return
		}
		if req.IDPedido == 0 || req.IDSucursal == 0 {
			writeErrorResponse(w, http.StatusBadRequest, "Campos requeridos faltantes", "IDPedido y IDSucursal son obligatorios")
			return
		}
		usuarioActual := r.Header.Get("X-User")
		if usuarioActual == "" {
			usuarioActual = "WolfSlayer04"
		}
		resp, err := sincronizarPedidoCore(dbc, req, usuarioActual)
		if err != nil {
			writeErrorResponse(w, http.StatusConflict, fmt.Sprintf("Error al sincronizar: %v", err), "")
			return
		}
		writeSuccessResponse(w, "Pedido sincronizado correctamente", resp)
	}
}

// --- Sincronización en background (automática) ---
func SincronizarPedidoBackground(dbc *db.DBConnection, idPedido int64) error {
	var idSucursal int64
	var claveUnica string
	err := dbc.Local.QueryRow("SELECT id_sucursal, clave_unica FROM pedidos WHERE id_pedido = ?", idPedido).Scan(&idSucursal, &claveUnica)
	if err != nil {
		return err
	}
	req := SincronizacionRequest{
		IDPedido:   idPedido,
		IDSucursal: idSucursal,
		ClaveUnica: claveUnica,
	}
	_, err = sincronizarPedidoCore(dbc, req, "DiablitoSincronizador")
	return err
}

func IniciarSincronizadorPedidos(dbc *db.DBConnection) {
	go func() {
		for {
			rows, err := dbc.Local.Query(`SELECT id_pedido FROM pedidos WHERE sincronizado = false OR sincronizado IS NULL`)
			if err != nil {
				log.Printf("Sincronizador: Error consultando pedidos pendientes: %v", err)
				time.Sleep(2 * time.Minute)
				continue
			}

			var idsPendientes []int64
			for rows.Next() {
				var id int64
				if err := rows.Scan(&id); err == nil {
					idsPendientes = append(idsPendientes, id)
				}
			}
			rows.Close()

			for _, id := range idsPendientes {
				err := SincronizarPedidoBackground(dbc, id)
				if err != nil {
					log.Printf("Sincronizador: Error sincronizando pedido %d: %v", id, err)
				} else {
					log.Printf("Sincronizador: Pedido %d sincronizado correctamente.", id)
				}
			}

			res, err := dbc.Local.Exec(`UPDATE pedidos SET estatus = 'procesando' WHERE sincronizado = true AND estatus = 'pendiente'`)
			if err != nil {
				log.Printf("Sincronizador: Error actualizando estatus a 'procesando': %v", err)
			} else if count, _ := res.RowsAffected(); count > 0 {
				log.Printf("Sincronizador: %d pedidos cambiados de 'pendiente' a 'procesando'.", count)
			} else {
				log.Println("Sincronizador: No había pedidos para cambiar a 'procesando'.")
			}

			time.Sleep(1 * time.Minute)
		}
	}()
}

// --- Verificación y consulta de sincronización ---

func VerificarSincronizacion(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idPedidoStr := r.URL.Query().Get("id_pedido")
		if idPedidoStr == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Falta el parámetro id_pedido", "")
			return
		}
		var idPedido int64
		if _, err := fmt.Sscanf(idPedidoStr, "%d", &idPedido); err != nil || idPedido <= 0 {
			writeErrorResponse(w, http.StatusBadRequest, "id_pedido inválido", "")
			return
		}
		var sincronizado bool
		var idPrincipal sql.NullInt64
		var idRemoto sql.NullInt64
		var fechaSincronizacion sql.NullTime
		err := dbc.Local.QueryRow(`
			SELECT COALESCE(sincronizado, false), id_principal, id_remoto, fecha_sincronizacion 
			FROM pedidos WHERE id_pedido = ?
		`, idPedido).Scan(&sincronizado, &idPrincipal, &idRemoto, &fechaSincronizacion)
		if err != nil {
			if err == sql.ErrNoRows {
				writeErrorResponse(w, http.StatusNotFound, "Pedido no encontrado", "")
			} else {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al verificar sincronización", err.Error())
			}
			return
		}
		response := map[string]interface{}{
			"id_pedido":    idPedido,
			"sincronizado": sincronizado,
		}
		if sincronizado {
			response["id_principal"] = NullToInt(idPrincipal)
			response["id_autoincremental"] = NullToInt(idRemoto)
			response["fecha_sincronizacion"] = NullTimeToStr(fechaSincronizacion)
		}
		writeSuccessResponse(w, "Estado de sincronización verificado", response)
	}
}

func PedidosPendientesSincronizacion(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dbc.Local.Query(`
			SELECT id_pedido, clave_unica, fecha_creacion, total, estatus 
			FROM pedidos 
			WHERE sincronizado = false OR sincronizado IS NULL
			ORDER BY fecha_creacion DESC
		`)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener pedidos pendientes", err.Error())
			return
		}
		defer rows.Close()
		var pendientes []map[string]interface{}
		for rows.Next() {
			var idPedido int64
			var claveUnica, fechaCreacion, estatus string
			var total float64
			if err := rows.Scan(&idPedido, &claveUnica, &fechaCreacion, &total, &estatus); err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al leer datos de pedido", err.Error())
				return
			}
			pendientes = append(pendientes, map[string]interface{}{
				"id_pedido":      idPedido,
				"clave_unica":    claveUnica,
				"fecha_creacion": fechaCreacion,
				"total":          total,
				"estatus":        estatus,
			})
		}
		writeSuccessResponse(w, fmt.Sprintf("Se encontraron %d pedidos pendientes de sincronizar", len(pendientes)), pendientes)
	}
}
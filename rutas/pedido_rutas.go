package rutas

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "math"
    "net/http"
    "strconv"
    "strings"
    "time"
    "github.com/WolfSlayer04/logica_tiendaenlina/db"
)

// ---------------------------
// ESTRUCTURAS DE RESPUESTA
// ---------------------------
type ErrorResponse struct {
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
}

type SuccessResponse struct {
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type PedidoDetalleRequest struct {
    IDProducto          int64    `json:"id_producto"`
    ClaveProducto       string   `json:"clave_producto"`
    Descripcion         string   `json:"descripcion"`
    Unidad              string   `json:"unidad"`
    Cantidad            float64  `json:"cantidad"`
    PrecioUnitario      float64  `json:"precio_unitario"`
    PorcentajeDescuento float64  `json:"porcentaje_descuento"`
    ImporteDescuento    float64  `json:"importe_descuento"`
    IVA                 float64  `json:"iva"`
    IEPS                float64  `json:"ieps"`
    Comentarios         string   `json:"comentarios"`
    LatitudEntrega      *float64 `json:"latitud_entrega,omitempty"`
    LongitudEntrega     *float64 `json:"longitud_entrega,omitempty"`
}

type PedidoRequest struct {
    IDUsuario         int                    `json:"id_usuario"`
    IDTienda          int                    `json:"id_tienda"`
    IDSucursal        int                    `json:"id_sucursal"`
    FechaEntrega      sql.NullTime           `json:"fecha_entrega"`
    IDMetodoPago      int                    `json:"id_metodo_pago"`
    ReferenciaPago    sql.NullString         `json:"referencia_pago"`
    DireccionEntrega  string                 `json:"direccion_entrega"`
    ColoniaEntrega    string                 `json:"colonia_entrega"`
    CPEntrega         string                 `json:"cp_entrega"`
    CiudadEntrega     string                 `json:"ciudad_entrega"`
    EstadoEntrega     string                 `json:"estado_entrega"`
    LatitudEntrega    *float64               `json:"latitud_entrega,omitempty"`
    LongitudEntrega   *float64               `json:"longitud_entrega,omitempty"`
    OrigenPedido      string                 `json:"origen_pedido"`
    Comentarios       sql.NullString         `json:"comentarios"`
    Detalles          []PedidoDetalleRequest `json:"detalles"`
}


// ---------------------------
// AUXILIARES
// ---------------------------
func writeErrorResponse(w http.ResponseWriter, status int, message, detail string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    resp := ErrorResponse{Message: message, Detail: detail}
    log.Printf("HTTP %d: %s | %s", status, message, detail)
    _ = json.NewEncoder(w).Encode(resp)
}

func writeSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(SuccessResponse{
        Message: message,
        Data:    data,
    })
}

func round(num float64, decimals ...int) float64 {
    d := 2
    if len(decimals) > 0 && decimals[0] >= 0 {
        d = decimals[0]
    }
    factor := math.Pow(10, float64(d))
    return math.Round(num*factor) / factor
}

func NullFloatToPtr(nf sql.NullFloat64) *float64 {
    if nf.Valid {
        return &nf.Float64
    }
    return nil
}

func NullToStrOrNil(ns sql.NullString) interface{} {
    if ns.Valid {
        return ns.String
    }
    return nil
}

func NullTimeToStrOrNil(nt sql.NullTime) interface{} {
    if nt.Valid {
        return nt.Time.Format("2006-01-02 15:04:05")
    }
    return nil
}

func NullToStr(ns sql.NullString) string {
    if ns.Valid {
        return ns.String
    }
    return ""
}

func NullTimeToStr(nt sql.NullTime) string {
    if nt.Valid {
        return nt.Time.Format("2006-01-02 15:04:05")
    }
    return ""
}

// ---------------------------
// FORMATO FECHA MES CORTO
// ---------------------------
var mesesAbrev = [...]string{"ENE", "FEB", "MAR", "ABR", "MAY", "JUN", "JUL", "AGO", "SEP", "OCT", "NOV", "DIC"}

func formateaFechaCorta(t time.Time) string {
    if t.IsZero() {
        return ""
    }
    mes := mesesAbrev[int(t.Month())-1]
    return fmt.Sprintf("%02d-%s-%04d %02d:%02d:%02d", t.Day(), mes, t.Year(), t.Hour(), t.Minute(), t.Second())
}

func NullTimeToStrMes(nt sql.NullTime) string {
    if nt.Valid {
        return formateaFechaCorta(nt.Time)
    }
    return ""
}

// ---------------------------
// FUNCIONES DE CÁLCULO DE FECHAS DE ENTREGA
// ---------------------------

func ObtenerConfigEntrega(dbc *db.DBConnection) (*ConfigEntrega, error) {
    var configJSON []byte
    err := dbc.Local.QueryRow("SELECT config_entrega FROM admin_usuarios WHERE tipo_usuario = 'Admin' LIMIT 1").Scan(&configJSON)
    if err != nil {
        if err == sql.ErrNoRows {
            return &ConfigEntrega{
                DiasHabiles:        []string{"LUNES", "MARTES", "MIERCOLES", "JUEVES", "VIERNES"},
                TiempoProcesamiento: 2,
                ReglasFindeSemana: struct {
                    ProcesarSabado       bool `json:"procesar_sabado"`
                    ProcesarDomingo      bool `json:"procesar_domingo"`
                    DiasAdicionalesSabado int  `json:"dias_adicionales_sabado"`
                    DiasAdicionalesDomingo int `json:"dias_adicionales_domingo"`
                }{
                    ProcesarSabado:       true,
                    ProcesarDomingo:      false,
                    DiasAdicionalesSabado: 1,
                    DiasAdicionalesDomingo: 2,
                },
                HorariosEntrega: []struct {
                    Etiqueta string `json:"etiqueta"`
                    Inicio   string `json:"inicio"`
                    Fin      string `json:"fin"`
                }{
                    {Etiqueta: "Mañana", Inicio: "09:00", Fin: "12:00"},
                    {Etiqueta: "Tarde", Inicio: "13:00", Fin: "18:00"},
                },
                DiasFeriados: []struct {
                    Fecha          string `json:"fecha"`
                    DiasAdicionales int   `json:"dias_adicionales"`
                }{},
            }, nil
        }
        return nil, err
    }

    if len(configJSON) == 0 {
        return &ConfigEntrega{
            DiasHabiles:        []string{"LUNES", "MARTES", "MIERCOLES", "JUEVES", "VIERNES"},
            TiempoProcesamiento: 2,
            ReglasFindeSemana: struct {
                ProcesarSabado       bool `json:"procesar_sabado"`
                ProcesarDomingo      bool `json:"procesar_domingo"`
                DiasAdicionalesSabado int  `json:"dias_adicionales_sabado"`
                DiasAdicionalesDomingo int `json:"dias_adicionales_domingo"`
            }{
                ProcesarSabado:       true,
                ProcesarDomingo:      false,
                DiasAdicionalesSabado: 1,
                DiasAdicionalesDomingo: 2,
            },
            HorariosEntrega: []struct {
                Etiqueta string `json:"etiqueta"`
                Inicio   string `json:"inicio"`
                Fin      string `json:"fin"`
            }{
                {Etiqueta: "Mañana", Inicio: "09:00", Fin: "12:00"},
                {Etiqueta: "Tarde", Inicio: "13:00", Fin: "18:00"},
            },
            DiasFeriados: []struct {
                Fecha          string `json:"fecha"`
                DiasAdicionales int   `json:"dias_adicionales"`
            }{},
        }, nil
    }

    var config ConfigEntrega
    if err := json.Unmarshal(configJSON, &config); err != nil {
        return nil, err
    }

    return &config, nil
}

func obtenerNombreDiaSemana(fecha time.Time) string {
    diasSemana := []string{"DOMINGO", "LUNES", "MARTES", "MIERCOLES", "JUEVES", "VIERNES", "SABADO"}
    return diasSemana[fecha.Weekday()]
}

func esDiaHabil(fecha time.Time, config *ConfigEntrega) bool {
    nombreDia := obtenerNombreDiaSemana(fecha)
    for _, diaHabil := range config.DiasHabiles {
        if diaHabil == nombreDia {
            return true
        }
    }
    return false
}

func esDiaFeriado(fecha time.Time, config *ConfigEntrega) (bool, int) {
    fechaStr := fecha.Format("2006-01-02")
    for _, feriado := range config.DiasFeriados {
        if strings.HasPrefix(feriado.Fecha, fechaStr) {
            return true, feriado.DiasAdicionales
        }
    }
    return false, 0
}

func CalcularFechaEntrega(fechaPedido time.Time, config *ConfigEntrega) time.Time {
    fechaEntrega := fechaPedido
    diaSemana := fechaEntrega.Weekday()
    esSabado := diaSemana == time.Saturday
    esDomingo := diaSemana == time.Sunday
    diasAdicionales := 0

    if esSabado {
        if !config.ReglasFindeSemana.ProcesarSabado {
            fechaEntrega = fechaEntrega.AddDate(0, 0, 1)
            if !config.ReglasFindeSemana.ProcesarDomingo {
                fechaEntrega = fechaEntrega.AddDate(0, 0, 1)
            }
        }
        diasAdicionales += config.ReglasFindeSemana.DiasAdicionalesSabado
    } else if esDomingo {
        if !config.ReglasFindeSemana.ProcesarDomingo {
            fechaEntrega = fechaEntrega.AddDate(0, 0, 1)
        }
        diasAdicionales += config.ReglasFindeSemana.DiasAdicionalesDomingo
    }

    esFeriado, diasAdicionalesFeriado := esDiaFeriado(fechaEntrega, config)
    if esFeriado {
        diasAdicionales += diasAdicionalesFeriado
    }

    diasProcesamiento := config.TiempoProcesamiento + diasAdicionales
    for i := 0; i < diasProcesamiento; {
        fechaEntrega = fechaEntrega.AddDate(0, 0, 1)
        if esDiaHabil(fechaEntrega, config) {
            esFeriado, _ := esDiaFeriado(fechaEntrega, config)
            if !esFeriado {
                i++
            }
        }
    }

    if len(config.HorariosEntrega) > 0 {
        horarioInicio := config.HorariosEntrega[0].Inicio
        horaParts := strings.Split(horarioInicio, ":")
        if len(horaParts) >= 2 {
            hora, _ := strconv.Atoi(horaParts[0])
            minuto, _ := strconv.Atoi(horaParts[1])
            fechaEntrega = time.Date(
                fechaEntrega.Year(),
                fechaEntrega.Month(),
                fechaEntrega.Day(),
                hora, minuto, 0, 0,
                fechaEntrega.Location(),
            )
        }
    }
    return fechaEntrega
}


// ---------------------------
// ENDPOINT: Obtener TODOS los pedidos
// ---------------------------
func GetAllPedidos(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        rows, err := dbc.Local.Query(`
            SELECT id_pedido, clave_unica, id_usuario, id_tienda, id_sucursal, fecha_creacion, fecha_entrega,
                   subtotal, descuento, iva, ieps, total, id_metodo_pago, referencia_pago, direccion_entrega,
                   colonia_entrega, cp_entrega, ciudad_entrega, estado_entrega, latitud_entrega, longitud_entrega,
                   estatus, comentarios, origen_pedido, id_lista_precio
            FROM pedidos ORDER BY fecha_creacion DESC
        `)
        if err != nil {
            writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener los pedidos", err.Error())
            return
        }
        defer rows.Close()

        var pedidos []map[string]interface{}
        for rows.Next() {
            var pedido Pedido
            var fechaEntregaStr sql.NullString

            err := rows.Scan(
                &pedido.IDPedido,
                &pedido.ClaveUnica,
                &pedido.IDUsuario,
                &pedido.IDTienda,
                &pedido.IDSucursal,
                &pedido.FechaCreacion,
                &fechaEntregaStr,
                &pedido.Subtotal,
                &pedido.Descuento,
                &pedido.IVA,
                &pedido.IEPS,
                &pedido.Total,
                &pedido.IDMetodoPago,
                &pedido.ReferenciaPago,
                &pedido.DireccionEntrega,
                &pedido.ColoniaEntrega,
                &pedido.CPEntrega,
                &pedido.CiudadEntrega,
                &pedido.EstadoEntrega,
                &pedido.LatitudEntrega,
                &pedido.LongitudEntrega,
                &pedido.Estatus,
                &pedido.Comentarios,
                &pedido.OrigenPedido,
                &pedido.IDListaPrecio,
            )
            if err != nil {
                writeErrorResponse(w, http.StatusInternalServerError, "Error al leer pedido", err.Error())
                return
            }
            if fechaEntregaStr.Valid {
                t, err := time.Parse("2006-01-02 15:04:05", fechaEntregaStr.String)
                if err == nil {
                    pedido.FechaEntrega = sql.NullTime{
                        Time: t,
                        Valid: true,
                    }
                }
            }
            fechaCreacionTime, _ := time.Parse("2006-01-02 15:04:05", pedido.FechaCreacion)
            pedidoMap := map[string]interface{}{
                "id_pedido":         pedido.IDPedido,
                "clave_unica":       pedido.ClaveUnica,
                "id_usuario":        pedido.IDUsuario,
                "id_tienda":         pedido.IDTienda,
                "id_sucursal":       pedido.IDSucursal,
                "fecha_creacion":    formateaFechaCorta(fechaCreacionTime),
                "fecha_entrega":     NullTimeToStrMes(pedido.FechaEntrega),
                "subtotal":          pedido.Subtotal,
                "descuento":         pedido.Descuento,
                "iva":               pedido.IVA,
                "ieps":              pedido.IEPS,
                "total":             pedido.Total,
                "id_metodo_pago":    pedido.IDMetodoPago,
                "referencia_pago":   NullToStr(pedido.ReferenciaPago),
                "direccion_entrega": pedido.DireccionEntrega,
                "colonia_entrega":   pedido.ColoniaEntrega,
                "cp_entrega":        pedido.CPEntrega,
                "ciudad_entrega":    pedido.CiudadEntrega,
                "estado_entrega":    pedido.EstadoEntrega,
                "latitud_entrega":   NullFloatToPtr(pedido.LatitudEntrega),
                "longitud_entrega":  NullFloatToPtr(pedido.LongitudEntrega),
                "estatus":           pedido.Estatus,
                "comentarios":       NullToStr(pedido.Comentarios),
                "origen_pedido":     pedido.OrigenPedido,
                "id_lista_precio":   pedido.IDListaPrecio,
            }

            detallesRows, err := dbc.Local.Query(`
                SELECT id_detalle, id_pedido, id_producto, clave_producto, descripcion, unidad, cantidad, precio_unitario, 
                       porcentaje_descuento, importe_descuento, subtotal, importe_iva, importe_ieps, total, latitud_entrega, 
                       longitud_entrega, estatus, comentarios, fecha_registro
                FROM detalle_pedidos WHERE id_pedido = ?`, pedido.IDPedido)
            if err != nil {
                writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener detalles", err.Error())
                return
            }
            var detalles []map[string]interface{}
            for detallesRows.Next() {
                var det DetallePedido
                err := detallesRows.Scan(
                    &det.IDDetalle,
                    &det.IDPedido,
                    &det.IDProducto,
                    &det.ClaveProducto,
                    &det.Descripcion,
                    &det.Unidad,
                    &det.Cantidad,
                    &det.PrecioUnitario,
                    &det.PorcentajeDescuento,
                    &det.ImporteDescuento,
                    &det.Subtotal,
                    &det.ImporteIVA,
                    &det.ImporteIEPS,
                    &det.Total,
                    &det.LatitudEntrega,
                    &det.LongitudEntrega,
                    &det.Estatus,
                    &det.Comentarios,
                    &det.FechaRegistro,
                )
                if err != nil {
                    detallesRows.Close()
                    writeErrorResponse(w, http.StatusInternalServerError, "Error al leer detalle", err.Error())
                    return
                }
                var fechaRegistroStr string
                if det.FechaRegistro.Valid {
                    fechaRegistroStr = det.FechaRegistro.String
                } else {
                    fechaRegistroStr = ""
                }
                detalles = append(detalles, map[string]interface{}{
                    "id_detalle":           det.IDDetalle,
                    "id_pedido":            det.IDPedido,
                    "id_producto":          det.IDProducto,
                    "clave_producto":       det.ClaveProducto,
                    "descripcion":          det.Descripcion,
                    "unidad":               det.Unidad,
                    "cantidad":             det.Cantidad,
                    "precio_unitario":      det.PrecioUnitario,
                    "porcentaje_descuento": det.PorcentajeDescuento,
                    "importe_descuento":    det.ImporteDescuento,
                    "subtotal":             det.Subtotal,
                    "importe_iva":          det.ImporteIVA,
                    "importe_ieps":         det.ImporteIEPS,
                    "total":                det.Total,
                    "latitud_entrega":      NullFloatToPtr(det.LatitudEntrega),
                    "longitud_entrega":     NullFloatToPtr(det.LongitudEntrega),
                    "estatus":              det.Estatus,
                    "comentarios":          NullToStr(det.Comentarios),
                    "fecha_registro":       fechaRegistroStr,
                })
            }
            detallesRows.Close()
            pedidoJSON := map[string]interface{}{
                "pedido":   pedidoMap,
                "detalles": detalles,
            }
            pedidos = append(pedidos, pedidoJSON)
        }
        writeSuccessResponse(w, "Pedidos obtenidos correctamente", pedidos)
    }
}

// ---------------------------
// ENDPOINT: Obtener pedidos por id_usuario
// ---------------------------
func GetPedidosByUsuario(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        if idUsuarioStr == "" {
            writeErrorResponse(w, http.StatusBadRequest, "Falta el parámetro id_usuario", "")
            return
        }
        var idUsuario int
        if _, err := fmt.Sscanf(idUsuarioStr, "%d", &idUsuario); err != nil || idUsuario <= 0 {
            writeErrorResponse(w, http.StatusBadRequest, "id_usuario inválido", "")
            return
        }

        rows, err := dbc.Local.Query(`
            SELECT id_pedido, clave_unica, id_usuario, id_tienda, id_sucursal, fecha_creacion, fecha_entrega,
                   subtotal, descuento, iva, ieps, total, id_metodo_pago, referencia_pago, direccion_entrega,
                   colonia_entrega, cp_entrega, ciudad_entrega, estado_entrega, latitud_entrega, longitud_entrega,
                   estatus, comentarios, origen_pedido, id_lista_precio
            FROM pedidos
            WHERE id_usuario = ?
            ORDER BY fecha_creacion DESC
        `, idUsuario)
        if err != nil {
            writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener los pedidos", err.Error())
            return
        }
        defer rows.Close()

        var pedidos []map[string]interface{}
        for rows.Next() {
            var pedido Pedido
            var fechaEntregaStr sql.NullString

            err := rows.Scan(
                &pedido.IDPedido,
                &pedido.ClaveUnica,
                &pedido.IDUsuario,
                &pedido.IDTienda,
                &pedido.IDSucursal,
                &pedido.FechaCreacion,
                &fechaEntregaStr,
                &pedido.Subtotal,
                &pedido.Descuento,
                &pedido.IVA,
                &pedido.IEPS,
                &pedido.Total,
                &pedido.IDMetodoPago,
                &pedido.ReferenciaPago,
                &pedido.DireccionEntrega,
                &pedido.ColoniaEntrega,
                &pedido.CPEntrega,
                &pedido.CiudadEntrega,
                &pedido.EstadoEntrega,
                &pedido.LatitudEntrega,
                &pedido.LongitudEntrega,
                &pedido.Estatus,
                &pedido.Comentarios,
                &pedido.OrigenPedido,
                &pedido.IDListaPrecio,
            )
            if err != nil {
                writeErrorResponse(w, http.StatusInternalServerError, "Error al leer pedido", err.Error())
                return
            }
            if fechaEntregaStr.Valid {
                t, err := time.Parse("2006-01-02 15:04:05", fechaEntregaStr.String)
                if err == nil {
                    pedido.FechaEntrega = sql.NullTime{
                        Time: t,
                        Valid: true,
                    }
                }
            }
            fechaCreacionTime, _ := time.Parse("2006-01-02 15:04:05", pedido.FechaCreacion)
            pedidoMap := map[string]interface{}{
                "id_pedido":         pedido.IDPedido,
                "clave_unica":       pedido.ClaveUnica,
                "id_usuario":        pedido.IDUsuario,
                "id_tienda":         pedido.IDTienda,
                "id_sucursal":       pedido.IDSucursal,
                "fecha_creacion":    formateaFechaCorta(fechaCreacionTime),
                "fecha_entrega":     NullTimeToStrMes(pedido.FechaEntrega),
                "subtotal":          pedido.Subtotal,
                "descuento":         pedido.Descuento,
                "iva":               pedido.IVA,
                "ieps":              pedido.IEPS,
                "total":             pedido.Total,
                "id_metodo_pago":    pedido.IDMetodoPago,
                "referencia_pago":   NullToStr(pedido.ReferenciaPago),
                "direccion_entrega": pedido.DireccionEntrega,
                "colonia_entrega":   pedido.ColoniaEntrega,
                "cp_entrega":        pedido.CPEntrega,
                "ciudad_entrega":    pedido.CiudadEntrega,
                "estado_entrega":    pedido.EstadoEntrega,
                "latitud_entrega":   NullFloatToPtr(pedido.LatitudEntrega),
                "longitud_entrega":  NullFloatToPtr(pedido.LongitudEntrega),
                "estatus":           pedido.Estatus,
                "comentarios":       NullToStr(pedido.Comentarios),
                "origen_pedido":     pedido.OrigenPedido,
                "id_lista_precio":   pedido.IDListaPrecio,
            }

            detallesRows, err := dbc.Local.Query(`
                SELECT id_detalle, id_pedido, id_producto, clave_producto, descripcion, unidad, cantidad, precio_unitario, 
                       porcentaje_descuento, importe_descuento, subtotal, importe_iva, importe_ieps, total, latitud_entrega, 
                       longitud_entrega, estatus, comentarios, fecha_registro
                FROM detalle_pedidos WHERE id_pedido = ?`, pedido.IDPedido)
            if err != nil {
                writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener detalles", err.Error())
                return
            }
            var detalles []map[string]interface{}
            for detallesRows.Next() {
                var det DetallePedido
                err := detallesRows.Scan(
                    &det.IDDetalle,
                    &det.IDPedido,
                    &det.IDProducto,
                    &det.ClaveProducto,
                    &det.Descripcion,
                    &det.Unidad,
                    &det.Cantidad,
                    &det.PrecioUnitario,
                    &det.PorcentajeDescuento,
                    &det.ImporteDescuento,
                    &det.Subtotal,
                    &det.ImporteIVA,
                    &det.ImporteIEPS,
                    &det.Total,
                    &det.LatitudEntrega,
                    &det.LongitudEntrega,
                    &det.Estatus,
                    &det.Comentarios,
                    &det.FechaRegistro,
                )
                if err != nil {
                    detallesRows.Close()
                    writeErrorResponse(w, http.StatusInternalServerError, "Error al leer detalle", err.Error())
                    return
                }
                fechaRegistroStr := ""
                if det.FechaRegistro.Valid {
                    fechaRegistroStr = det.FechaRegistro.String
                }
                detalles = append(detalles, map[string]interface{}{
                    "id_detalle":           det.IDDetalle,
                    "id_pedido":            det.IDPedido,
                    "id_producto":          det.IDProducto,
                    "clave_producto":       det.ClaveProducto,
                    "descripcion":          det.Descripcion,
                    "unidad":               det.Unidad,
                    "cantidad":             det.Cantidad,
                    "precio_unitario":      det.PrecioUnitario,
                    "porcentaje_descuento": det.PorcentajeDescuento,
                    "importe_descuento":    det.ImporteDescuento,
                    "subtotal":             det.Subtotal,
                    "importe_iva":          det.ImporteIVA,
                    "importe_ieps":         det.ImporteIEPS,
                    "total":                det.Total,
                    "latitud_entrega":      NullFloatToPtr(det.LatitudEntrega),
                    "longitud_entrega":     NullFloatToPtr(det.LongitudEntrega),
                    "estatus":              det.Estatus,
                    "comentarios":          NullToStr(det.Comentarios),
                    "fecha_registro":       fechaRegistroStr,
                })
            }
            detallesRows.Close()
            pedidoJSON := map[string]interface{}{
                "pedido":   pedidoMap,
                "detalles": detalles,
            }
            pedidos = append(pedidos, pedidoJSON)
        }
        writeSuccessResponse(w, "Pedidos del usuario obtenidos correctamente", pedidos)
    }
}

// ---------------------------
// ENDPOINT: Obtener fechas disponibles para entrega
// ---------------------------
func GetFechasEntregaDisponibles(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        config, err := ObtenerConfigEntrega(dbc)
        if err != nil {
            writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener configuración de entregas", err.Error())
            return
        }
        now := time.Now()
        fechasDisponibles := []map[string]interface{}{}
        fechaActual := now
        diasCalculados := 0
        for diasCalculados < 7 {
            fechaActual = fechaActual.AddDate(0, 0, 1)
            if esDiaHabil(fechaActual, config) {
                esFeriado, _ := esDiaFeriado(fechaActual, config)
                if !esFeriado {
                    for _, horario := range config.HorariosEntrega {
                        horaParts := strings.Split(horario.Inicio, ":")
                        if len(horaParts) >= 2 {
                            hora, _ := strconv.Atoi(horaParts[0])
                            minuto, _ := strconv.Atoi(horaParts[1])
                            fechaHora := time.Date(
                                fechaActual.Year(),
                                fechaActual.Month(),
                                fechaActual.Day(),
                                hora, minuto, 0, 0,
                                fechaActual.Location(),
                            )
                            fechasDisponibles = append(fechasDisponibles, map[string]interface{}{
                                "fecha": fechaHora.Format("2006-01-02 15:04:05"),
                                "fecha_formateada": formateaFechaCorta(fechaHora),
                                "etiqueta": horario.Etiqueta,
                                "timestamp": fechaHora.Unix(),
                            })
                        }
                    }
                    diasCalculados++
                }
            }
        }
        writeSuccessResponse(w, "Fechas de entrega disponibles", fechasDisponibles)
    }
}

// ---------------------------
// ENDPOINT: Crear Pedido (ajustado a detalle_pedidos)
// ---------------------------
func CreatePedido(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req PedidoRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            writeErrorResponse(w, http.StatusBadRequest, "Datos inválidos", err.Error())
            return
        }

        if req.IDSucursal == 0 {
            var idSucursal int
            err := dbc.Local.QueryRow("SELECT idsucursal FROM adm_sucursales WHERE estatus = 'S' ORDER BY idsucursal LIMIT 1").Scan(&idSucursal)
            if err != nil {
                writeErrorResponse(w, http.StatusInternalServerError, "No se pudo obtener la sucursal", err.Error())
                return
            }
            req.IDSucursal = idSucursal
        }

        if req.IDUsuario == 0 || req.IDTienda == 0 || req.IDSucursal == 0 || req.IDMetodoPago == 0 {
            writeErrorResponse(w, http.StatusBadRequest, "Campos requeridos faltan", "IDUsuario, IDTienda, IDSucursal, IDMetodoPago son obligatorios")
            return
        }
        if len(req.Detalles) == 0 {
            writeErrorResponse(w, http.StatusBadRequest, "Debe incluir al menos un producto", "")
            return
        }

        tx, err := dbc.Local.Begin()
        if err != nil {
            writeErrorResponse(w, http.StatusInternalServerError, "No se pudo iniciar la transacción", err.Error())
            return
        }
        defer func() {
            if r := recover(); r != nil {
                tx.Rollback()
                writeErrorResponse(w, http.StatusInternalServerError, "Error inesperado en el servidor", fmt.Sprintf("%v", r))
            }
        }()

        now := time.Now()

        if !req.FechaEntrega.Valid {
            config, err := ObtenerConfigEntrega(dbc)
            if err != nil {
                writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener configuración de entregas", err.Error())
                return
            }
            fechaEntrega := CalcularFechaEntrega(now, config)
            req.FechaEntrega = sql.NullTime{
                Time:  fechaEntrega,
                Valid: true,
            }
        }

        columnas := []string{
            "clave_unica", "id_usuario", "id_tienda", "id_sucursal", "fecha_creacion", "fecha_entrega",
            "subtotal", "descuento", "iva", "ieps", "total", "id_metodo_pago", "referencia_pago",
            "direccion_entrega", "colonia_entrega", "cp_entrega", "ciudad_entrega", "estado_entrega",
            "latitud_entrega", "longitud_entrega",
            "estatus", "comentarios", "origen_pedido", "id_lista_precio",
        }
        placeholders := strings.Repeat("?,", len(columnas))
        placeholders = strings.TrimSuffix(placeholders, ",")

        query := fmt.Sprintf("INSERT INTO pedidos (%s) VALUES (%s)", strings.Join(columnas, ","), placeholders)

        valores := []interface{}{
            fmt.Sprintf("PED-%d", now.UnixNano()),
            req.IDUsuario,
            req.IDTienda,
            req.IDSucursal,
            now.Format("2006-01-02 15:04:05"),
            NullTimeToStrOrNil(req.FechaEntrega),
            0.0, 0.0, 0.0, 0.0, 0.0,
            req.IDMetodoPago,
            NullToStrOrNil(req.ReferenciaPago),
            req.DireccionEntrega,
            req.ColoniaEntrega,
            req.CPEntrega,
            req.CiudadEntrega,
            req.EstadoEntrega,
            req.LatitudEntrega,
            req.LongitudEntrega,
            "pendiente",
            NullToStrOrNil(req.Comentarios),
            req.OrigenPedido,
            1,
        }

        if len(columnas) != len(valores) {
            errMsg := fmt.Sprintf("Columnas y valores no coinciden: columnas=%d, valores=%d\nColumnas: %v\nValores: %v", len(columnas), len(valores), columnas, valores)
            writeErrorResponse(w, http.StatusInternalServerError, "Error interno: columnas y valores no coinciden", errMsg)
            tx.Rollback()
            return
        }

        result, err := tx.Exec(query, valores...)
        if err != nil {
            detalle := fmt.Sprintf("QUERY: %s\nColumnas: %v\nValores: %v\nError: %s", query, columnas, valores, err.Error())
            writeErrorResponse(w, http.StatusInternalServerError, "Error al crear el pedido", detalle)
            tx.Rollback()
            return
        }

        idPedido, _ := result.LastInsertId()
        var subtotal, totalDescuento, totalIVA, totalIEPS, total float64

        stmt, err := tx.Prepare(`
            INSERT INTO detalle_pedidos (
                id_pedido, id_producto, clave_producto, descripcion, unidad, cantidad,
                precio_unitario, porcentaje_descuento, importe_descuento, subtotal,
                importe_iva, importe_ieps, total, latitud_entrega, longitud_entrega, estatus, comentarios, fecha_registro
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `)
        if err != nil {
            tx.Rollback()
            writeErrorResponse(w, http.StatusInternalServerError, "Error al preparar el statement de detalle", err.Error())
            return
        }
        defer stmt.Close()

        for _, d := range req.Detalles {
            if d.PrecioUnitario <= 0 || d.Cantidad <= 0 {
                tx.Rollback()
                writeErrorResponse(w, http.StatusBadRequest, "Precio o cantidad no válida", "Precio y cantidad deben ser mayores a 0")
                return
            }

            // Consulta el IVA y suma de IEPS desde la base
            var porcentajeIVA, porcentajeIEPS float64
            err := dbc.Local.QueryRow(`
                SELECT IFNULL(i.iva, 0), IFNULL(i.ieps1, 0) + IFNULL(i.ieps2, 0) + IFNULL(i.ieps3, 0)
                FROM crm_productos p
                LEFT JOIN crm_impuestos i ON p.idiva = i.idiva
                WHERE p.idproducto = ?
            `, d.IDProducto).Scan(&porcentajeIVA, &porcentajeIEPS)
            if err != nil {
                tx.Rollback()
                writeErrorResponse(w, http.StatusInternalServerError, "No se pudo obtener impuestos del producto", err.Error())
                return
            }

            // Calcular precio neto (precio original sin IVA/IEPS)
            precioConIVA := d.PrecioUnitario // Si te llega con IVA, hay que sacar el neto
            precioNeto := precioConIVA / (1 + (porcentajeIVA/100) + (porcentajeIEPS/100))
            subt := precioNeto * d.Cantidad
            importeDescuento := d.ImporteDescuento
            porcentajeDescuento := d.PorcentajeDescuento
            if porcentajeDescuento > 0 {
                importeDescuento = subt * (porcentajeDescuento / 100)
            }
            ivaImporte := subt * (porcentajeIVA / 100)
            iepsImporte := subt * (porcentajeIEPS / 100)
            tot := subt - importeDescuento + ivaImporte + iepsImporte

            _, err = stmt.Exec(
                idPedido,
                d.IDProducto,
                d.ClaveProducto,
                d.Descripcion,
                d.Unidad,
                d.Cantidad,
                precioNeto, // <--- precio_unitario: el precio SIN IVA/IEPS
                porcentajeDescuento,
                importeDescuento,
                subt,
                ivaImporte,
                iepsImporte,
                tot,
                d.LatitudEntrega,
                d.LongitudEntrega,
                "solicitado",
                d.Comentarios,
                now.Format("2006-01-02 15:04:05"),
            )
            if err != nil {
                tx.Rollback()
                writeErrorResponse(w, http.StatusInternalServerError, "Error al insertar detalle", err.Error())
                return
            }
            subtotal += subt
            totalDescuento += importeDescuento
            totalIVA += ivaImporte
            totalIEPS += iepsImporte
            total += tot
        }

        _, err = tx.Exec(`
            UPDATE pedidos SET
                subtotal = ?, descuento = ?, iva = ?, ieps = ?, total = ?
            WHERE id_pedido = ?
        `, subtotal, totalDescuento, totalIVA, totalIEPS, total, idPedido)
        if err != nil {
            tx.Rollback()
            writeErrorResponse(w, http.StatusInternalServerError, "Error al actualizar el pedido", err.Error())
            return
        }

        // Indicadores diarios/mensuales
        fechaStr := now.Format("2006-01-02 15:04:05")
        if err := AcumulaIndicadores(tx, total, req.IDUsuario, fechaStr); err != nil {
            tx.Rollback()
            writeErrorResponse(w, http.StatusInternalServerError, "Error al actualizar indicadores diarios/mensuales", err.Error())
            return
        }

        if err := tx.Commit(); err != nil {
            tx.Rollback()
            writeErrorResponse(w, http.StatusInternalServerError, "Error al confirmar el pedido", err.Error())
            return
        }

        writeSuccessResponse(w, "Pedido creado exitosamente", map[string]interface{}{
            "id_pedido":    idPedido,
            "total":        round(total, 2),
            "fecha_entrega": NullTimeToStrMes(req.FechaEntrega),
        })
    }
}
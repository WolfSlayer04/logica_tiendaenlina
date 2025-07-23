package db

import (
    "database/sql"
    "time"

)


// Estructura Pedido para pedidos y handlers de pedidos
type Pedido struct {
    IDPedido        int64   `json:"id_pedido"`
    ClaveUnica      string  `json:"clave_unica"`
    IDUsuario       int     `json:"id_usuario"`
    IDTienda        int     `json:"id_tienda"`
    IDSucursal      int     `json:"id_sucursal"`
    FechaCreacion   string  `json:"fecha_creacion"`   // Cambiado a string
    FechaEntrega    string  `json:"fecha_entrega"`    // Opcional: también como string si da problemas
    Subtotal        float64 `json:"subtotal"`
    Descuento       float64 `json:"descuento"`
    IVA             float64 `json:"iva"`
    IEPS            float64 `json:"ieps"`
    Total           float64 `json:"total"`
    IDMetodoPago    int     `json:"id_metodo_pago"`
    ReferenciaPago  string  `json:"referencia_pago"`
    DireccionEntrega string `json:"direccion_entrega"`
    ColoniaEntrega   string `json:"colonia_entrega"`
    CPEntrega        string `json:"cp_entrega"`
    CiudadEntrega    string `json:"ciudad_entrega"`
    EstadoEntrega    string `json:"estado_entrega"`
    Estatus          string `json:"estatus"`
    Comentarios      string `json:"comentarios"`
    OrigenPedido     string `json:"origen_pedido"`
    IDListaPrecio    int    `json:"id_lista_precio"`
}

// Agrega también tu struct DetallePedido si lo usas, ejemplo:
type DetallePedido struct {
    IDDetalle        int             `json:"id_detalle"`
    Descripcion      string          `json:"descripcion"`
    Cantidad         float64         `json:"cantidad"`
    PrecioUnitario   float64         `json:"precio_unitario"`
    Subtotal         float64         `json:"subtotal"`
    ImporteDescuento float64         `json:"descuento"`
    ImporteIVA       float64         `json:"iva"`
    ImporteIEPS      float64         `json:"ieps"`
    Total            float64         `json:"total"`
    Comentarios      sql.NullString  `json:"comentarios"`
}

type Usuario struct {
    IDUsuario            int            `json:"id_usuario"`
    IDEmpresa            int            `json:"id_empresa"`
    TipoUsuario          string         `json:"tipo_usuario"`
    NombreCompleto       string         `json:"nombre_completo"`
    Correo               string         `json:"correo"`
    Telefono             string         `json:"telefono"`
    Clave                string         `json:"clave"`
    FechaRegistro        time.Time      `json:"fecha_registro"`
    UltimoAcceso         sql.NullTime   `json:"ultimo_acceso"`
    Estatus              string         `json:"estatus"`
    RequiereCambiarClave bool           `json:"requiere_cambiar_clave"`
}
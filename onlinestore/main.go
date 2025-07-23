package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "onlinestore/db"
    "onlinestore/rutas"
    "onlinestore/Middleware"

    "github.com/gorilla/mux"
    "github.com/rs/cors"
)

func main() {
    // Usa el singleton para obtener la conexión
    dbConn, err := db.GetDBConnection()
    if err != nil {
        log.Fatalf("Error al conectar a las bases de datos: %v", err)
    }

    if err := dbConn.CheckConnections(); err != nil {
        log.Fatalf("Error verificando conexiones: %v", err)
    }
    fmt.Println("Conexión a las bases de datos establecida correctamente")

    // Iniciar el sincronizador de pedidos (diablito)
    rutas.IniciarSincronizadorPedidos(dbConn)

    // -------- INICIALIZA INDICADORES HISTÓRICOS (solo la primera vez) --------
    if err := rutas.InicializaIndicadoresHistoricos(dbConn); err != nil {
        log.Fatalf("Error inicializando indicadores históricos: %v", err)
    }
    // -------------------------------------------------------------------------

    r := mux.NewRouter()

    setupRoutes(r, dbConn)

    handler := cors.AllowAll().Handler(r)

    port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}

fmt.Printf("Servidor corriendo en http://0.0.0.0:%s\n", port)
log.Fatal(http.ListenAndServe(":"+port, handler))

}

func setupRoutes(r *mux.Router, dbConn *db.DBConnection) {
    // Rutas para categorías
    r.HandleFunc("/api/categorias", rutas.GetCategorias(dbConn)).Methods("GET")
    // Rutas para productos
    r.HandleFunc("/api/productos", rutas.GetProductos(dbConn)).Methods("GET")
    r.HandleFunc("/api/productos/iva", rutas.GetProductosConIVA(dbConn)).Methods("GET")
    r.HandleFunc("/api/productos/iva/categoria/{idcategoria}", rutas.GetProductosConIVAPorCategoria(dbConn)).Methods("GET")
    r.HandleFunc("/api/productos/buscar", rutas.SearchProductos(dbConn)).Methods("GET")
    r.HandleFunc("/api/productos/categoria/{idcategoria}", rutas.GetProductosByCategoria(dbConn)).Methods("GET")
    r.HandleFunc("/api/productos/sugerencias", rutas.GetProductoSuggestions(dbConn)).Methods("GET")
    r.HandleFunc("/api/productos/iva/buscar", rutas.GetProductosConIVABuscar(dbConn)).Methods("GET")
    // ADMIN productos
    r.HandleFunc("/api/productos", rutas.AddProducto(dbConn)).Methods("POST")
    r.HandleFunc("/api/productos/estatus", rutas.GetEstatusProductos(dbConn)).Methods("GET")
    r.HandleFunc("/api/productos/{idproducto}", rutas.EditProducto(dbConn)).Methods("PUT")
    r.HandleFunc("/api/productos/{id}", rutas.GetProductoByID(dbConn)).Methods("GET")
    // ENDPOINT DE IVAS POR EMPRESA
    r.HandleFunc("/api/impuestos", rutas.GetImpuestosPorEmpresa(dbConn)).Methods("GET")
    // Rutas para carrito
    r.HandleFunc("/api/carrito/agregar", rutas.AddToCart(dbConn)).Methods("POST")
    r.HandleFunc("/api/carrito", rutas.GetCart(dbConn)).Methods("GET")
    r.HandleFunc("/api/carrito/vaciar", rutas.ClearCart(dbConn)).Methods("DELETE")
    r.HandleFunc("/api/carrito/eliminar/{idproducto}", rutas.RemoveFromCart(dbConn)).Methods("DELETE")
    r.HandleFunc("/api/carrito/actualizar", rutas.UpdateCartItem(dbConn)).Methods("PUT")
    // Rutas para pedidos
    r.HandleFunc("/api/pedidos", rutas.CreatePedido(dbConn)).Methods("POST")
    r.HandleFunc("/api/pedidos", rutas.AdminGetPedidosConDetalles(dbConn)).Methods("GET")
    r.HandleFunc("/api/pedidos/usuario", rutas.GetPedidosByUsuario(dbConn)).Methods("GET")
    // Usuarios
    r.HandleFunc("/api/usuarios", rutas.GetUsuarios(dbConn)).Methods("GET")
    // Tiendas
    r.HandleFunc("/api/tiendas", rutas.GetTiendas(dbConn)).Methods("GET")
    // Registro combinado usuario+tienda
    r.HandleFunc("/api/registro", rutas.RegistroUsuarioTienda(dbConn)).Methods("POST")
    // Login
    r.HandleFunc("/api/login", rutas.LoginUsuario(dbConn)).Methods("POST")
    // Obtener tienda por usuario
    r.HandleFunc("/api/tiendas/por_usuario", rutas.GetTiendaByUsuario(dbConn)).Methods("GET")
    // Perfil de usuario y edición
    r.HandleFunc("/api/perfil", rutas.GetPerfilUsuario(dbConn)).Methods("GET")
    r.HandleFunc("/api/usuarios/editar", rutas.EditarUsuario(dbConn)).Methods("PUT")
    r.HandleFunc("/api/tiendas/eliminar", rutas.EliminarTienda(dbConn)).Methods("DELETE")
    // Indicadores diarios y mensuales
    r.HandleFunc("/api/indicadores/diario", rutas.GetIndicadoresDiarioAll(dbConn)).Methods("GET")
    r.HandleFunc("/api/indicadores/diario/fecha", rutas.GetIndicadorDiarioByFecha(dbConn)).Methods("GET")
    r.HandleFunc("/api/indicadores/mensual", rutas.GetIndicadoresMensualAll(dbConn)).Methods("GET")
    r.HandleFunc("/api/indicadores/mensual/fecha", rutas.GetIndicadorMensualByFecha(dbConn)).Methods("GET")
    r.HandleFunc("/api/dashboard/stats", rutas.GetDashboardStats(dbConn)).Methods("GET")
    // Administración de pedidos
    r.HandleFunc("/api/pedidos/{id_pedido}/sucursal", rutas.AdminActualizarSucursalPedido(dbConn)).Methods("PUT")
    r.HandleFunc("/api/pedidos/{id_pedido}/estatus", rutas.AdminActualizarEstatusPedido(dbConn)).Methods("PUT")
    r.HandleFunc("/api/pedidos/{id_pedido}/descuento", rutas.AdminAplicarDescuentoPedido(dbConn)).Methods("PUT")
    r.HandleFunc("/api/pedidos/{id_pedido}/detalles", rutas.AdminActualizarDetallesPedido(dbConn)).Methods("PUT")
    r.HandleFunc("/api/pedidos/{id_pedido}", rutas.AdminGetPedidoByID(dbConn)).Methods("GET")
    // ADMIN clientes (usuarios + tienda)
    r.HandleFunc("/api/admin/clientes", rutas.GetAllUsuariosConTienda(dbConn)).Methods("GET")
    // ADMIN usuarios (admin_usuarios)
    r.HandleFunc("/api/admin/usuarios", rutas.GetAllAdminUsuarios(dbConn)).Methods("GET")
    r.Handle("/api/admin/usuarios", middlewares.RequireAdmin(http.HandlerFunc(rutas.CreateAdminUsuario(dbConn)))).Methods("POST")
    r.Handle("/api/admin/usuarios", middlewares.RequireAdmin(http.HandlerFunc(rutas.UpdateAdminUsuario(dbConn)))).Methods("PUT")
    r.Handle("/api/admin/usuarios", rutas.DeleteAdminUsuario(dbConn)).Methods("DELETE")

    // ADMIN personalizar empresa
    r.HandleFunc("/api/admin/personalizar", rutas.AdminGetAllPersonalizaciones(dbConn)).Methods("GET")
    r.HandleFunc("/api/admin/personalizar/{id}", rutas.AdminGetPersonalizacionByID(dbConn)).Methods("GET")
    r.Handle("/api/admin/personalizar", middlewares.RequireAdmin(http.HandlerFunc(rutas.AdminCreatePersonalizacion(dbConn)))).Methods("POST")
    r.Handle("/api/admin/personalizar/{id}", middlewares.RequireAdmin(http.HandlerFunc(rutas.AdminUpdatePersonalizacionByID(dbConn)))).Methods("PUT")
    r.Handle("/api/admin/personalizar/{id}", middlewares.RequireAdmin(http.HandlerFunc(rutas.AdminDeletePersonalizacionByID(dbConn)))).Methods("DELETE")

    // Logo empresa
    r.Handle("/api/empresa/logo", middlewares.RequireAdmin(http.HandlerFunc(rutas.EmpresaUploadLogo(dbConn)))).Methods("POST")
    r.Handle("/api/empresa/logo", middlewares.RequireAdmin(http.HandlerFunc(rutas.EmpresaUpdateLogo(dbConn)))).Methods("PUT")
    r.Handle("/api/empresa/logo", middlewares.RequireAdmin(http.HandlerFunc(rutas.EmpresaDeleteLogo(dbConn)))).Methods("DELETE")
    r.HandleFunc("/api/empresa/logo", rutas.EmpresaGetLogo(dbConn)).Methods("GET")

    // Sincronización de pedidos con sistema principal
    r.Handle("/api/pedidos/sincronizar", middlewares.RequireAdmin(http.HandlerFunc(rutas.SincronizarPedido(dbConn)))).Methods("POST")
    r.HandleFunc("/api/pedidos/verificar_sincronizacion", rutas.VerificarSincronizacion(dbConn)).Methods("GET")
    r.HandleFunc("/api/pedidos/pendientes_sincronizacion", rutas.PedidosPendientesSincronizacion(dbConn)).Methods("GET")
    r.Handle("/api/pedidos/actualizar_fecha_entrega", middlewares.RequireAdmin(http.HandlerFunc(rutas.ActualizarFechaEntrega(dbConn)))).Methods("POST")
    r.HandleFunc("/api/pedidos/obtener_por_id_remoto", rutas.ObtenerPedidoPorIDRemoto(dbConn)).Methods("GET")

    // ADMIN configuración de entregas
    r.HandleFunc("/api/admin/config-entrega", rutas.GetConfigEntrega(dbConn)).Methods("GET")
    r.Handle("/api/admin/config-entrega", middlewares.RequireAdmin(http.HandlerFunc(rutas.UpdateConfigEntrega(dbConn)))).Methods("POST")

    // Nuevo endpoint para fechas de entrega disponibles
    r.HandleFunc("/api/fechas-entrega-disponibles", rutas.GetFechasEntregaDisponibles(dbConn)).Methods("GET")

    // --- SUCURSALES: lista de precios y productos según sucursal ---
    r.HandleFunc("/api/sucursales", rutas.GetSucursalALL(dbConn)).Methods("GET")
    r.HandleFunc("/api/sucursales/{id}", rutas.GetSucursal(dbConn)).Methods("GET")
    r.Handle("/api/sucursales/{id}/lista-precios", middlewares.RequireAdmin(http.HandlerFunc(rutas.UpdateListaPreciosSucursal(dbConn)))).Methods("PUT")
    r.HandleFunc("/api/sucursales/{id}/productos", rutas.GetProductosSucursal(dbConn)).Methods("GET")

    // NUEVA RUTA GET para consultar la lista de precios de una sucursal
    r.HandleFunc("/api/sucursales/{id}/lista-precios", rutas.GetListaPreciosSucursal(dbConn)).Methods("GET")
}
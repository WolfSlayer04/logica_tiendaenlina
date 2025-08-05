package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	middlewares "github.com/WolfSlayer04/logica_tiendaenlina/Middleware"
	"github.com/WolfSlayer04/logica_tiendaenlina/db"
	"github.com/WolfSlayer04/logica_tiendaenlina/rutas"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	dbConn, err := db.GetDBConnection()
	if err != nil {
		log.Fatalf("Error al conectar a las bases de datos: %v", err)
	}

	if err := dbConn.CheckConnections(); err != nil {
		log.Fatalf("Error verificando conexiones: %v", err)
	}
	fmt.Println("Conexión a las bases de datos establecida correctamente")

	rutas.IniciarSincronizadorPedidos(dbConn)

	if err := rutas.InicializaIndicadoresHistoricos(dbConn); err != nil {
		log.Fatalf("Error inicializando indicadores históricos: %v", err)
	}

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
	// Rutas públicas
	r.HandleFunc("/api/registro", rutas.RegistroUsuarioTienda(dbConn)).Methods("POST")
	r.HandleFunc("/api/login", rutas.LoginUsuario(dbConn)).Methods("POST")
	r.HandleFunc("/api/empresa/logo", rutas.EmpresaGetLogo(dbConn)).Methods("GET")
	r.HandleFunc("/api/refresh", rutas.RefreshTokenEndpoint(dbConn)).Methods("POST")
	r.Handle("/api/admin/personalizar", rutas.AdminGetAllPersonalizaciones(dbConn)).Methods("GET")
	r.Handle("/api/admin/personalizar/{id}",rutas.AdminGetPersonalizacionByID(dbConn)).Methods("GET")

	// Protegidas con JWT
	r.Handle("/api/categorias", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetCategorias(dbConn)))).Methods("GET")

	// PRODUCTOS
	r.Handle("/api/productos/estatus", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetEstatusProductos(dbConn)))).Methods("GET")
	r.Handle("/api/productos/iva", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetProductosConIVA(dbConn)))).Methods("GET")
	r.Handle("/api/productos/iva/buscar", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetProductosConIVABuscar(dbConn)))).Methods("GET")
	r.Handle("/api/productos/buscar", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.SearchProductos(dbConn)))).Methods("GET")
	r.Handle("/api/productos/sugerencias", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetProductoSuggestions(dbConn)))).Methods("GET")
	r.Handle("/api/productos", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetProductos(dbConn)))).Methods("GET")
	r.Handle("/api/productos", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.AddProducto(dbConn))))).Methods("POST")

	r.Handle("/api/productos/iva/categoria/{idcategoria}", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetProductosConIVAPorCategoria(dbConn)))).Methods("GET")
	r.Handle("/api/productos/categoria/{idcategoria}", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetProductosByCategoria(dbConn)))).Methods("GET")
	r.Handle("/api/productos/{id}", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetProductoByID(dbConn)))).Methods("GET")
	r.Handle("/api/productos/{idproducto}", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.EditProducto(dbConn))))).Methods("PUT")

	// Impuestos por empresa
	r.Handle("/api/impuestos", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetImpuestosPorEmpresa(dbConn)))).Methods("GET")

	// Carrito
	r.Handle("/api/carrito/agregar", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.AddToCart(dbConn)))).Methods("POST")
	r.Handle("/api/carrito", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetCart(dbConn)))).Methods("GET")
	r.Handle("/api/carrito/vaciar", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.ClearCart(dbConn)))).Methods("DELETE")
	r.Handle("/api/carrito/eliminar/{idproducto}", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.RemoveFromCart(dbConn)))).Methods("DELETE")
	r.Handle("/api/carrito/actualizar", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.UpdateCartItem(dbConn)))).Methods("PUT")

	// Pedidos
	r.Handle("/api/pedidos", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.CreatePedido(dbConn)))).Methods("POST")
	r.Handle("/api/pedidos", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.AdminGetPedidosConDetallesPaginado(dbConn)))).Methods("GET")
	r.Handle("/api/pedidos/usuario", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetPedidosByUsuario(dbConn)))).Methods("GET")

	// Usuarios/tiendas
	r.Handle("/api/usuarios", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetUsuarios(dbConn)))).Methods("GET")
	r.Handle("/api/tiendas", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetTiendas(dbConn)))).Methods("GET")
	r.Handle("/api/tiendas/por_usuario", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetTiendaByUsuario(dbConn)))).Methods("GET")
	r.Handle("/api/perfil", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetPerfilUsuario(dbConn)))).Methods("GET")
	r.Handle("/api/usuarios/editar", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.EditarUsuario(dbConn)))).Methods("PUT")
	r.Handle("/api/tiendas/eliminar", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.EliminarTienda(dbConn)))).Methods("DELETE")

	// Indicadores
	r.Handle("/api/indicadores/diario", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetIndicadoresDiarioAll(dbConn)))).Methods("GET")
	r.Handle("/api/indicadores/diario/fecha", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetIndicadorDiarioByFecha(dbConn)))).Methods("GET")
	r.Handle("/api/indicadores/mensual", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetIndicadoresMensualAll(dbConn)))).Methods("GET")
	r.Handle("/api/indicadores/mensual/fecha", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetIndicadorMensualByFecha(dbConn)))).Methods("GET")
	r.Handle("/api/dashboard/stats", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetDashboardStats(dbConn)))).Methods("GET")

	// Pedidos administración y sincronización (solo admin)
	r.Handle("/api/pedidos/{id_pedido}/sucursal", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.AdminActualizarSucursalPedido(dbConn))))).Methods("PUT")
	r.Handle("/api/pedidos/{id_pedido}/estatus", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.AdminActualizarEstatusPedido(dbConn))))).Methods("PUT")
	r.Handle("/api/pedidos/{id_pedido}/descuento", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.AdminAplicarDescuentoPedido(dbConn))))).Methods("PUT")
	r.Handle("/api/pedidos/{id_pedido}/detalles", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.AdminActualizarDetallesPedido(dbConn))))).Methods("PUT")
	r.Handle("/api/pedidos/{id_pedido}", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.AdminGetPedidoByID(dbConn))))).Methods("GET")

	// ADMIN clientes
	r.Handle("/api/admin/clientes", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.GetAllUsuariosConTienda(dbConn))))).Methods("GET")

	// ADMIN usuarios
	r.Handle("/api/admin/usuarios", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.GetAllAdminUsuarios(dbConn))))).Methods("GET")
	r.Handle("/api/admin/usuarios", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.CreateAdminUsuario(dbConn))))).Methods("POST")
	r.Handle("/api/admin/usuarios", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.UpdateAdminUsuario(dbConn))))).Methods("PUT")
	r.Handle("/api/admin/usuarios", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.DeleteAdminUsuario(dbConn))))).Methods("DELETE")

	// ADMIN personalizaciones
	r.Handle("/api/admin/personalizar", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.AdminCreatePersonalizacion(dbConn))))).Methods("POST")
	r.Handle("/api/admin/personalizar/{id}", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.AdminUpdatePersonalizacionByID(dbConn))))).Methods("PUT")
	r.Handle("/api/admin/personalizar/{id}", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.AdminDeletePersonalizacionByID(dbConn))))).Methods("DELETE")
	

	// Logo empresa edición (solo admin)
	r.Handle("/api/empresa/logo", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.EmpresaUploadLogo(dbConn))))).Methods("POST")
	r.Handle("/api/empresa/logo", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.EmpresaUpdateLogo(dbConn))))).Methods("PUT")
	r.Handle("/api/empresa/logo", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.EmpresaDeleteLogo(dbConn))))).Methods("DELETE")

	// Sincronización de pedidos
	r.Handle("/api/pedidos/sincronizar", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.SincronizarPedido(dbConn))))).Methods("POST")
	r.Handle("/api/pedidos/verificar_sincronizacion", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.VerificarSincronizacion(dbConn)))).Methods("GET")
	r.Handle("/api/pedidos/pendientes_sincronizacion", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.PedidosPendientesSincronizacion(dbConn)))).Methods("GET")
	r.Handle("/api/pedidos/actualizar_fecha_entrega", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.ActualizarFechaEntrega(dbConn))))).Methods("POST")
	r.Handle("/api/pedidos/obtener_por_id_remoto", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.ObtenerPedidoPorIDRemoto(dbConn)))).Methods("GET")

	// ADMIN config entregas
	r.Handle("/api/admin/config-entrega", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.UpdateConfigEntrega(dbConn))))).Methods("POST")
	r.Handle("/api/admin/config-entrega", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.GetConfigEntrega(dbConn))))).Methods("GET")

	// Fechas de entrega
	r.Handle("/api/fechas-entrega-disponibles", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetFechasEntregaDisponibles(dbConn)))).Methods("GET")

	// Sucursales
	r.Handle("/api/sucursales", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetSucursalALL(dbConn)))).Methods("GET")
	r.Handle("/api/sucursales/{id}", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetSucursal(dbConn)))).Methods("GET")
	r.Handle("/api/sucursales/{id}/lista-precios", middlewares.JWTAuthMiddleware(middlewares.RequireAdminPermisos(http.HandlerFunc(rutas.UpdateListaPreciosSucursal(dbConn))))).Methods("PUT")
	r.Handle("/api/sucursales/{id}/productos", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetProductosSucursal(dbConn)))).Methods("GET")
	r.Handle("/api/sucursales/{id}/lista-precios", middlewares.JWTAuthMiddleware(http.HandlerFunc(rutas.GetListaPreciosSucursal(dbConn)))).Methods("GET")
}
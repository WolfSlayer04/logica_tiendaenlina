package rutas

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"

    "github.com/WolfSlayer04/logica_tiendaenlina/db"
    "github.com/gorilla/mux"
)

// NullString serializa como string o null
type NullString struct {
    sql.NullString
}

func (ns NullString) MarshalJSON() ([]byte, error) {
    if !ns.Valid {
        return []byte("null"), nil
    }
    return json.Marshal(ns.String)
}

// NullFloat64 serializa como float64 o null
type NullFloat64 struct {
    sql.NullFloat64
}

func (nf NullFloat64) MarshalJSON() ([]byte, error) {
    if !nf.Valid {
        return []byte("null"), nil
    }
    return json.Marshal(nf.Float64)
}

// NullInt64 serializa como int64 o null
type NullInt64 struct {
    sql.NullInt64
}

func (ni NullInt64) MarshalJSON() ([]byte, error) {
    if !ni.Valid {
        return []byte("null"), nil
    }
    return json.Marshal(ni.Int64)
}

// Categoria para respuesta de categorías
type Categoria struct {
    IDCategoria int        `json:"idcategoria"`
    Categoria   NullString `json:"categoria"`
    Estatus     string     `json:"estatus"`
}

// Producto para respuesta de productos
type Producto struct {
    IDProducto  int         `json:"idproducto"`
    Descripcion NullString  `json:"descripcion"`
    Precio      NullFloat64 `json:"precio"`
    Estatus     string      `json:"estatus"`
    Categoria   NullString  `json:"categoria"`
}

// ProductoConIVA para productos con impuesto
type ProductoConIVA struct {
    IDProducto  int         `json:"idproducto"`
    Descripcion NullString  `json:"descripcion"`
    Precio      NullFloat64 `json:"precio"`
    Estatus     string      `json:"estatus"`
    Categoria   NullString  `json:"categoria"`
    IDIVA       NullInt64   `json:"idiva"`
    IVA         NullFloat64 `json:"iva"`
    TipoIVA     NullString  `json:"tipo_iva"`
    PrecioFinal float64     `json:"precio_final"`
}

// ProductoImpuesto para consulta individual
type ProductoImpuesto struct {
    IDProducto   int         `json:"idproducto"`
    Descripcion  NullString  `json:"descripcion"`
    PrecioBase   NullFloat64 `json:"precio_base"`
    IDIVA        int         `json:"idiva"`
    IVA          NullFloat64 `json:"iva"`
    TipoIVA      NullString  `json:"tipo_iva"`
    PrecioFinal  float64     `json:"precio_final"`
}

// ProductoSugerencia para autocomplete
type ProductoSugerencia struct {
    IDProducto  int        `json:"idproducto"`
    Nombre      NullString `json:"nombre"`
    Descripcion NullString `json:"descripcion"`
}

// Utilidad para paginación
func getPagination(r *http.Request) (limit, offset int) {
    pageStr := r.URL.Query().Get("page")
    limitStr := r.URL.Query().Get("limit")
    page, _ := strconv.Atoi(pageStr)
    if page < 1 {
        page = 1
    }
    limit, _ = strconv.Atoi(limitStr)
    if limit < 1 {
        limit = 20
    }
    offset = (page - 1) * limit
    return
}

// Utilidad para obtener lista de precios por usuario (CORREGIDO: adm_sucursales)
func getListaPreciosPorUsuario(db *db.DBConnection, idUsuario int) (int, error) {
    var listaPrecios int
    query := `
        SELECT s.lista_precios
        FROM tiendas t
        JOIN adm_sucursales s ON t.idsucursal = s.idsucursal
        WHERE t.id_usuario = ?
    `
    err := db.Local.QueryRow(query, idUsuario).Scan(&listaPrecios)
    return listaPrecios, err
}

// GetProductos obtiene productos paginados con estatus = 'S' y precio según sucursal/usuario
func GetProductos(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        limit, offset := getPagination(r)
        query := fmt.Sprintf(`
            SELECT p.idproducto, p.descripcion, p.precio%d, p.estatus, c.categoria
            FROM crm_productos p
            LEFT JOIN categorias c ON p.idcategoria = c.idcategoria
            WHERE p.estatus = 'S' AND p.precio%d IS NOT NULL AND p.precio%d > 0
            LIMIT ? OFFSET ?
        `, listaPrecios, listaPrecios, listaPrecios)
        rows, err := db.Local.Query(query, limit, offset)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        var productos []Producto
        for rows.Next() {
            var p Producto
            if err := rows.Scan(&p.IDProducto, &p.Descripcion, &p.Precio, &p.Estatus, &p.Categoria); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            productos = append(productos, p)
        }
        countQuery := fmt.Sprintf(`
            SELECT COUNT(*) FROM crm_productos 
            WHERE estatus = 'S' AND precio%d IS NOT NULL AND precio%d > 0
        `, listaPrecios, listaPrecios)
        var total int
        if err := db.Local.QueryRow(countQuery).Scan(&total); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "productos": productos,
            "total":     total,
        })
    }
}

// GetProductosByCategoria obtiene productos filtrados por categoría y paginados y precio sucursal
func GetProductosByCategoria(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        params := mux.Vars(r)
        catIDStr := params["idcategoria"]
        catID, err := strconv.Atoi(catIDStr)
        if err != nil {
            http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
            return
        }
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        limit, offset := getPagination(r)
        query := fmt.Sprintf(`
            SELECT p.idproducto, p.descripcion, p.precio%d, p.estatus, c.categoria
            FROM crm_productos p
            LEFT JOIN categorias c ON p.idcategoria = c.idcategoria
            WHERE p.idcategoria = ? AND p.estatus = 'S' AND p.precio%d IS NOT NULL AND p.precio%d > 0
            LIMIT ? OFFSET ?
        `, listaPrecios, listaPrecios, listaPrecios)
        rows, err := db.Local.Query(query, catID, limit, offset)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        var productos []Producto
        for rows.Next() {
            var p Producto
            if err := rows.Scan(&p.IDProducto, &p.Descripcion, &p.Precio, &p.Estatus, &p.Categoria); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            productos = append(productos, p)
        }
        countQuery := fmt.Sprintf(`
            SELECT COUNT(*) FROM crm_productos 
            WHERE idcategoria = ? AND estatus = 'S' AND precio%d IS NOT NULL AND precio%d > 0
        `, listaPrecios, listaPrecios)
        var total int
        if err := db.Local.QueryRow(countQuery, catID).Scan(&total); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "productos": productos,
            "total":     total,
        })
    }
}

// GetProductoByID obtiene un producto por su ID, activo, y precio sucursal
func GetProductoByID(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        params := mux.Vars(r)
        idStr := params["id"]
        id, err := strconv.Atoi(idStr)
        if err != nil {
            http.Error(w, "ID inválido", http.StatusBadRequest)
            return
        }
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        query := fmt.Sprintf(`
            SELECT p.idproducto, p.descripcion, p.precio%d, p.estatus, c.categoria
            FROM crm_productos p
            LEFT JOIN categorias c ON p.idcategoria = c.idcategoria
            WHERE p.idproducto = ? AND p.estatus = 'S' AND p.precio%d IS NOT NULL AND p.precio%d > 0
        `, listaPrecios, listaPrecios, listaPrecios)
        var p Producto
        err = db.Local.QueryRow(query, id).Scan(&p.IDProducto, &p.Descripcion, &p.Precio, &p.Estatus, &p.Categoria)
        if err == sql.ErrNoRows {
            http.Error(w, "Producto no encontrado", http.StatusNotFound)
            return
        } else if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(p)
    }
}

// GetCategorias obtiene todas las categorías con estatus = 'S', ordenadas alfabéticamente y con productos disponibles según sucursal
func GetCategorias(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        query := fmt.Sprintf(`
            SELECT c.idcategoria, c.categoria, c.estatus
            FROM categorias c
            JOIN crm_productos p ON p.idcategoria = c.idcategoria 
                AND p.estatus = 'S'
                AND p.precio%d IS NOT NULL
                AND p.precio%d > 0
            WHERE c.estatus = 'S'
            GROUP BY c.idcategoria, c.categoria, c.estatus
            HAVING COUNT(p.idproducto) > 0
            ORDER BY c.categoria ASC
        `, listaPrecios, listaPrecios)
        rows, err := db.Local.Query(query)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        var categorias []Categoria
        for rows.Next() {
            var c Categoria
            if err := rows.Scan(&c.IDCategoria, &c.Categoria, &c.Estatus); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            categorias = append(categorias, c)
        }
        json.NewEncoder(w).Encode(categorias)
    }
}

// GetProductosConIVA obtiene productos paginados con estatus = 'S', su IVA, tipo_iva y precio final calculado según sucursal
func GetProductosConIVA(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        limit, offset := getPagination(r)
        countQuery := fmt.Sprintf(`
            SELECT COUNT(*) FROM crm_productos WHERE estatus = 'S' AND precio%d IS NOT NULL AND precio%d > 0
        `, listaPrecios, listaPrecios)
        var total int
        if err := db.Local.QueryRow(countQuery).Scan(&total); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        query := fmt.Sprintf(`
            SELECT p.idproducto, p.descripcion, p.precio%d, p.estatus, COALESCE(c.categoria, 'Sin categoría') as categoria,
                   p.idiva, IFNULL(i.iva, 0), IFNULL(i.tipo_iva, '')
            FROM crm_productos p
            LEFT JOIN categorias c ON p.idcategoria = c.idcategoria
            LEFT JOIN crm_impuestos i ON p.idiva = i.idiva
            WHERE p.estatus = 'S' AND p.precio%d IS NOT NULL AND p.precio%d > 0
            LIMIT ? OFFSET ?
        `, listaPrecios, listaPrecios, listaPrecios)
        rows, err := db.Local.Query(query, limit, offset)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        var productos []ProductoConIVA
        for rows.Next() {
            var p ProductoConIVA
            if err := rows.Scan(
                &p.IDProducto,
                &p.Descripcion,
                &p.Precio,
                &p.Estatus,
                &p.Categoria,
                &p.IDIVA,
                &p.IVA,
                &p.TipoIVA,
            ); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            base := p.Precio.Float64
            iva := p.IVA.Float64
            if !p.IDIVA.Valid || !p.IVA.Valid || p.IDIVA.Int64 == 0 {
                p.PrecioFinal = base
            } else {
                p.PrecioFinal = base + (base * iva / 100)
            }
            productos = append(productos, p)
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "productos": productos,
            "total":     total,
        })
    }
}

// GetProductosConIVAPorCategoria obtiene productos con IVA y precio final filtrando por categoría y paginados y precio sucursal
func GetProductosConIVAPorCategoria(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        params := mux.Vars(r)
        catIDStr := params["idcategoria"]
        catID, err := strconv.Atoi(catIDStr)
        if err != nil {
            http.Error(w, "ID de categoría inválido", http.StatusBadRequest)
            return
        }
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        limit, offset := getPagination(r)
        var total int
        var rows *sql.Rows
        var query string
        if catID == 0 {
            countQuery := fmt.Sprintf(`
                SELECT COUNT(*) FROM crm_productos WHERE estatus = 'S' AND precio%d IS NOT NULL AND precio%d > 0
            `, listaPrecios, listaPrecios)
            if err := db.Local.QueryRow(countQuery).Scan(&total); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            query = fmt.Sprintf(`
                SELECT p.idproducto, p.descripcion, p.precio%d, p.estatus, c.categoria, 
                    p.idiva, IFNULL(i.iva, 0), IFNULL(i.tipo_iva, '')
                FROM crm_productos p
                LEFT JOIN categorias c ON p.idcategoria = c.idcategoria
                LEFT JOIN crm_impuestos i ON p.idiva = i.idiva
                WHERE p.estatus = 'S' AND p.precio%d IS NOT NULL AND p.precio%d > 0
                LIMIT ? OFFSET ?
            `, listaPrecios, listaPrecios, listaPrecios)
            rows, err = db.Local.Query(query, limit, offset)
        } else {
            countQuery := fmt.Sprintf(`
                SELECT COUNT(*) FROM crm_productos WHERE idcategoria = ? AND estatus = 'S' AND precio%d IS NOT NULL AND precio%d > 0
            `, listaPrecios, listaPrecios)
            if err := db.Local.QueryRow(countQuery, catID).Scan(&total); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            query = fmt.Sprintf(`
                SELECT p.idproducto, p.descripcion, p.precio%d, p.estatus, c.categoria, 
                    p.idiva, IFNULL(i.iva, 0), IFNULL(i.tipo_iva, '')
                FROM crm_productos p
                LEFT JOIN categorias c ON p.idcategoria = c.idcategoria
                LEFT JOIN crm_impuestos i ON p.idiva = i.idiva
                WHERE p.idcategoria = ? AND p.estatus = 'S' AND p.precio%d IS NOT NULL AND p.precio%d > 0
                LIMIT ? OFFSET ?
            `, listaPrecios, listaPrecios, listaPrecios)
            rows, err = db.Local.Query(query, catID, limit, offset)
        }

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        var productos []ProductoConIVA
        for rows.Next() {
            var p ProductoConIVA
            if err := rows.Scan(
                &p.IDProducto,
                &p.Descripcion,
                &p.Precio,
                &p.Estatus,
                &p.Categoria,
                &p.IDIVA,
                &p.IVA,
                &p.TipoIVA,
            ); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            base := p.Precio.Float64
            iva := p.IVA.Float64
            if !p.IDIVA.Valid || !p.IVA.Valid || p.IDIVA.Int64 == 0 {
                p.PrecioFinal = base
            } else {
                p.PrecioFinal = base + (base * iva / 100)
            }
            productos = append(productos, p)
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "productos": productos,
            "total":     total,
        })
    }
}

// GetProductoImpuesto obtiene el producto por ID y calcula precio final según sucursal
func GetProductoImpuesto(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        params := mux.Vars(r)
        idStr := params["id"]
        id, err := strconv.Atoi(idStr)
        if err != nil {
            http.Error(w, "ID inválido", http.StatusBadRequest)
            return
        }
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        var prod ProductoImpuesto
        queryProd := fmt.Sprintf(`
            SELECT idproducto, descripcion, precio%d, idiva
            FROM crm_productos
            WHERE idproducto = ? AND estatus = 'S' AND precio%d IS NOT NULL AND precio%d > 0
        `, listaPrecios, listaPrecios, listaPrecios)
        err = db.Local.QueryRow(queryProd, id).Scan(
            &prod.IDProducto,
            &prod.Descripcion,
            &prod.PrecioBase,
            &prod.IDIVA,
        )
        if err == sql.ErrNoRows {
            http.Error(w, "Producto no encontrado", http.StatusNotFound)
            return
        } else if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        queryImp := `
            SELECT iva, tipo_iva
            FROM crm_impuestos
            WHERE idiva = ?
        `
        err = db.Local.QueryRow(queryImp, prod.IDIVA).Scan(&prod.IVA, &prod.TipoIVA)
        if err == sql.ErrNoRows {
            http.Error(w, "Impuesto no encontrado", http.StatusNotFound)
            return
        } else if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        precioBase := prod.PrecioBase.Float64
        iva := prod.IVA.Float64
        prod.PrecioFinal = precioBase + (precioBase * iva / 100)
        json.NewEncoder(w).Encode(prod)
    }
}

// GetProductosConIVABuscar busca productos por término o ID, con IVA, paginación y precio sucursal
func GetProductosConIVABuscar(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        q := r.URL.Query().Get("q")
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        limit, offset := getPagination(r)
        var total int
        var rows *sql.Rows
        var query string
        var args []interface{}

        if id, err := strconv.Atoi(q); err == nil {
            countQuery := fmt.Sprintf("SELECT COUNT(*) FROM crm_productos WHERE idproducto = ? AND estatus = 'S' AND precio%d IS NOT NULL AND precio%d > 0", listaPrecios, listaPrecios)
            db.Local.QueryRow(countQuery, id).Scan(&total)
            query = fmt.Sprintf(`
                SELECT p.idproducto, p.descripcion, p.precio%d, p.estatus, c.categoria,
                       p.idiva, IFNULL(i.iva, 0), IFNULL(i.tipo_iva, '')
                FROM crm_productos p
                LEFT JOIN categorias c ON p.idcategoria = c.idcategoria
                LEFT JOIN crm_impuestos i ON p.idiva = i.idiva
                WHERE p.idproducto = ? AND p.estatus = 'S' AND p.precio%d IS NOT NULL AND p.precio%d > 0
                LIMIT ? OFFSET ?
            `, listaPrecios, listaPrecios, listaPrecios)
            args = []interface{}{id, limit, offset}
        } else {
            like := "%" + q + "%"
            countQuery := fmt.Sprintf(
                `SELECT COUNT(*) FROM crm_productos WHERE estatus = 'S' AND (LOWER(descripcion) LIKE LOWER(?) OR LOWER(clave) LIKE LOWER(?)) AND precio%d IS NOT NULL AND precio%d > 0`,
                listaPrecios, listaPrecios)
            db.Local.QueryRow(countQuery, like, like).Scan(&total)
            query = fmt.Sprintf(`
                SELECT p.idproducto, p.descripcion, p.precio%d, p.estatus, c.categoria,
                       p.idiva, IFNULL(i.iva, 0), IFNULL(i.tipo_iva, '')
                FROM crm_productos p
                LEFT JOIN categorias c ON p.idcategoria = c.idcategoria
                LEFT JOIN crm_impuestos i ON p.idiva = i.idiva
                WHERE p.estatus = 'S' AND (LOWER(p.descripcion) LIKE LOWER(?) OR LOWER(p.clave) LIKE LOWER(?))
                AND p.precio%d IS NOT NULL AND p.precio%d > 0
                LIMIT ? OFFSET ?
            `, listaPrecios, listaPrecios, listaPrecios, listaPrecios)
            args = []interface{}{like, like, limit, offset}
        }

        rows, err = db.Local.Query(query, args...)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        var productos []ProductoConIVA
        for rows.Next() {
            var p ProductoConIVA
            if err := rows.Scan(
                &p.IDProducto,
                &p.Descripcion,
                &p.Precio,
                &p.Estatus,
                &p.Categoria,
                &p.IDIVA,
                &p.IVA,
                &p.TipoIVA,
            ); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            base := p.Precio.Float64
            iva := p.IVA.Float64
            if !p.IDIVA.Valid || p.IDIVA.Int64 == 0 || !p.IVA.Valid {
                p.PrecioFinal = base
            } else {
                p.PrecioFinal = base + (base * iva / 100)
            }
            productos = append(productos, p)
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "productos": productos,
            "total":     total,
        })
    }
}

// SearchProductos busca productos por término (clave o descripción) con paginación y precio sucursal
func SearchProductos(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        q := r.URL.Query().Get("q")
        if q == "" {
            http.Error(w, "Falta el parámetro de búsqueda 'q'", http.StatusBadRequest)
            return
        }
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        limit, offset := getPagination(r)
        like := "%" + q + "%"
        countQuery := fmt.Sprintf(`
            SELECT COUNT(*) FROM crm_productos 
            WHERE estatus = 'S' AND (descripcion LIKE ? OR clave LIKE ?) AND precio%d IS NOT NULL AND precio%d > 0
        `, listaPrecios, listaPrecios)
        var total int
        if err := db.Local.QueryRow(countQuery, like, like).Scan(&total); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        query := fmt.Sprintf(`
            SELECT p.idproducto, p.descripcion, p.precio%d, p.estatus, c.categoria
            FROM crm_productos p
            LEFT JOIN categorias c ON p.idcategoria = c.idcategoria
            WHERE p.estatus = 'S' AND (p.descripcion LIKE ? OR p.clave LIKE ?)
            AND p.precio%d IS NOT NULL AND p.precio%d > 0
            LIMIT ? OFFSET ?
        `, listaPrecios, listaPrecios, listaPrecios, listaPrecios)
        rows, err := db.Local.Query(query, like, like, limit, offset)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        var productos []Producto
        for rows.Next() {
            var p Producto
            if err := rows.Scan(&p.IDProducto, &p.Descripcion, &p.Precio, &p.Estatus, &p.Categoria); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            productos = append(productos, p)
        }
        json.NewEncoder(w).Encode(map[string]interface{}{
            "productos": productos,
            "total":     total,
        })
    }
}

// GetProductoSuggestions devuelve sugerencias para autocomplete (nombre/descripcion), mínimo 3 caracteres y precio sucursal
func GetProductoSuggestions(db *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        q := r.URL.Query().Get("q")
        q = strings.TrimSpace(q)
        if len([]rune(q)) < 3 {
            http.Error(w, "Debe ingresar al menos 3 caracteres", http.StatusBadRequest)
            return
        }
        idUsuarioStr := r.URL.Query().Get("id_usuario")
        idUsuario, err := strconv.Atoi(idUsuarioStr)
        if err != nil {
            http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
            return
        }
        listaPrecios, err := getListaPreciosPorUsuario(db, idUsuario)
        if err != nil {
            http.Error(w, "No se pudo obtener lista de precios", http.StatusInternalServerError)
            return
        }
        query := fmt.Sprintf(`
            SELECT idproducto, descripcion as nombre, descripcion
            FROM crm_productos
            WHERE estatus = 'S'
            AND (LOWER(descripcion) LIKE LOWER(?) OR LOWER(clave) LIKE LOWER(?))
            AND precio%d IS NOT NULL AND precio%d > 0
            ORDER BY descripcion ASC
            LIMIT 10
        `, listaPrecios, listaPrecios)
        likeQuery := "%" + q + "%"
        rows, err := db.Local.Query(query, likeQuery, likeQuery)
        if err != nil {
            http.Error(w, "Error al buscar sugerencias", http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        var resultados []ProductoSugerencia
        for rows.Next() {
            var p ProductoSugerencia
            if err := rows.Scan(&p.IDProducto, &p.Nombre, &p.Descripcion); err == nil {
                resultados = append(resultados, p)
            }
        }
        json.NewEncoder(w).Encode(resultados)
    }
}
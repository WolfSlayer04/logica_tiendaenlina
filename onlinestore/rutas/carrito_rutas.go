package rutas

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "sync"

    "onlinestore/db"

    "github.com/gorilla/mux"
)

// CartItem representa un producto en el carrito
type CartItem struct {
    IDProducto  int     `json:"idproducto"`
    Cantidad    int     `json:"cantidad"`
    Precio      float64 `json:"precio"`
    Descripcion string  `json:"descripcion,omitempty"`
}

// Carrito es una estructura temporal para almacenar productos en memoria
var (
    cartItems = make(map[int]CartItem)
    mu        sync.RWMutex
)

// AddToCart agrega o actualiza un producto en el carrito (suma cantidad)
func AddToCart(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            IDProducto int     `json:"idproducto"`
            Cantidad   int     `json:"cantidad"`
            Precio     float64 `json:"precio"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Datos inválidos", http.StatusBadRequest)
            return
        }

        // Validación básica
        if req.IDProducto <= 0 || req.Cantidad <= 0 || req.Precio < 0 {
            http.Error(w, "Datos incompletos", http.StatusBadRequest)
            return
        }

        // Verificar que el producto exista y esté activo
        var exists bool
        err := dbc.Local.QueryRow("SELECT EXISTS(SELECT 1 FROM crm_productos WHERE idproducto = ? AND estatus = 'S')", req.IDProducto).Scan(&exists)
        if err != nil {
            http.Error(w, "Error al verificar el producto", http.StatusInternalServerError)
            return
        }
        if !exists {
            http.Error(w, "Producto no encontrado o inactivo", http.StatusNotFound)
            return
        }

        // Agregar o actualizar en el carrito
        mu.Lock()
        defer mu.Unlock()

        item, exists := cartItems[req.IDProducto]
        if exists {
            item.Cantidad += req.Cantidad
        } else {
            item = CartItem{
                IDProducto: req.IDProducto,
                Cantidad:   req.Cantidad,
                Precio:     req.Precio,
            }
        }
        cartItems[req.IDProducto] = item

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Producto agregado al carrito"})
    }
}

// UpdateCartItem actualiza la cantidad de un producto en el carrito (puede bajar o subir cantidad, si es 0 elimina)
func UpdateCartItem(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            IDProducto int `json:"idproducto"`
            Cantidad   int `json:"cantidad"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Datos inválidos", http.StatusBadRequest)
            return
        }

        if req.IDProducto <= 0 || req.Cantidad < 0 {
            http.Error(w, "Datos incompletos", http.StatusBadRequest)
            return
        }

        mu.Lock()
        defer mu.Unlock()

        item, exists := cartItems[req.IDProducto]
        if !exists {
            http.Error(w, "Producto no está en el carrito", http.StatusNotFound)
            return
        }

        if req.Cantidad == 0 {
            delete(cartItems, req.IDProducto)
        } else {
            item.Cantidad = req.Cantidad
            cartItems[req.IDProducto] = item
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"message": "Cantidad actualizada"})
    }
}

// RemoveFromCart elimina un producto del carrito completamente
func RemoveFromCart(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        idStr, ok := vars["idproducto"]
        if !ok {
            http.Error(w, "ID de producto faltante", http.StatusBadRequest)
            return
        }
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
            http.Error(w, "ID de producto inválido", http.StatusBadRequest)
            return
        }

        mu.Lock()
        defer mu.Unlock()
        _, exists := cartItems[id]
        if !exists {
            http.Error(w, "Producto no está en el carrito", http.StatusNotFound)
            return
        }
        delete(cartItems, id)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Producto eliminado del carrito"})
    }
}

// GetCart devuelve todos los productos en el carrito
func GetCart(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        mu.RLock()
        defer mu.RUnlock()

        // Obtener IDs de productos en el carrito
        var ids []int
        for id := range cartItems {
            ids = append(ids, id)
        }

        if len(ids) == 0 {
            // Si no hay productos en el carrito, responde con un slice vacío de CartItem
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode([]CartItem{})
            return
        }

        // Construir consulta dinámica
        placeholders := ""
        args := make([]interface{}, len(ids))
        for i, id := range ids {
            if i > 0 {
                placeholders += ", "
            }
            placeholders += "?"
            args[i] = id
        }

        query := "SELECT idproducto, descripcion FROM crm_productos WHERE idproducto IN (" + placeholders + ")"

        rows, err := dbc.Local.Query(query, args...)
        if err != nil {
            http.Error(w, "Error al cargar descripciones", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        descripciones := make(map[int]string)
        for rows.Next() {
            var id int
            var desc sql.NullString
            if err := rows.Scan(&id, &desc); err == nil && desc.Valid {
                descripciones[id] = desc.String
            }
        }

        // Enviar respuesta con descripciones
        var list []CartItem
        for _, item := range cartItems {
            item.Descripcion = descripciones[item.IDProducto]
            list = append(list, item)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(list)
    }
}

// ClearCart vacía el carrito
func ClearCart(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        mu.Lock()
        defer mu.Unlock()

        cartItems = make(map[int]CartItem)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Carrito vaciado"})
    }
}
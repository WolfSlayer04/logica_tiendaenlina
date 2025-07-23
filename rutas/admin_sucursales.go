package rutas

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/WolfSlayer04/logica_tiendaenlina/db"

	"github.com/gorilla/mux"
)

// Estructura para sucursal
type Sucursal struct {
	IDSucursal   int     `json:"idsucursal"`
	IDEmpresa    int     `json:"idempresa"`
	Sucursal     string  `json:"sucursal"`
	Direccion    string  `json:"direccion"`
	Ciudad       string  `json:"ciudad"`
	Colonia      string  `json:"colonia"`
	CP           string  `json:"cp"`
	Estatus      string  `json:"estatus"`
	TipoObjeto   string  `json:"tipo_objeto"`
	Radio        float64 `json:"radio"`
	ListaPrecios int     `json:"lista_precios"`
}

func GetSucursalALL(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `SELECT idsucursal, idempresa, sucursal, direccion, ciudad, colonia, cp, estatus, tipo_objeto, radio, lista_precios
				  FROM adm_sucursales`
		rows, err := dbConn.Local.Query(query)
		if err != nil {
			http.Error(w, "Error de base de datos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var sucursales []Sucursal
		for rows.Next() {
			var s Sucursal
			err := rows.Scan(
				&s.IDSucursal, &s.IDEmpresa, &s.Sucursal, &s.Direccion, &s.Ciudad,
				&s.Colonia, &s.CP, &s.Estatus, &s.TipoObjeto, &s.Radio, &s.ListaPrecios,
			)
			if err != nil {
				http.Error(w, "Error leyendo sucursales", http.StatusInternalServerError)
				return
			}
			sucursales = append(sucursales, s)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error finalizando la consulta", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sucursales)
	}
}

// GET /api/sucursales/{id}: Obtiene la info completa de la sucursal
func GetSucursal(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "ID de sucursal inválido", http.StatusBadRequest)
			return
		}
		var s Sucursal
		query := `SELECT idsucursal, idempresa, sucursal, direccion, ciudad, colonia, cp, estatus, tipo_objeto, radio, lista_precios
				  FROM adm_sucursales WHERE idsucursal = ?`
		err = dbConn.Local.QueryRow(query, id).Scan(
			&s.IDSucursal, &s.IDEmpresa, &s.Sucursal, &s.Direccion, &s.Ciudad,
			&s.Colonia, &s.CP, &s.Estatus, &s.TipoObjeto, &s.Radio, &s.ListaPrecios,
		)
		if err == sql.ErrNoRows {
			http.Error(w, "Sucursal no encontrada", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Error de base de datos", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	}
}

// PUT /api/sucursales/{id}/lista-precios: Cambia la lista de precios para la sucursal
func UpdateListaPreciosSucursal(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "ID de sucursal inválido", http.StatusBadRequest)
			return
		}
		var body struct {
			ListaPrecios int `json:"lista_precios"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}
		if body.ListaPrecios < 1 || body.ListaPrecios > 25 {
			http.Error(w, "La lista de precios debe estar entre 1 y 25", http.StatusBadRequest)
			return
		}
		res, err := dbConn.Local.Exec("UPDATE adm_sucursales SET lista_precios = ? WHERE idsucursal = ?", body.ListaPrecios, id)
		if err != nil {
			http.Error(w, "Error actualizando lista de precios", http.StatusInternalServerError)
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			http.Error(w, "Sucursal no encontrada", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Lista de precios actualizada"}`))
	}
}

// GET /api/sucursales/{id}/productos: Lista productos con precio de la lista seleccionada en la sucursal
func GetProductosSucursal(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "ID de sucursal inválido", http.StatusBadRequest)
			return
		}
		// Obtener lista_precios
		var listaPrecios int
		err = dbConn.Local.QueryRow("SELECT lista_precios FROM adm_sucursales WHERE idsucursal = ?", id).Scan(&listaPrecios)
		if err == sql.ErrNoRows {
			http.Error(w, "Sucursal no encontrada", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Error de base de datos", http.StatusInternalServerError)
			return
		}
		if listaPrecios < 1 || listaPrecios > 25 {
			listaPrecios = 1 // fallback seguro
		}
		// precioField no es necesario, no lo usamos porque accedemos todos los precios y seleccionamos en Go
		query := `
			SELECT idproducto, idempresa, idlinea, descripcion, estatus, tipo_prod, idcategoria, clasif, con_formula, clave, idiva, cod_barras,
			    precio1, precio2, precio3, precio4, precio5, precio6, precio7, precio8, precio9, precio10,
				precio11, precio12, precio13, precio14, precio15, precio16, precio17, precio18, precio19, precio20,
				precio21, precio22, precio23, precio24, precio25,
				ieps_adic, con_ieps_adic, unidad, unidad_ent, factor_conversion, sat_clave, sat_medida, volumen, peso,
				idmoneda, lote, desc_ticket, cant_sig_lista, en_venta
			FROM crm_productos
			WHERE estatus = 'S'
		`
		rows, err := dbConn.Local.Query(query)
		if err != nil {
			http.Error(w, "Error obteniendo productos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		productos := []map[string]interface{}{}
		for rows.Next() {
			var p ProductoDB
			err := rows.Scan(
				&p.IDProducto, &p.IDEmpresa, &p.IDLinea, &p.Descripcion, &p.Estatus, &p.TipoProd, &p.IDCategoria, &p.Clasif, &p.ConFormula,
				&p.Clave, &p.IDIVA, &p.CodBarras, &p.Precio1, &p.Precio2, &p.Precio3, &p.Precio4, &p.Precio5, &p.Precio6, &p.Precio7,
				&p.Precio8, &p.Precio9, &p.Precio10, &p.Precio11, &p.Precio12, &p.Precio13, &p.Precio14, &p.Precio15, &p.Precio16,
				&p.Precio17, &p.Precio18, &p.Precio19, &p.Precio20, &p.Precio21, &p.Precio22, &p.Precio23, &p.Precio24, &p.Precio25,
				&p.IEPSAdic, &p.ConIEPSAdic, &p.Unidad, &p.UnidadEnt, &p.FactorConversion, &p.SATClave, &p.SATMedida, &p.Volumen,
				&p.Peso, &p.IDMoneda, &p.Lote, &p.DescTicket, &p.CantSigLista, &p.EnVenta,
			)
			if err != nil {
				http.Error(w, "Error leyendo productos", http.StatusInternalServerError)
				return
			}

			// Obtener el precio correspondiente
			var precio sql.NullFloat64
			switch listaPrecios {
			case 1:
				precio = p.Precio1
			case 2:
				precio = p.Precio2
			case 3:
				precio = p.Precio3
			case 4:
				precio = p.Precio4
			case 5:
				precio = p.Precio5
			case 6:
				precio = p.Precio6
			case 7:
				precio = p.Precio7
			case 8:
				precio = p.Precio8
			case 9:
				precio = p.Precio9
			case 10:
				precio = p.Precio10
			case 11:
				precio = p.Precio11
			case 12:
				precio = p.Precio12
			case 13:
				precio = p.Precio13
			case 14:
				precio = p.Precio14
			case 15:
				precio = p.Precio15
			case 16:
				precio = p.Precio16
			case 17:
				precio = p.Precio17
			case 18:
				precio = p.Precio18
			case 19:
				precio = p.Precio19
			case 20:
				precio = p.Precio20
			case 21:
				precio = p.Precio21
			case 22:
				precio = p.Precio22
			case 23:
				precio = p.Precio23
			case 24:
				precio = p.Precio24
			case 25:
				precio = p.Precio25
			default:
				precio = p.Precio1
			}

			resp := map[string]interface{}{
				"idproducto":   p.IDProducto,
				"idempresa":    p.IDEmpresa,
				"idlinea":      p.IDLinea.Int64,
				"descripcion":  p.Descripcion,
				"estatus":      p.Estatus,
				"tipo_prod":    p.TipoProd.String,
				"idcategoria":  p.IDCategoria.Int64,
				"clasif":       p.Clasif.String,
				"con_formula":  p.ConFormula.String,
				"clave":        p.Clave.String,
				"idiva":        p.IDIVA.Int64,
				"cod_barras":   p.CodBarras.String,
				"precio":       precio.Float64,
				"unidad":       p.Unidad.String,
				"unidad_ent":   p.UnidadEnt.String,
				"factor_conversion": p.FactorConversion.Float64,
				"sat_clave":    p.SATClave.String,
				"sat_medida":   p.SATMedida.String,
				"volumen":      p.Volumen.Float64,
				"peso":         p.Peso.Float64,
				"idmoneda":     p.IDMoneda.Int64,
				"lote":         p.Lote.String,
				"desc_ticket":  p.DescTicket.String,
				"cant_sig_lista": p.CantSigLista.Int64,
				"en_venta":     p.EnVenta.String,
			}

			productos = append(productos, resp)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(productos)
	}
}

// GET /api/sucursales/{id}/lista-precios
func GetListaPreciosSucursal(dbConn *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.Atoi(vars["id"])
        if err != nil {
            http.Error(w, "ID de sucursal inválido", http.StatusBadRequest)
            return
        }
        var listaPrecios int
        err = dbConn.Local.QueryRow("SELECT lista_precios FROM adm_sucursales WHERE idsucursal = ?", id).Scan(&listaPrecios)
        if err == sql.ErrNoRows {
            http.Error(w, "Sucursal no encontrada", http.StatusNotFound)
            return
        } else if err != nil {
            http.Error(w, "Error de base de datos", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]int{"lista_precios": listaPrecios})
    }
}
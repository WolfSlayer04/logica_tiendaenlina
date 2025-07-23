package rutas

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"fmt"
	"strings"

	"github.com/WolfSlayer04/logica_tiendaenlina/db"
	"github.com/gorilla/mux"
)

// ProductoInput define los campos aceptados para crear o editar producto
type ProductoInput struct {
	IDEmpresa        int      `json:"idempresa"`
	IDLinea          *int64   `json:"idlinea,omitempty"`
	Descripcion      string   `json:"descripcion"`
	Estatus          string   `json:"estatus"`
	TipoProd         *string  `json:"tipo_prod,omitempty"`
	IDCategoria      *int64   `json:"idcategoria,omitempty"`
	Clasif           *string  `json:"clasif,omitempty"`
	ConFormula       *string  `json:"con_formula,omitempty"`
	Clave            *string  `json:"clave,omitempty"`
	IDIVA            *int64   `json:"idiva,omitempty"`
	CodBarras        *string  `json:"cod_barras,omitempty"`
	Precio1          *float64 `json:"precio1,omitempty"`
	Precio2          *float64 `json:"precio2,omitempty"`
	Precio3          *float64 `json:"precio3,omitempty"`
	Precio4          *float64 `json:"precio4,omitempty"`
	Precio5          *float64 `json:"precio5,omitempty"`
	Precio6          *float64 `json:"precio6,omitempty"`
	Precio7          *float64 `json:"precio7,omitempty"`
	Precio8          *float64 `json:"precio8,omitempty"`
	Precio9          *float64 `json:"precio9,omitempty"`
	Precio10         *float64 `json:"precio10,omitempty"`
	Precio11         *float64 `json:"precio11,omitempty"`
	Precio12         *float64 `json:"precio12,omitempty"`
	Precio13         *float64 `json:"precio13,omitempty"`
	Precio14         *float64 `json:"precio14,omitempty"`
	Precio15         *float64 `json:"precio15,omitempty"`
	Precio16         *float64 `json:"precio16,omitempty"`
	Precio17         *float64 `json:"precio17,omitempty"`
	Precio18         *float64 `json:"precio18,omitempty"`
	Precio19         *float64 `json:"precio19,omitempty"`
	Precio20         *float64 `json:"precio20,omitempty"`
	Precio21         *float64 `json:"precio21,omitempty"`
	Precio22         *float64 `json:"precio22,omitempty"`
	Precio23         *float64 `json:"precio23,omitempty"`
	Precio24         *float64 `json:"precio24,omitempty"`
	Precio25         *float64 `json:"precio25,omitempty"`
	IEPSAdic         *float64 `json:"ieps_adic,omitempty"`
	ConIEPSAdic      *string  `json:"con_ieps_adic,omitempty"`
	Unidad           *string  `json:"unidad,omitempty"`
	UnidadEnt        *string  `json:"unidad_ent,omitempty"`
	FactorConversion *float64 `json:"factor_conversion,omitempty"`
	SATClave         *string  `json:"sat_clave,omitempty"`
	SATMedida        *string  `json:"sat_medida,omitempty"`
	Volumen          *float64 `json:"volumen,omitempty"`
	Peso             *float64 `json:"peso,omitempty"`
	IDMoneda         *int64   `json:"idmoneda,omitempty"`
	Lote             *string  `json:"lote,omitempty"`
	DescTicket       *string  `json:"desc_ticket,omitempty"`
	CantSigLista     *int64   `json:"cant_sig_lista,omitempty"`
	EnVenta          *string  `json:"en_venta,omitempty"`
}

// ProductoDB define la estructura completa de un producto para DB scan y respuestas administrativas
type ProductoDB struct {
	IDProducto        int             `json:"idproducto"`
	IDEmpresa         int             `json:"idempresa"`
	IDLinea           sql.NullInt64   `json:"idlinea"`
	Descripcion       string          `json:"descripcion"`
	Estatus           string          `json:"estatus"`
	TipoProd          sql.NullString  `json:"tipo_prod"`
	IDCategoria       sql.NullInt64   `json:"idcategoria"`
	Clasif            sql.NullString  `json:"clasif"`
	ConFormula        sql.NullString  `json:"con_formula"`
	Clave             sql.NullString  `json:"clave"`
	IDIVA             sql.NullInt64   `json:"idiva"`
	CodBarras         sql.NullString  `json:"cod_barras"`
	Precio1           sql.NullFloat64 `json:"precio1"`
	Precio2           sql.NullFloat64 `json:"precio2"`
	Precio3           sql.NullFloat64 `json:"precio3"`
	Precio4           sql.NullFloat64 `json:"precio4"`
	Precio5           sql.NullFloat64 `json:"precio5"`
	Precio6           sql.NullFloat64 `json:"precio6"`
	Precio7           sql.NullFloat64 `json:"precio7"`
	Precio8           sql.NullFloat64 `json:"precio8"`
	Precio9           sql.NullFloat64 `json:"precio9"`
	Precio10          sql.NullFloat64 `json:"precio10"`
	Precio11          sql.NullFloat64 `json:"precio11"`
	Precio12          sql.NullFloat64 `json:"precio12"`
	Precio13          sql.NullFloat64 `json:"precio13"`
	Precio14          sql.NullFloat64 `json:"precio14"`
	Precio15          sql.NullFloat64 `json:"precio15"`
	Precio16          sql.NullFloat64 `json:"precio16"`
	Precio17          sql.NullFloat64 `json:"precio17"`
	Precio18          sql.NullFloat64 `json:"precio18"`
	Precio19          sql.NullFloat64 `json:"precio19"`
	Precio20          sql.NullFloat64 `json:"precio20"`
	Precio21          sql.NullFloat64 `json:"precio21"`
	Precio22          sql.NullFloat64 `json:"precio22"`
	Precio23          sql.NullFloat64 `json:"precio23"`
	Precio24          sql.NullFloat64 `json:"precio24"`
	Precio25          sql.NullFloat64 `json:"precio25"`
	IEPSAdic          sql.NullFloat64 `json:"ieps_adic"`
	ConIEPSAdic       sql.NullString  `json:"con_ieps_adic"`
	Unidad            sql.NullString  `json:"unidad"`
	UnidadEnt         sql.NullString  `json:"unidad_ent"`
	FactorConversion  sql.NullFloat64 `json:"factor_conversion"`
	SATClave          sql.NullString  `json:"sat_clave"`
	SATMedida         sql.NullString  `json:"sat_medida"`
	Volumen           sql.NullFloat64 `json:"volumen"`
	Peso              sql.NullFloat64 `json:"peso"`
	IDMoneda          sql.NullInt64   `json:"idmoneda"`
	Lote              sql.NullString  `json:"lote"`
	DescTicket        sql.NullString  `json:"desc_ticket"`
	CantSigLista      sql.NullInt64   `json:"cant_sig_lista"`
	EnVenta           sql.NullString  `json:"en_venta"`
}

// ProductoDBConIVA extiende ProductoDB con información de IVA, precio final y nombre de la empresa
type ProductoDBConIVA struct {
	ProductoDB
	NombreEmpresa sql.NullString  `json:"nombre_empresa"`
	IVA           sql.NullFloat64 `json:"iva"`
	TipoIVA       sql.NullString  `json:"tipo_iva"`
	PrecioFinal   float64         `json:"precio_final"`
}

// AddProducto mejorado: solo incluye los campos que realmente vienen en el input (no nulos)
func AddProducto(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		

		var input ProductoInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		

		if input.IDEmpresa == 0 || input.Descripcion == "" || input.Estatus == "" {
			
			http.Error(w, "Faltan campos obligatorios", http.StatusBadRequest)
			return
		}

		columns := []string{"idempresa", "descripcion", "estatus"}
		args := []interface{}{input.IDEmpresa, input.Descripcion, input.Estatus}

		// Solo agregamos los campos que realmente vienen
		if input.IDLinea != nil         { columns = append(columns, "idlinea");         args = append(args, sqlNullInt64(input.IDLinea)) }
		if input.TipoProd != nil        { columns = append(columns, "tipo_prod");       args = append(args, sqlNullString(input.TipoProd)) }
		if input.IDCategoria != nil     { columns = append(columns, "idcategoria");     args = append(args, sqlNullInt64(input.IDCategoria)) }
		if input.Clasif != nil          { columns = append(columns, "clasif");          args = append(args, sqlNullString(input.Clasif)) }
		if input.ConFormula != nil      { columns = append(columns, "con_formula");     args = append(args, sqlNullString(input.ConFormula)) }
		if input.Clave != nil           { columns = append(columns, "clave");           args = append(args, sqlNullString(input.Clave)) }
		if input.IDIVA != nil           { columns = append(columns, "idiva");           args = append(args, sqlNullInt64(input.IDIVA)) }
		if input.CodBarras != nil       { columns = append(columns, "cod_barras");      args = append(args, sqlNullString(input.CodBarras)) }
		if input.Precio1 != nil         { columns = append(columns, "precio1");         args = append(args, sqlNullFloat64(input.Precio1)) }
		if input.Precio2 != nil         { columns = append(columns, "precio2");         args = append(args, sqlNullFloat64(input.Precio2)) }
		if input.Precio3 != nil         { columns = append(columns, "precio3");         args = append(args, sqlNullFloat64(input.Precio3)) }
		if input.Precio4 != nil         { columns = append(columns, "precio4");         args = append(args, sqlNullFloat64(input.Precio4)) }
		if input.Precio5 != nil         { columns = append(columns, "precio5");         args = append(args, sqlNullFloat64(input.Precio5)) }
		if input.Precio6 != nil         { columns = append(columns, "precio6");         args = append(args, sqlNullFloat64(input.Precio6)) }
		if input.Precio7 != nil         { columns = append(columns, "precio7");         args = append(args, sqlNullFloat64(input.Precio7)) }
		if input.Precio8 != nil         { columns = append(columns, "precio8");         args = append(args, sqlNullFloat64(input.Precio8)) }
		if input.Precio9 != nil         { columns = append(columns, "precio9");         args = append(args, sqlNullFloat64(input.Precio9)) }
		if input.Precio10 != nil        { columns = append(columns, "precio10");        args = append(args, sqlNullFloat64(input.Precio10)) }
		if input.Precio11 != nil        { columns = append(columns, "precio11");        args = append(args, sqlNullFloat64(input.Precio11)) }
		if input.Precio12 != nil        { columns = append(columns, "precio12");        args = append(args, sqlNullFloat64(input.Precio12)) }
		if input.Precio13 != nil        { columns = append(columns, "precio13");        args = append(args, sqlNullFloat64(input.Precio13)) }
		if input.Precio14 != nil        { columns = append(columns, "precio14");        args = append(args, sqlNullFloat64(input.Precio14)) }
		if input.Precio15 != nil        { columns = append(columns, "precio15");        args = append(args, sqlNullFloat64(input.Precio15)) }
		if input.Precio16 != nil        { columns = append(columns, "precio16");        args = append(args, sqlNullFloat64(input.Precio16)) }
		if input.Precio17 != nil        { columns = append(columns, "precio17");        args = append(args, sqlNullFloat64(input.Precio17)) }
		if input.Precio18 != nil        { columns = append(columns, "precio18");        args = append(args, sqlNullFloat64(input.Precio18)) }
		if input.Precio19 != nil        { columns = append(columns, "precio19");        args = append(args, sqlNullFloat64(input.Precio19)) }
		if input.Precio20 != nil        { columns = append(columns, "precio20");        args = append(args, sqlNullFloat64(input.Precio20)) }
		if input.Precio21 != nil        { columns = append(columns, "precio21");        args = append(args, sqlNullFloat64(input.Precio21)) }
		if input.Precio22 != nil        { columns = append(columns, "precio22");        args = append(args, sqlNullFloat64(input.Precio22)) }
		if input.Precio23 != nil        { columns = append(columns, "precio23");        args = append(args, sqlNullFloat64(input.Precio23)) }
		if input.Precio24 != nil        { columns = append(columns, "precio24");        args = append(args, sqlNullFloat64(input.Precio24)) }
		if input.Precio25 != nil        { columns = append(columns, "precio25");        args = append(args, sqlNullFloat64(input.Precio25)) }
		if input.IEPSAdic != nil        { columns = append(columns, "ieps_adic");       args = append(args, sqlNullFloat64(input.IEPSAdic)) }
		if input.ConIEPSAdic != nil     { columns = append(columns, "con_ieps_adic");   args = append(args, sqlNullString(input.ConIEPSAdic)) }
		if input.Unidad != nil          { columns = append(columns, "unidad");          args = append(args, sqlNullString(input.Unidad)) }
		if input.UnidadEnt != nil       { columns = append(columns, "unidad_ent");      args = append(args, sqlNullString(input.UnidadEnt)) }
		if input.FactorConversion != nil{ columns = append(columns, "factor_conversion");args = append(args, sqlNullFloat64(input.FactorConversion)) }
		if input.SATClave != nil        { columns = append(columns, "sat_clave");       args = append(args, sqlNullString(input.SATClave)) }
		if input.SATMedida != nil       { columns = append(columns, "sat_medida");      args = append(args, sqlNullString(input.SATMedida)) }
		if input.Volumen != nil         { columns = append(columns, "volumen");         args = append(args, sqlNullFloat64(input.Volumen)) }
		if input.Peso != nil            { columns = append(columns, "peso");            args = append(args, sqlNullFloat64(input.Peso)) }
		if input.IDMoneda != nil        { columns = append(columns, "idmoneda");        args = append(args, sqlNullInt64(input.IDMoneda)) }
		if input.Lote != nil            { columns = append(columns, "lote");            args = append(args, sqlNullString(input.Lote)) }
		if input.DescTicket != nil      { columns = append(columns, "desc_ticket");     args = append(args, sqlNullString(input.DescTicket)) }
		if input.CantSigLista != nil    { columns = append(columns, "cant_sig_lista");  args = append(args, sqlNullInt64(input.CantSigLista)) }
		if input.EnVenta != nil         { columns = append(columns, "en_venta");        args = append(args, sqlNullString(input.EnVenta)) }

		query := fmt.Sprintf(
			"INSERT INTO crm_productos (%s) VALUES (%s)",
			strings.Join(columns, ", "),
			strings.TrimRight(strings.Repeat("?, ", len(args)), ", "),
		)
		res, err := dbConn.Local.Exec(query, args...)
		if err != nil {
			
			http.Error(w, "Error al insertar producto: "+err.Error(), http.StatusInternalServerError)
			return
		}

		insertedID, _ := res.LastInsertId()
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"idproducto": insertedID,
			"msg":        "Producto agregado correctamente",
		})
	}
}

// EditProducto mejorado: solo actualiza los campos recibidos en el input, el resto se conserva
func EditProducto(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		idStr := params["idproducto"]
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			http.Error(w, "ID de producto inválido", http.StatusBadRequest)
			return
		}

		// Obtener datos actuales del producto
		var current struct {
			Clave     sql.NullString
			IDEmpresa int
		}
		row := dbConn.Local.QueryRow("SELECT clave, idempresa FROM crm_productos WHERE idproducto=?", id)
		if err := row.Scan(&current.Clave, &current.IDEmpresa); err != nil {
			http.Error(w, "No se encontró el producto", http.StatusNotFound)
			return
		}

		var input ProductoInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		// Validar idiva si viene en input
		var idivaToSet sql.NullInt64
		if input.IDIVA != nil {
			var count int
			err := dbConn.Local.QueryRow("SELECT COUNT(*) FROM crm_impuestos WHERE idiva=? AND idempresa=?", *input.IDIVA, current.IDEmpresa).Scan(&count)
			if err != nil || count == 0 {
				http.Error(w, "idiva no válido para esta empresa", http.StatusBadRequest)
				return
			}
			idivaToSet = sqlNullInt64(input.IDIVA)
		} else {
			row := dbConn.Local.QueryRow("SELECT idiva FROM crm_productos WHERE idproducto=?", id)
			var idivaActual sql.NullInt64
			row.Scan(&idivaActual)
			idivaToSet = idivaActual
		}

		// Forzar reglas (siempre actualiza estos campos)
		idlineaToSet := sql.NullInt64{Int64: 0, Valid: true}
		tipoProdToSet := sql.NullString{String: "P", Valid: true}
		clasifToSet := sql.NullString{String: "", Valid: true}
		conFormulaToSet := sql.NullString{String: "N", Valid: true}
		claveToSet := current.Clave

		updateFields := []string{"idlinea = ?", "tipo_prod = ?", "clasif = ?", "con_formula = ?", "idiva = ?"}
		args := []interface{}{idlineaToSet, tipoProdToSet, clasifToSet, conFormulaToSet, idivaToSet}

		// Solo actualiza los campos recibidos (excepto clave)
		if input.IDEmpresa != 0            { updateFields = append(updateFields, "idempresa = ?");        args = append(args, input.IDEmpresa) }
		if input.Descripcion != ""         { updateFields = append(updateFields, "descripcion = ?");      args = append(args, input.Descripcion) }
		if input.Estatus != ""             { updateFields = append(updateFields, "estatus = ?");          args = append(args, input.Estatus) }
		if input.IDCategoria != nil        { updateFields = append(updateFields, "idcategoria = ?");      args = append(args, sqlNullInt64(input.IDCategoria)) }
		if input.ConIEPSAdic != nil        { updateFields = append(updateFields, "con_ieps_adic = ?");    args = append(args, sqlNullString(input.ConIEPSAdic)) }
		if input.CodBarras != nil          { updateFields = append(updateFields, "cod_barras = ?");       args = append(args, sqlNullString(input.CodBarras)) }
		if input.Precio1 != nil            { updateFields = append(updateFields, "precio1 = ?");          args = append(args, sqlNullFloat64(input.Precio1)) }
		if input.Precio2 != nil            { updateFields = append(updateFields, "precio2 = ?");          args = append(args, sqlNullFloat64(input.Precio2)) }
		if input.Precio3 != nil            { updateFields = append(updateFields, "precio3 = ?");          args = append(args, sqlNullFloat64(input.Precio3)) }
		if input.Precio4 != nil            { updateFields = append(updateFields, "precio4 = ?");          args = append(args, sqlNullFloat64(input.Precio4)) }
		if input.Precio5 != nil            { updateFields = append(updateFields, "precio5 = ?");          args = append(args, sqlNullFloat64(input.Precio5)) }
		if input.Precio6 != nil            { updateFields = append(updateFields, "precio6 = ?");          args = append(args, sqlNullFloat64(input.Precio6)) }
		if input.Precio7 != nil            { updateFields = append(updateFields, "precio7 = ?");          args = append(args, sqlNullFloat64(input.Precio7)) }
		if input.Precio8 != nil            { updateFields = append(updateFields, "precio8 = ?");          args = append(args, sqlNullFloat64(input.Precio8)) }
		if input.Precio9 != nil            { updateFields = append(updateFields, "precio9 = ?");          args = append(args, sqlNullFloat64(input.Precio9)) }
		if input.Precio10 != nil           { updateFields = append(updateFields, "precio10 = ?");         args = append(args, sqlNullFloat64(input.Precio10)) }
		if input.Precio11 != nil           { updateFields = append(updateFields, "precio11 = ?");         args = append(args, sqlNullFloat64(input.Precio11)) }
		if input.Precio12 != nil           { updateFields = append(updateFields, "precio12 = ?");         args = append(args, sqlNullFloat64(input.Precio12)) }
		if input.Precio13 != nil           { updateFields = append(updateFields, "precio13 = ?");         args = append(args, sqlNullFloat64(input.Precio13)) }
		if input.Precio14 != nil           { updateFields = append(updateFields, "precio14 = ?");         args = append(args, sqlNullFloat64(input.Precio14)) }
		if input.Precio15 != nil           { updateFields = append(updateFields, "precio15 = ?");         args = append(args, sqlNullFloat64(input.Precio15)) }
		if input.Precio16 != nil           { updateFields = append(updateFields, "precio16 = ?");         args = append(args, sqlNullFloat64(input.Precio16)) }
		if input.Precio17 != nil           { updateFields = append(updateFields, "precio17 = ?");         args = append(args, sqlNullFloat64(input.Precio17)) }
		if input.Precio18 != nil           { updateFields = append(updateFields, "precio18 = ?");         args = append(args, sqlNullFloat64(input.Precio18)) }
		if input.Precio19 != nil           { updateFields = append(updateFields, "precio19 = ?");         args = append(args, sqlNullFloat64(input.Precio19)) }
		if input.Precio20 != nil           { updateFields = append(updateFields, "precio20 = ?");         args = append(args, sqlNullFloat64(input.Precio20)) }
		if input.Precio21 != nil           { updateFields = append(updateFields, "precio21 = ?");         args = append(args, sqlNullFloat64(input.Precio21)) }
		if input.Precio22 != nil           { updateFields = append(updateFields, "precio22 = ?");         args = append(args, sqlNullFloat64(input.Precio22)) }
		if input.Precio23 != nil           { updateFields = append(updateFields, "precio23 = ?");         args = append(args, sqlNullFloat64(input.Precio23)) }
		if input.Precio24 != nil           { updateFields = append(updateFields, "precio24 = ?");         args = append(args, sqlNullFloat64(input.Precio24)) }
		if input.Precio25 != nil           { updateFields = append(updateFields, "precio25 = ?");         args = append(args, sqlNullFloat64(input.Precio25)) }
		if input.IEPSAdic != nil           { updateFields = append(updateFields, "ieps_adic = ?");        args = append(args, sqlNullFloat64(input.IEPSAdic)) }
		if input.Unidad != nil             { updateFields = append(updateFields, "unidad = ?");           args = append(args, sqlNullString(input.Unidad)) }
		if input.UnidadEnt != nil          { updateFields = append(updateFields, "unidad_ent = ?");       args = append(args, sqlNullString(input.UnidadEnt)) }
		if input.FactorConversion != nil   { updateFields = append(updateFields, "factor_conversion = ?");args = append(args, sqlNullFloat64(input.FactorConversion)) }
		if input.SATClave != nil           { updateFields = append(updateFields, "sat_clave = ?");        args = append(args, sqlNullString(input.SATClave)) }
		if input.SATMedida != nil          { updateFields = append(updateFields, "sat_medida = ?");       args = append(args, sqlNullString(input.SATMedida)) }
		if input.Volumen != nil            { updateFields = append(updateFields, "volumen = ?");          args = append(args, sqlNullFloat64(input.Volumen)) }
		if input.Peso != nil               { updateFields = append(updateFields, "peso = ?");             args = append(args, sqlNullFloat64(input.Peso)) }
		if input.IDMoneda != nil           { updateFields = append(updateFields, "idmoneda = ?");         args = append(args, sqlNullInt64(input.IDMoneda)) }
		if input.Lote != nil               { updateFields = append(updateFields, "lote = ?");             args = append(args, sqlNullString(input.Lote)) }
		if input.DescTicket != nil         { updateFields = append(updateFields, "desc_ticket = ?");      args = append(args, sqlNullString(input.DescTicket)) }
		if input.CantSigLista != nil       { updateFields = append(updateFields, "cant_sig_lista = ?");   args = append(args, sqlNullInt64(input.CantSigLista)) }
		if input.EnVenta != nil            { updateFields = append(updateFields, "en_venta = ?");         args = append(args, sqlNullString(input.EnVenta)) }

		// Clave nunca se actualiza, siempre se mantiene igual
		updateFields = append(updateFields, "clave = ?")
		args = append(args, claveToSet)

		// Al final, WHERE
		args = append(args, id)

		query := "UPDATE crm_productos SET " + join(updateFields, ", ") + " WHERE idproducto = ?"

		_, err = dbConn.Local.Exec(query, args...)
		if err != nil {
			http.Error(w, "Error al editar producto: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"idproducto": id,
			"msg":        "Producto editado correctamente",
		})
	}
}

// GetEstatusProductos retorna productos filtrados por estatus (paginado, incluye IVA, precio final y nombre de empresa)
func GetEstatusProductos(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		estatus := r.URL.Query().Get("estatus")
		if estatus == "" {
			http.Error(w, "Falta el parámetro 'estatus'", http.StatusBadRequest)
			return
		}

		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 20
		}
		offset := (page - 1) * limit

		// Total para paginación
		var total int
		countQuery := "SELECT COUNT(*) FROM crm_productos WHERE estatus = ?"
		if err := dbConn.Local.QueryRow(countQuery, estatus).Scan(&total); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		query := `
			SELECT 
				p.idproducto, 
				p.idempresa, 
				e.nombre_comercial,
				p.idlinea, p.descripcion, p.estatus, p.tipo_prod, p.idcategoria, p.clasif, p.con_formula, p.clave, p.idiva, p.cod_barras,
				p.precio1, p.precio2, p.precio3, p.precio4, p.precio5, p.precio6, p.precio7, p.precio8, p.precio9, p.precio10, p.precio11, p.precio12,
				p.precio13, p.precio14, p.precio15, p.precio16, p.precio17, p.precio18, p.precio19, p.precio20, p.precio21, p.precio22, p.precio23,
				p.precio24, p.precio25, p.ieps_adic, p.con_ieps_adic, p.unidad, p.unidad_ent, p.factor_conversion, p.sat_clave, p.sat_medida, p.volumen,
				p.peso, p.idmoneda, p.lote, p.desc_ticket, p.cant_sig_lista, p.en_venta,
				IFNULL(i.iva, 0), IFNULL(i.tipo_iva, '')
			FROM crm_productos p
			LEFT JOIN adm_empresas e ON p.idempresa = e.idempresa
			LEFT JOIN crm_impuestos i ON p.idiva = i.idiva
			WHERE p.estatus = ?
			LIMIT ? OFFSET ?`
		rows, err := dbConn.Local.Query(query, estatus, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var productos []ProductoDBConIVA

		for rows.Next() {
			var p ProductoDBConIVA
			err := rows.Scan(
				&p.IDProducto, &p.IDEmpresa, &p.NombreEmpresa,
				&p.IDLinea, &p.Descripcion, &p.Estatus, &p.TipoProd, &p.IDCategoria, &p.Clasif, &p.ConFormula, &p.Clave, &p.IDIVA, &p.CodBarras,
				&p.Precio1, &p.Precio2, &p.Precio3, &p.Precio4, &p.Precio5, &p.Precio6, &p.Precio7, &p.Precio8, &p.Precio9, &p.Precio10, &p.Precio11, &p.Precio12,
				&p.Precio13, &p.Precio14, &p.Precio15, &p.Precio16, &p.Precio17, &p.Precio18, &p.Precio19, &p.Precio20, &p.Precio21, &p.Precio22, &p.Precio23,
				&p.Precio24, &p.Precio25, &p.IEPSAdic, &p.ConIEPSAdic, &p.Unidad, &p.UnidadEnt, &p.FactorConversion, &p.SATClave, &p.SATMedida, &p.Volumen,
				&p.Peso, &p.IDMoneda, &p.Lote, &p.DescTicket, &p.CantSigLista, &p.EnVenta,
				&p.IVA, &p.TipoIVA,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Calcula el precio final usando Precio1 y el IVA
			base := p.Precio1.Float64
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

// Auxiliares para nulos
func sqlNullString(p *string) sql.NullString {
	if p != nil {
		return sql.NullString{String: *p, Valid: true}
	}
	return sql.NullString{Valid: false}
}
func sqlNullInt64(p *int64) sql.NullInt64 {
	if p != nil {
		return sql.NullInt64{Int64: *p, Valid: true}
	}
	return sql.NullInt64{Valid: false}
}
func sqlNullFloat64(p *float64) sql.NullFloat64 {
	if p != nil {
		return sql.NullFloat64{Float64: *p, Valid: true}
	}
	return sql.NullFloat64{Valid: false}
}

// Ayuda para join de slice de strings
func join(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// Estructura de IVA
type Impuesto struct {
	IDIVA       int64           `json:"idiva"`
	Descripcion string          `json:"descripcion"`
	IVA         sql.NullFloat64 `json:"iva"`
	TipoIVA     sql.NullString  `json:"tipo_iva"`
}

// Handler para obtener IVAs por empresa
func GetImpuestosPorEmpresa(dbConn *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaStr := r.URL.Query().Get("empresa")
		if empresaStr == "" {
			http.Error(w, "Falta parámetro empresa", http.StatusBadRequest)
			return
		}
		empresaID, err := strconv.Atoi(empresaStr)
		if err != nil || empresaID <= 0 {
			http.Error(w, "ID de empresa inválido", http.StatusBadRequest)
			return
		}

		rows, err := dbConn.Local.Query(`
			SELECT idiva, descripcion, iva, tipo_iva 
			FROM crm_impuestos 
			WHERE idempresa = ?
			ORDER BY descripcion
		`, empresaID)
		if err != nil {
			http.Error(w, "Error de consulta: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var impuestos []Impuesto
		for rows.Next() {
			var imp Impuesto
			if err := rows.Scan(&imp.IDIVA, &imp.Descripcion, &imp.IVA, &imp.TipoIVA); err != nil {
				http.Error(w, "Error al leer datos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			impuestos = append(impuestos, imp)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(impuestos)
	}
}
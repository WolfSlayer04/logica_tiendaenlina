package rutas

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"onlinestore/db"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ---------------------------
// INICIALIZA RAND PARA CLAVES ALEATORIAS DIFERENTES
// ---------------------------
func init() {
	rand.Seed(time.Now().UnixNano())
}

// ---------------------------
// HELPERS RESPUESTA
// ---------------------------

func NullToInt(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0
}
func NullToFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0.0
}

// ---------------------------
// USUARIOS
// ---------------------------

type UsuarioRequest struct {
	TipoUsuario    string `json:"tipo_usuario"`
	NombreCompleto string `json:"nombre_completo"`
	Correo         string `json:"correo"`
	Telefono       string `json:"telefono"`
	Clave          string `json:"clave"`
}

type usuarioRow struct {
	IDUsuario      int
	IDEmpresa      int
	TipoUsuario    string
	NombreCompleto string
	Correo         string
	Telefono       string
	Estatus        string
	ClaveRemota    sql.NullString
	IDRemoto       sql.NullInt64
}

func encriptarContraseña(contraseña string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(contraseña), 10)
	if err != nil {
		return "", fmt.Errorf("error al encriptar la contraseña: %v", err)
	}
	return string(bytes), nil
}

func generarClaveAleatoria() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// ---------------------------
// TIENDAS
// ---------------------------

type TiendaRequest struct {
	IDUsuario    int     `json:"id_usuario"`
	NombreTienda string  `json:"nombre_tienda"`
	RazonSocial  string  `json:"razon_social"`
	RFC          string  `json:"rfc"`
	Direccion    string  `json:"direccion"`
	Colonia      string  `json:"colonia"`
	CodigoPostal string  `json:"codigo_postal"`
	Ciudad       string  `json:"ciudad"`
	Estado       string  `json:"estado"`
	Pais         string  `json:"pais"`
	TipoTienda   string  `json:"tipo_tienda"`
	Latitud      float64 `json:"latitud"`
	Longitud     float64 `json:"longitud"`
}

type tiendaData struct {
	IDTienda     int     `json:"id_tienda"`
	NombreTienda string  `json:"nombre_tienda"`
	Direccion    string  `json:"direccion"`
	Colonia      string  `json:"colonia"`
	CodigoPostal string  `json:"codigo_postal"`
	Ciudad       string  `json:"ciudad"`
	Estado       string  `json:"estado"`
	Pais         string  `json:"pais"`
	Latitud      float64 `json:"latitud"`
	Longitud     float64 `json:"longitud"`
	LatitudUbic  float64 `json:"latitud_ubic"`
	LongitudUbic float64 `json:"longitud_ubic"`
}

type tiendaRow struct {
	IDTienda       int
	IDUsuario      int
	IDEmpresa      int
	IDSucursal     sql.NullInt64
	NombreSucursal sql.NullString
	NombreTienda   string
	RazonSocial    string
	RFC            string
	Direccion      string
	Colonia        string
	CodigoPostal   string
	Ciudad         string
	Estado         string
	Pais           string
	TipoTienda     string
	Estatus        string
}

// ---------------------------
// CREAR CLIENTE REMOTO (nombre_comercial = nombre_tienda, razon_social)
// ---------------------------

func BuscaClienteRemoto(
	dbc *sql.DB,
	idsucursal int,
	clave string,
	nombreComercial string,
	razonSocial string,
	rfc string,
	direccion string,
	calle string,
	numero string,
	ciudad string,
	estado string,
	cp string,
	telefono string,
) (idCliente int64, err error) {
	// 1. Buscar por RFC si se proporciona
	if rfc != "" {
		err = dbc.QueryRow(`
			SELECT id_cliente FROM crm_clientes
			WHERE idsucursal = ? AND rfc = ? AND estatus = 'S' LIMIT 1
		`, idsucursal, rfc).Scan(&idCliente)
		if err == nil {
			return 0, fmt.Errorf("Ya existe un cliente remoto con este RFC")
		}
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("error buscando cliente por RFC: %w", err)
		}
	}

	// 2. Buscar por nombre_comercial si no se encontró por RFC
	if nombreComercial != "" {
		err = dbc.QueryRow(`
			SELECT id_cliente FROM crm_clientes
			WHERE idsucursal = ? AND nombre_comercial = ? AND estatus = 'S' LIMIT 1
		`, idsucursal, nombreComercial).Scan(&idCliente)
		if err == nil {
			return 0, fmt.Errorf("Ya existe un cliente remoto con este nombre comercial")
		}
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("error buscando cliente por nombre_comercial: %w", err)
		}
	}

	// 3. Si no existe, obtener la colonia desde dirección (como en tu PHP)
	colonia := ""
	partes := strings.Split(strings.ReplaceAll(direccion, "\"", ""), ",")
	for i := 0; i < len(partes); i++ {
		if strings.HasPrefix(strings.TrimSpace(partes[i]), cp) && i > 0 {
			colonia = strings.TrimSpace(partes[i-1])
			break
		}
	}

	// 4. Crear o actualizar correlativo en crm_indices
	_, err = dbc.Exec(`
		INSERT INTO crm_indices (idsucursal, idpedido, idcliente)
		VALUES (?, 0, 1)
		ON DUPLICATE KEY UPDATE idcliente = idcliente + 1
	`, idsucursal)
	if err != nil {
		return 0, fmt.Errorf("error actualizando/creando índice de cliente: %w", err)
	}

	// 5. Obtener el nuevo correlativo
	var correlativo int64
	err = dbc.QueryRow(`
		SELECT idcliente FROM crm_indices WHERE idsucursal = ?
	`, idsucursal).Scan(&correlativo)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo correlativo de cliente: %w", err)
	}

	// 6. Generar clave_mobile
	claveMobile := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d%s%s%d", time.Now().Unix(), clave, nombreComercial, rand.Intn(10000)))))

	// 7. Insertar cliente en crm_clientes (agregando razon_social)
	res, err := dbc.Exec(`
		INSERT INTO crm_clientes (
			id_cliente, idsucursal, estatus, clave_mobile, nombre_comercial, razon_social, rfc, idcliente, tipo_cliente, lista_precio,
			ofi_calle, ofi_num_ext, ofi_colonia, ofi_ciudad, ofi_estado, ofi_cod_postal, tel_contacto, clave
		) VALUES (
			0, ?, 'S', ?, ?, ?, ?, ?, 'C', 1,
			?, ?, ?, ?, ?, ?, ?, ?
		)
	`, idsucursal, claveMobile, nombreComercial, razonSocial, rfc, correlativo,
		calle, numero, colonia, ciudad, estado, cp, telefono, clave)
	if err != nil {
		return 0, fmt.Errorf("error insertando cliente remoto: %w", err)
	}
	idCliente, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo id_cliente insertado: %w", err)
	}
	return idCliente, nil
}

// ---------------------------
// Buscar sucursal por ubicación
// ---------------------------

// Cambia la firma de la función a recibir *db.DBConnection:
func BuscarSucursalPorUbicacion(dbconn *db.DBConnection, idEmpresa int, lat, lng float64) (idsucursal int, nombreSucursal string, err error) {
	rows, err := dbconn.Local.Query(`
        SELECT idsucursal, sucursal, tipo_objeto, radio
        FROM adm_sucursales
        WHERE idempresa = ? AND estatus = 'S'
    `, idEmpresa)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()

	var (
		bestID     int
		bestNombre string
		bestDist   float64 = math.MaxFloat64
		found      bool
	)

	for rows.Next() {
		var id int
		var nombre, tipoObjeto string
		var radio float64
		if err := rows.Scan(&id, &nombre, &tipoObjeto, &radio); err != nil {
			continue
		}
		// Ajusta la llamada a SeActivoEstaAlerta
		_, dist, err := db.SeActivoEstaAlerta(dbconn, tipoObjeto, nombre, id, "E", radio, lat, lng)
		if err != nil {
			continue
		}
		if dist == 0.0 {
			continue
		}
		if dist < bestDist {
			bestDist = dist
			bestID = id
			bestNombre = nombre
			found = true
		}
	}
	if !found {
		return 0, "", fmt.Errorf("No hay sucursales válidas")
	}
	return bestID, bestNombre, nil
}
// ---------------------------
// OBTENER USUARIO Y TIENDA COMO EN LOGIN
// ---------------------------

func GetUserAndTiendaForLogin(dbconn *sql.DB, idUsuario int64) (map[string]interface{}, error) {
	// Usuario
	var u usuarioRow
	err := dbconn.QueryRow(`SELECT id_usuario, id_empresa, tipo_usuario, nombre_completo, correo, telefono, estatus, clave_remota, id_remoto FROM usuarios WHERE id_usuario = ?`, idUsuario).
		Scan(&u.IDUsuario, &u.IDEmpresa, &u.TipoUsuario, &u.NombreCompleto, &u.Correo, &u.Telefono, &u.Estatus, &u.ClaveRemota, &u.IDRemoto)
	if err != nil {
		return nil, err
	}
	usuario := map[string]interface{}{
		"id_usuario":      u.IDUsuario,
		"id_empresa":      u.IDEmpresa,
		"tipo_usuario":    u.TipoUsuario,
		"nombre_completo": u.NombreCompleto,
		"correo":          u.Correo,
		"telefono":        u.Telefono,
		"estatus":         u.Estatus,
		"id_remoto":       NullToInt(u.IDRemoto),
		"clave_remota":    NullToStr(u.ClaveRemota),
	}

	// Tienda
	var tienda tiendaData
	var lat, lon, latPoint, lonPoint sql.NullFloat64
	err = dbconn.QueryRow(`
            SELECT id_tienda, nombre_tienda, direccion, colonia, codigo_postal, ciudad, estado, pais,
                   latitud, longitud, IFNULL(ST_Y(ubicacion), 0) as latitud_ubic, IFNULL(ST_X(ubicacion), 0) as longitud_ubic
            FROM tiendas
            WHERE id_usuario = ?
            LIMIT 1
        `, idUsuario).Scan(
		&tienda.IDTienda, &tienda.NombreTienda, &tienda.Direccion, &tienda.Colonia,
		&tienda.CodigoPostal, &tienda.Ciudad, &tienda.Estado, &tienda.Pais,
		&lat, &lon, &latPoint, &lonPoint,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	tienda.Latitud = NullToFloat(lat)
	tienda.Longitud = NullToFloat(lon)
	tienda.LatitudUbic = NullToFloat(latPoint)
	tienda.LongitudUbic = NullToFloat(lonPoint)

	return map[string]interface{}{
		"usuario": usuario,
		"tienda":  tienda,
	}, nil
}

// ---------------------------
// ENDPOINT: Registro combinado usuario+tienda
// ---------------------------

func RegistroUsuarioTienda(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Usuario UsuarioRequest `json:"usuario"`
			Tienda  TiendaRequest  `json:"tienda"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Datos inválidos", err.Error())
			return
		}

		claveEncriptada, err := encriptarContraseña(req.Usuario.Clave)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al procesar la contraseña", err.Error())
			return
		}

		var idEmpresa, idSucursal int
		err = dbc.Local.QueryRow("SELECT idempresa FROM adm_empresas WHERE estatus = 'S' LIMIT 1").Scan(&idEmpresa)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo obtener la empresa", err.Error())
			return
		}
		err = dbc.Local.QueryRow("SELECT idsucursal FROM adm_sucursales WHERE idempresa = ? AND estatus = 'S' LIMIT 1", idEmpresa).Scan(&idSucursal)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo obtener la sucursal", err.Error())
			return
		}

		// --- Generar clave aleatoria para remota ---
		claveAleatoria := generarClaveAleatoria()

		// --- Crear cliente remoto con datos de tienda, razon_social y claveAleatoria ---
		idClienteRemoto, err := BuscaClienteRemoto(
			dbc.Remote,
			idSucursal,
			claveAleatoria, // clave (6 dígitos) para remota
			req.Tienda.NombreTienda,
			req.Tienda.RazonSocial,
			req.Tienda.RFC,
			req.Tienda.Direccion,
			req.Tienda.Direccion,
			"",
			req.Tienda.Ciudad,
			req.Tienda.Estado,
			req.Tienda.CodigoPostal,
			req.Usuario.Telefono,
		)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo guardar cliente en remota", err.Error())
			return
		}

		// --- Guardar usuario en local (clave_remota = claveAleatoria) ---
		fechaRegistro := time.Now()
		userResult, err := dbc.Local.Exec(`
            INSERT INTO usuarios (
				id_empresa, tipo_usuario, nombre_completo, correo, telefono, clave, clave_remota, fecha_registro, estatus, requiere_cambiar_clave, id_remoto
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'activo', false, ?)
        `, idEmpresa, req.Usuario.TipoUsuario, req.Usuario.NombreCompleto, req.Usuario.Correo, req.Usuario.Telefono, claveEncriptada, claveAleatoria, fechaRegistro, idClienteRemoto)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al crear usuario", err.Error())
			return
		}
		idUsuario, _ := userResult.LastInsertId()

		// --- Crear tienda en local ---
		now := time.Now()
		idsucursal, nombreSucursal, err := BuscarSucursalPorUbicacion(dbc, idEmpresa, req.Tienda.Latitud, req.Tienda.Longitud) // Corregido aquí
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "No hay sucursal que cubra esa ubicación", "")
			return
		}
		_, err = dbc.Local.Exec(`
            INSERT INTO tiendas (
                id_usuario, id_empresa, idsucursal, nombre_sucursal, nombre_tienda, razon_social, rfc, direccion, colonia, codigo_postal, ciudad, estado, pais, tipo_tienda, latitud, longitud, ubicacion, fecha_registro, ultima_actualizacion, estatus
            ) VALUES (
                ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, POINT(?, ?), ?, ?, 'activo'
            )
        `,
			idUsuario,
			idEmpresa,
			idsucursal,
			nombreSucursal,
			req.Tienda.NombreTienda,
			req.Tienda.RazonSocial,
			req.Tienda.RFC,
			req.Tienda.Direccion,
			req.Tienda.Colonia,
			req.Tienda.CodigoPostal,
			req.Tienda.Ciudad,
			req.Tienda.Estado,
			req.Tienda.Pais,
			req.Tienda.TipoTienda,
			req.Tienda.Latitud,
			req.Tienda.Longitud,
			req.Tienda.Longitud, // X
			req.Tienda.Latitud,  // Y
			now, now,
		)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al crear tienda", err.Error())
			return
		}

		// --- Respuesta igual a login: usuario y tienda ---
		loginData, err := GetUserAndTiendaForLogin(dbc.Local, idUsuario)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "No se pudo obtener usuario y tienda para login automático", err.Error())
			return
		}
		writeSuccessResponse(w, "Usuario y tienda creados correctamente", loginData)
	}
}

// ---------------------------
// ENDPOINTS GET USUARIOS Y TIENDAS
// ---------------------------

func GetUsuarios(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dbc.Local.Query(`SELECT id_usuario, id_empresa, tipo_usuario, nombre_completo, correo, telefono, estatus, id_remoto, clave_remota FROM usuarios`)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener usuarios", err.Error())
			return
		}
		defer rows.Close()

		var usuarios []map[string]interface{}
		for rows.Next() {
			var u usuarioRow
			var idRemoto sql.NullInt64
			var claveRemota sql.NullString
			err := rows.Scan(&u.IDUsuario, &u.IDEmpresa, &u.TipoUsuario, &u.NombreCompleto, &u.Correo, &u.Telefono, &u.Estatus, &idRemoto, &claveRemota)
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al leer usuario", err.Error())
				return
			}
			usuarios = append(usuarios, map[string]interface{}{
				"id_usuario":      u.IDUsuario,
				"id_empresa":      u.IDEmpresa,
				"tipo_usuario":    u.TipoUsuario,
				"nombre_completo": u.NombreCompleto,
				"correo":          u.Correo,
				"telefono":        u.Telefono,
				"estatus":         u.Estatus,
				"id_remoto":       NullToInt(idRemoto),
				"clave_remota":    NullToStr(claveRemota),
			})
		}
		writeSuccessResponse(w, "Usuarios obtenidos correctamente", usuarios)
	}
}

func GetTiendas(dbc *db.DBConnection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dbc.Local.Query(`
            SELECT t.id_tienda, t.id_usuario, t.id_empresa, t.idsucursal, t.nombre_sucursal, t.nombre_tienda, t.razon_social, t.rfc, t.direccion, t.colonia, t.codigo_postal, t.ciudad, t.estado, t.pais, t.tipo_tienda, t.estatus
            FROM tiendas t
        `)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Error al obtener tiendas", err.Error())
			return
		}
		defer rows.Close()

		var tiendas []map[string]interface{}
		for rows.Next() {
			var t tiendaRow
			err := rows.Scan(
				&t.IDTienda, &t.IDUsuario, &t.IDEmpresa, &t.IDSucursal, &t.NombreSucursal, &t.NombreTienda, &t.RazonSocial, &t.RFC, &t.Direccion, &t.Colonia, &t.CodigoPostal, &t.Ciudad, &t.Estado, &t.Pais, &t.TipoTienda, &t.Estatus,
			)
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Error al leer tienda", err.Error())
				return
			}
			tiendas = append(tiendas, map[string]interface{}{
				"id_tienda":       t.IDTienda,
				"id_usuario":      t.IDUsuario,
				"id_empresa":      t.IDEmpresa,
				"idsucursal":      NullToInt(t.IDSucursal),
				"nombre_sucursal": NullToStr(t.NombreSucursal),
				"nombre_tienda":   t.NombreTienda,
				"razon_social":    t.RazonSocial,
				"rfc":             t.RFC,
				"direccion":       t.Direccion,
				"colonia":         t.Colonia,
				"codigo_postal":   t.CodigoPostal,
				"ciudad":          t.Ciudad,
				"estado":          t.Estado,
				"pais":            t.Pais,
				"tipo_tienda":     t.TipoTienda,
				"estatus":         t.Estatus,
			})
		}
		writeSuccessResponse(w, "Tiendas obtenidas correctamente", tiendas)
	}
}
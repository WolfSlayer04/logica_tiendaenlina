package rutas

import (
    "net/http"
    "github.com/WolfSlayer04/logica_tiendaenlina/db"
    "strconv"
    "database/sql"
)

func GetTiendaByUsuario(dbc *db.DBConnection) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        usuarioIDStr := r.URL.Query().Get("usuario")
        usuarioID, err := strconv.Atoi(usuarioIDStr)
        if err != nil || usuarioID <= 0 {
            writeErrorResponse(w, http.StatusBadRequest, "ID de usuario inválido", "")
            return
        }

        // Incluye latitud, longitud y extrae de ST_Y/ST_X para POINT.
        row := dbc.Local.QueryRow(`
            SELECT 
                id_tienda, id_usuario, id_empresa, nombre_tienda, razon_social, rfc, direccion, colonia, codigo_postal, ciudad, estado, pais, tipo_tienda, estatus,
                latitud, longitud, 
                IFNULL(ST_Y(ubicacion), 0) AS latitud_ubic, IFNULL(ST_X(ubicacion), 0) AS longitud_ubic
            FROM tiendas 
            WHERE id_usuario = ? LIMIT 1
        `, usuarioID)

        var t tiendaRow
        var lat sql.NullFloat64
        var lon sql.NullFloat64
        var latPoint sql.NullFloat64
        var lonPoint sql.NullFloat64
        err = row.Scan(
            &t.IDTienda, &t.IDUsuario, &t.IDEmpresa, &t.NombreTienda, &t.RazonSocial, &t.RFC, &t.Direccion, &t.Colonia, &t.CodigoPostal,
            &t.Ciudad, &t.Estado, &t.Pais, &t.TipoTienda, &t.Estatus,
            &lat, &lon, &latPoint, &lonPoint,
        )
        if err != nil {
            writeErrorResponse(w, http.StatusNotFound, "No se encontró la tienda para este usuario", err.Error())
            return
        }

        tienda := map[string]interface{}{
            "id_tienda":     t.IDTienda,
            "id_usuario":    t.IDUsuario,
            "id_empresa":    t.IDEmpresa,
            "nombre_tienda": t.NombreTienda,
            "razon_social":  t.RazonSocial,
            "rfc":           t.RFC,
            "direccion":     t.Direccion,
            "colonia":       t.Colonia,
            "codigo_postal": t.CodigoPostal,
            "ciudad":        t.Ciudad,
            "estado":        t.Estado,
            "pais":          t.Pais,
            "tipo_tienda":   t.TipoTienda,
            "estatus":       t.Estatus,
            "latitud":       db.NullToFloat(lat),
            "longitud":      db.NullToFloat(lon),
            "latitud_ubic":  db.NullToFloat(latPoint), // del campo POINT
            "longitud_ubic": db.NullToFloat(lonPoint), // del campo POINT
        }
        writeSuccessResponse(w, "Tienda obtenida correctamente", tienda)
    }
}
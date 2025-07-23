package db

import (
	"database/sql"
	"fmt"
	"math"
)

// SeActivoEstaAlerta determina si una posición cae dentro de una geocerca (círculo, rectángulo, polígono) definida en adm_sucursales_ptos.
// Recibe el singleton dbConn y la base ("local" o "remote").
// Ahora retorna: (bool, distancia al centro, error)
func SeActivoEstaAlerta(dbConn *DBConnection, base string, tipo_objeto string, idsucursal int, aviso string, radio float64, latitud float64, longitud float64) (bool, float64, error) {
	var db *sql.DB
	if base == "local" {
		db = dbConn.Local
	} else {
		db = dbConn.Remote
	}
	switch tipo_objeto {
	case "P":
		return ValidaPoligono(db, idsucursal, aviso, latitud, longitud)
	case "R":
		return ValidaRectangulo(db, idsucursal, aviso, latitud, longitud)
	case "C":
		return ValidaCirculo(db, idsucursal, aviso, latitud, longitud, radio)
	default:
		return false, 0, nil // Si el tipo de objeto es "N" o desconocido, retorna false sin error
	}
}

// Las demás funciones no cambian, porque reciben *sql.DB, que ahora es la conexión única del singleton.

func ValidaCirculo(db *sql.DB, idsucursal int, aviso string, latitud float64, longitud float64, radio float64) (bool, float64, error) {
	var x1, y1 float64
	query := `SELECT ST_X(punto) as latitud, ST_Y(punto) as longitud FROM adm_sucursales_ptos WHERE idsucursal = ? ORDER BY orden LIMIT 1`
	err := db.QueryRow(query, idsucursal).Scan(&x1, &y1)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, fmt.Errorf("no se encontró ningún punto para la sucursal (ValidaCirculo) con idsucursal %v", idsucursal)
		}
		return false, 0, fmt.Errorf("consulta SQL error: %v", err)
	}

	dist := distance(x1, y1, latitud, longitud)
	if aviso == "E" {
		return dist <= radio, dist, nil
	} else {
		return dist > radio, dist, nil
	}
}

func ValidaRectangulo(db *sql.DB, idsucursal int, aviso string, latitud float64, longitud float64) (bool, float64, error) {
	var lat1, lng1, lat2, lng2 float64
	rows, err := db.Query(`SELECT ST_X(punto), ST_Y(punto) FROM adm_sucursales_ptos WHERE idsucursal = ? ORDER BY orden LIMIT 2`, idsucursal)
	if err != nil {
		return false, 0, fmt.Errorf("error en la consulta SQL: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var lat, lng float64
		if err := rows.Scan(&lat, &lng); err != nil {
			return false, 0, fmt.Errorf("error al leer lat/lng: %v", err)
		}
		if count == 0 {
			lat1, lng1 = lat, lng
		} else {
			lat2, lng2 = lat, lng
		}
		count++
	}
	if count < 2 {
		return false, 0, fmt.Errorf("menos de 2 puntos en la sucursal para el rectángulo")
	}

	cadena := fmt.Sprintf(
		"%f %f, %f %f, %f %f, %f %f, %f %f",
		lat1, lng1, lat1, lng2, lat2, lng2, lat2, lng1, lat1, lng1,
	)
	string_query := fmt.Sprintf(
		"SELECT ST_Contains(ST_GeomFromText('POLYGON((%s))'), ST_GeomFromText('POINT(%f %f)')) as esta",
		cadena, latitud, longitud,
	)

	var esta int
	err = db.QueryRow(string_query).Scan(&esta)
	if err != nil {
		return false, 0, fmt.Errorf("error en ST_Contains: %v", err)
	}

	// Usamos el primer punto como centro aproximado del rectángulo para la distancia
	dist := distance(lat1, lng1, latitud, longitud)

	if aviso == "E" {
		return esta == 1, dist, nil
	} else {
		return esta == 0, dist, nil
	}
}

func ValidaPoligono(db *sql.DB, idsucursal int, aviso string, latitud float64, longitud float64) (bool, float64, error) {
	rows, err := db.Query(`SELECT ST_X(punto), ST_Y(punto) FROM adm_sucursales_ptos WHERE idsucursal = ? ORDER BY orden`, idsucursal)
	if err != nil {
		return false, 0, fmt.Errorf("error al buscar puntos del polígono: %v", err)
	}
	defer rows.Close()

	var cadena string
	var lat0, lng0 float64
	first := true
	for rows.Next() {
		var lat, lng float64
		if err := rows.Scan(&lat, &lng); err != nil {
			return false, 0, fmt.Errorf("error inesperado: %v", err)
		}
		if first {
			lat0 = lat
			lng0 = lng
			first = false
		}
		cadena += fmt.Sprintf("%f %f,", lat, lng)
	}
	if first {
		return false, 0, fmt.Errorf("no hay puntos para el polígono de la sucursal %d", idsucursal)
	}
	cadena += fmt.Sprintf("%f %f", lat0, lng0)

	string_query := fmt.Sprintf(
		"SELECT ST_Contains(ST_GeomFromText('POLYGON((%s))'), ST_GeomFromText('POINT(%f %f)')) as esta",
		cadena, latitud, longitud,
	)
	var esta int
	err = db.QueryRow(string_query).Scan(&esta)
	if err != nil {
		return false, 0, fmt.Errorf("error en ST_Contains: %v", err)
	}

	// Usamos el primer punto como centro aproximado del polígono para la distancia
	dist := distance(lat0, lng0, latitud, longitud)

	if aviso == "E" {
		return esta == 1, dist, nil
	} else {
		return esta == 0, dist, nil
	}
}

// distance calcula la distancia en kilómetros entre dos puntos geográficos (Haversine)
func distance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Radio de la Tierra en km
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
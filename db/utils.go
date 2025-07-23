package db

import (
	"database/sql"
	"time"
)

// NullToStr convierte sql.NullString a string ("" si es nulo)
func NullToStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// NullToStrOrNil convierte sql.NullString a interface{} (string o nil para SQL)
func NullToStrOrNil(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

// NullToInt convierte sql.NullInt64 a int (0 si es nulo)
func NullToInt(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0
}

// NullToIntOrNil convierte sql.NullInt64 a interface{} (int64 o nil para SQL)
func NullToIntOrNil(ni sql.NullInt64) interface{} {
	if ni.Valid {
		return ni.Int64
	}
	return nil
}

// NullToFloat convierte sql.NullFloat64 a float64 (0.0 si es nulo)
func NullToFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0.0
}

// NullToFloatOrNil convierte sql.NullFloat64 a interface{} (float64 o nil para SQL)
func NullToFloatOrNil(nf sql.NullFloat64) interface{} {
	if nf.Valid {
		return nf.Float64
	}
	return nil
}

// NullToTime convierte sql.NullTime a time.Time (zero time si es nulo)
func NullToTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

// NullToTimeOrNil convierte sql.NullTime a interface{} (time.Time o nil para SQL)
func NullToTimeOrNil(nt sql.NullTime) interface{} {
	if nt.Valid {
		return nt.Time
	}
	return nil
}

// NullToBool convierte sql.NullBool a bool (false si es nulo)
func NullToBool(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

// NullToBoolOrNil convierte sql.NullBool a interface{} (bool o nil para SQL)
func NullToBoolOrNil(nb sql.NullBool) interface{} {
	if nb.Valid {
		return nb.Bool
	}
	return nil
}
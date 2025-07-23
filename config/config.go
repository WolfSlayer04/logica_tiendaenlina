package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type DBConfig struct {
	User     string
	Password string
	DBName   string
	Host     string
	Port     int
}

type Configuration struct {
	LocalDB  DBConfig
	RemoteDB DBConfig
}

var Config Configuration

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() error {
	// Intentar cargar el archivo .env
	if err := loadEnvFile(".env"); err != nil {
		log.Printf("Advertencia: %v", err)
	}

	var missingVars []string

	// Cargar configuración de la base de datos local
	Config.LocalDB.User = os.Getenv("LOCAL_DB_USER")
	if Config.LocalDB.User == "" {
		missingVars = append(missingVars, "LOCAL_DB_USER")
	}

	Config.LocalDB.Password = os.Getenv("LOCAL_DB_PASSWORD")
	if Config.LocalDB.Password == "" {
		missingVars = append(missingVars, "LOCAL_DB_PASSWORD")
	}

	Config.LocalDB.DBName = os.Getenv("LOCAL_DB_NAME")
	if Config.LocalDB.DBName == "" {
		missingVars = append(missingVars, "LOCAL_DB_NAME")
	}

	Config.LocalDB.Host = os.Getenv("LOCAL_DB_HOST")
	if Config.LocalDB.Host == "" {
		missingVars = append(missingVars, "LOCAL_DB_HOST")
	}

	portStr := os.Getenv("LOCAL_DB_PORT")
	if portStr == "" {
		missingVars = append(missingVars, "LOCAL_DB_PORT")
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("LOCAL_DB_PORT no es un número válido: %w", err)
		}
		Config.LocalDB.Port = port
	}

	// Cargar configuración de la base de datos remota
	Config.RemoteDB.User = os.Getenv("REMOTE_DB_USER")
	if Config.RemoteDB.User == "" {
		missingVars = append(missingVars, "REMOTE_DB_USER")
	}

	Config.RemoteDB.Password = os.Getenv("REMOTE_DB_PASSWORD")
	if Config.RemoteDB.Password == "" {
		missingVars = append(missingVars, "REMOTE_DB_PASSWORD")
	}

	Config.RemoteDB.DBName = os.Getenv("REMOTE_DB_NAME")
	if Config.RemoteDB.DBName == "" {
		missingVars = append(missingVars, "REMOTE_DB_NAME")
	}

	Config.RemoteDB.Host = os.Getenv("REMOTE_DB_HOST")
	if Config.RemoteDB.Host == "" {
		missingVars = append(missingVars, "REMOTE_DB_HOST")
	}

	portStr = os.Getenv("REMOTE_DB_PORT")
	if portStr == "" {
		missingVars = append(missingVars, "REMOTE_DB_PORT")
	} else {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("REMOTE_DB_PORT no es un número válido: %w", err)
		}
		Config.RemoteDB.Port = port
	}

	// Verificar si faltan variables de entorno
	if len(missingVars) > 0 {
		return fmt.Errorf("faltan las siguientes variables de entorno: %s", strings.Join(missingVars, ", "))
	}

	return nil
}

// loadEnvFile carga variables desde un archivo .env
func loadEnvFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("no se pudo leer el archivo %s: %w", filename, err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		// Ignorar comentarios y líneas vacías
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Dividir por el primer signo igual
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// No sobrescribir si ya existe
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
	
	return nil
}

// GetLocalDBConnectionString devuelve la cadena de conexión para la BD local
func GetLocalDBConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", 
		Config.LocalDB.User, 
		Config.LocalDB.Password, 
		Config.LocalDB.Host, 
		Config.LocalDB.Port, 
		Config.LocalDB.DBName)
}

// GetRemoteDBConnectionString devuelve la cadena de conexión para la BD remota
func GetRemoteDBConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", 
		Config.RemoteDB.User, 
		Config.RemoteDB.Password, 
		Config.RemoteDB.Host, 
		Config.RemoteDB.Port, 
		Config.RemoteDB.DBName)
}
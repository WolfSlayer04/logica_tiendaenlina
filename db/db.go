package db

import (
    "database/sql"
    "fmt"
    "os"
    "strconv"
    "strings"
    "sync"

    _ "github.com/go-sql-driver/mysql"
)

type DBConfig struct {
    User     string
    Password string
    DBName   string
    Host     string
    Port     int
}

type Config struct {
    LocalDB  DBConfig
    RemoteDB DBConfig
}

type DBConnection struct {
    Local  *sql.DB
    Remote *sql.DB
}

var (
    dbConnInstance *DBConnection
    once           sync.Once
)

// LoadConfig carga la configuración desde variables de entorno (.env)
func LoadConfig() (*Config, error) {
    // Cargar variables de entorno desde .env
    if err := loadEnvFile(".env"); err != nil {
        return nil, fmt.Errorf("error cargando variables de entorno: %v", err)
    }

    var cfg Config
    var missingVars []string

    // Configuración BD Local
    cfg.LocalDB.User = os.Getenv("LOCAL_DB_USER")
    if cfg.LocalDB.User == "" {
        missingVars = append(missingVars, "LOCAL_DB_USER")
    }

    cfg.LocalDB.Password = os.Getenv("LOCAL_DB_PASSWORD")
    if cfg.LocalDB.Password == "" {
        missingVars = append(missingVars, "LOCAL_DB_PASSWORD")
    }

    cfg.LocalDB.DBName = os.Getenv("LOCAL_DB_NAME")
    if cfg.LocalDB.DBName == "" {
        missingVars = append(missingVars, "LOCAL_DB_NAME")
    }

    cfg.LocalDB.Host = os.Getenv("LOCAL_DB_HOST")
    if cfg.LocalDB.Host == "" {
        missingVars = append(missingVars, "LOCAL_DB_HOST")
    }

    portStr := os.Getenv("LOCAL_DB_PORT")
    if portStr == "" {
        missingVars = append(missingVars, "LOCAL_DB_PORT")
    } else {
        port, err := strconv.Atoi(portStr)
        if err != nil {
            return nil, fmt.Errorf("LOCAL_DB_PORT no es un número válido: %v", err)
        }
        cfg.LocalDB.Port = port
    }

    // Configuración BD Remota
    cfg.RemoteDB.User = os.Getenv("REMOTE_DB_USER")
    if cfg.RemoteDB.User == "" {
        missingVars = append(missingVars, "REMOTE_DB_USER")
    }

    cfg.RemoteDB.Password = os.Getenv("REMOTE_DB_PASSWORD")
    if cfg.RemoteDB.Password == "" {
        missingVars = append(missingVars, "REMOTE_DB_PASSWORD")
    }

    cfg.RemoteDB.DBName = os.Getenv("REMOTE_DB_NAME")
    if cfg.RemoteDB.DBName == "" {
        missingVars = append(missingVars, "REMOTE_DB_NAME")
    }

    cfg.RemoteDB.Host = os.Getenv("REMOTE_DB_HOST")
    if cfg.RemoteDB.Host == "" {
        missingVars = append(missingVars, "REMOTE_DB_HOST")
    }

    portStr = os.Getenv("REMOTE_DB_PORT")
    if portStr == "" {
        missingVars = append(missingVars, "REMOTE_DB_PORT")
    } else {
        port, err := strconv.Atoi(portStr)
        if err != nil {
            return nil, fmt.Errorf("REMOTE_DB_PORT no es un número válido: %v", err)
        }
        cfg.RemoteDB.Port = port
    }

    if len(missingVars) > 0 {
        return nil, fmt.Errorf("faltan las siguientes variables de entorno: %s", strings.Join(missingVars, ", "))
    }

    return &cfg, nil
}

// loadEnvFile carga variables desde un archivo .env
func loadEnvFile(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("no se pudo leer el archivo %s: %v", filename, err)
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

        // Eliminar comillas si existen
        if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"' ||
            value[0] == '\'' && value[len(value)-1] == '\'') {
            value = value[1 : len(value)-1]
        }

        // No sobrescribir si ya existe
        if _, exists := os.LookupEnv(key); !exists {
            os.Setenv(key, value)
        }
    }

    return nil
}

// ConnectDB crea y devuelve conexiones a las bases de datos usando variables de entorno
func ConnectDB(configPath string) (*DBConnection, error) {
    // Cargar configuración desde .env
    cfg, err := LoadConfig()
    if err != nil {
        return nil, err
    }

    // Conexión Local
    localDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        cfg.LocalDB.User,
        cfg.LocalDB.Password,
        cfg.LocalDB.Host,
        cfg.LocalDB.Port,
        cfg.LocalDB.DBName)

    localDB, err := sql.Open("mysql", localDsn)
    if err != nil {
        return nil, fmt.Errorf("error en conexión local: %v", err)
    }

    if err = localDB.Ping(); err != nil {
        localDB.Close()
        return nil, fmt.Errorf("error al verificar conexión local: %v", err)
    }

    // Conexión Remota
    remoteDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        cfg.RemoteDB.User,
        cfg.RemoteDB.Password,
        cfg.RemoteDB.Host,
        cfg.RemoteDB.Port,
        cfg.RemoteDB.DBName)

    remoteDB, err := sql.Open("mysql", remoteDsn)
    if err != nil {
        localDB.Close()
        return nil, fmt.Errorf("error en conexión remota: %v", err)
    }

    if err = remoteDB.Ping(); err != nil {
        localDB.Close()
        remoteDB.Close()
        return nil, fmt.Errorf("error al verificar conexión remota: %v", err)
    }

    // Configurar pool de conexiones
    localDB.SetMaxOpenConns(10)
    localDB.SetMaxIdleConns(5)
    remoteDB.SetMaxOpenConns(10)
    remoteDB.SetMaxIdleConns(5)

    return &DBConnection{
        Local:  localDB,
        Remote: remoteDB,
    }, nil
}

// GetDBConnection retorna la instancia singleton de DBConnection
func GetDBConnection() (*DBConnection, error) {
    var err error
    once.Do(func() {
        dbConnInstance, err = ConnectDB("")
    })
    return dbConnInstance, err
}

// Close cierra ambas conexiones a las bases de datos
func (dbc *DBConnection) Close() {
    if dbc.Local != nil {
        if err := dbc.Local.Close(); err != nil {
            fmt.Printf("Error al cerrar la conexión local: %v\n", err)
        }
    }
    if dbc.Remote != nil {
        if err := dbc.Remote.Close(); err != nil {
            fmt.Printf("Error al cerrar la conexión remota: %v\n", err)
        }
    }
}

// CheckConnections verifica el estado de ambas conexiones
func (dbc *DBConnection) CheckConnections() error {
    if err := dbc.Local.Ping(); err != nil {
        return fmt.Errorf("error en conexión local: %v", err)
    }
    if err := dbc.Remote.Ping(); err != nil {
        return fmt.Errorf("error en conexión remota: %v", err)
    }
    return nil
}
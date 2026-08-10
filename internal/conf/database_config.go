package conf

import "fmt"

// DatabaseConfig holds MySQL connection settings.
type DatabaseConfig struct {
	Driver   string `json:"driver" yaml:"driver"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	DBName   string `json:"dbname" yaml:"dbname"`
}

// DSN builds a MySQL connection string.
func (d *DatabaseConfig) DSN() string {
	if d == nil {
		return ""
	}
	host := d.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := d.Port
	if port == 0 {
		port = 3306
	}
	user := d.User
	if user == "" {
		user = "root"
	}
	dbName := d.DBName
	if dbName == "" {
		dbName = "backend"
	}
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, d.Password, host, port, dbName,
	)
}

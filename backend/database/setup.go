package database

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB    *gorm.DB // SQLite (Auth)
	MySQL *gorm.DB // MySQL (Produksi)
)

// --- KONFIGURASI DATABASE ---
const (
	DB_USER = "root"
	DB_PASS = ""
	DB_HOST = "127.0.0.1"
	DB_PORT = "3306"
	DB_NAME = "factory_db"
)

func ConnectDatabase() {
	var err error

	// 1. SQLITE (WAJIB untuk Login)
	// Jika ini gagal, aplikasi harus mati karena tidak bisa login
	DB, err = gorm.Open(sqlite.Open("besq.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal koneksi ke SQLite:", err)
	}
	log.Println("✅ Terhubung ke SQLite (User DB)")

	// 2. MYSQL (OPSIONAL untuk Test Local)
	// Jika gagal, aplikasi TETAP LANJUT agar bisa test login/UI
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		DB_USER, DB_PASS, DB_HOST, DB_PORT, DB_NAME,
	)

	MySQL, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("⚠️  PERINGATAN: Tidak dapat terhubung ke MySQL Kantor.")
		log.Println("    (Fitur Login tetap jalan, tapi Grafik akan error/kosong)")
		log.Println("    Error:", err)
		MySQL = nil // Set nil agar bisa dicek di controller
	} else {
		log.Println("✅ Terhubung ke MySQL Kantor (Read-Only)")
	}
}
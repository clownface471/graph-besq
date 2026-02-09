package database

import (
	"graph/backend/models"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func SeedUsers() {
	// Migrasi Schema User ke SQLite
	err := DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Gagal migrasi database user:", err)
	}

	// Cek apakah user sudah ada (supaya tidak duplikat setiap restart)
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Database user sudah berisi data, skip seeding.")
		return
	}

	log.Println("Memulai seeding data user dummy...")

	// Password default untuk semua akun: "123456"
	password, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	pwString := string(password)

	users := []models.User{
		// 1. MANAGER (Akses Level 1, 2, 3) - Tidak butuh departemen
		{Username: "manager", Password: pwString, Role: "MANAGER", Department: ""},

		// 2. LEADER (Akses Level 2, 3) - Departemen PRESSING
		{Username: "leader_prs", Password: pwString, Role: "LEADER", Department: "PRESSING"},

		// 3. OPERATOR (Akses Level 3) - Departemen PRESSING
		{Username: "op_prs_01", Password: pwString, Role: "OPERATOR", Department: "PRESSING"},
	}

	if err := DB.Create(&users).Error; err != nil {
		log.Fatal("Gagal seeding user:", err)
	}

	log.Println("✅ Berhasil membuat 3 akun dummy!")
	log.Println("   - Manager: manager / 123456")
	log.Println("   - Leader:  leader_prs  / 123456")
	log.Println("   - Operator: op_prs_01  / 123456")
}
package main

import (
	"graph/backend/controllers"
	"graph/backend/database"
	"graph/backend/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// CORS Setup
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 1. Konek Database
	database.ConnectDatabase()

	// 2. Jalankan Seeder
	database.SeedUsers()

	// 3. Route Public
	r.POST("/login", controllers.Login)

	// 4. API Chart
	chartApi := r.Group("/api/chart")
	{
		// Level 1: Manager
		chartApi.GET("/manager", 
			middleware.AuthAndRoleMiddleware("MANAGER"), 
			controllers.GetManagerOverview,
		)

		// Level 2: Leader/Process
		chartApi.GET("/process", 
			middleware.AuthAndRoleMiddleware("MANAGER", "LEADER"), 
			controllers.GetLeaderProcessView,
		)

		// Level 3: Machine Detail
		chartApi.GET("/machine", 
			middleware.AuthAndRoleMiddleware("MANAGER", "LEADER", "OPERATOR"), 
			controllers.GetMachineDetail,
		)

        // --- ENDPOINT BARU: LIST MESIN ---
        // Diakses oleh semua role yang bisa lihat chart
		chartApi.GET("/machines", 
			middleware.AuthAndRoleMiddleware("MANAGER", "LEADER", "OPERATOR"), 
			controllers.GetMachineList,
		)
	}

	r.Run(":8080")
}
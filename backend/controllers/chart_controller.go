package controllers

import (
	"graph/backend/database"
	"graph/backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Helper: Cek koneksi DB
func checkDB(c *gin.Context) bool {
	if database.MySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL Offline/Tidak Terhubung"})
		return false
	}
	return true
}

// 1. MANAGER OVERVIEW
// Karena tabel vtrx_lwp_prs kemungkinan besar khusus "Pressing",
// kita hitung data dari tabel itu untuk divisi Pressing, sisanya 0 (kecuali ada tabel lain).
func GetManagerOverview(c *gin.Context) {
	if !checkDB(c) { return }

	tanggal := c.Query("tanggal")
	var results []models.ChartSeries

	// Query khusus untuk Pressing berdasarkan tabel vtrx_lwp_prs
	// Asumsi: qty_actual adalah Total Output, target diambil dari kolom target_jam
	query := `
		SELECT 
			'PRESSING' AS label,
			COALESCE(SUM(target_jam), 0) AS target,
			COALESCE(SUM(qty_actual), 0) AS actual
		FROM vtrx_lwp_prs 
		WHERE tanggal = ?
	`
	// Kita buat variabel penampung sementara
	var pressingData models.ChartSeries
	
	if err := database.MySQL.Raw(query, tanggal).Scan(&pressingData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Masukkan data Pressing, dan dummy 0 untuk divisi lain (karena belum ada tabelnya)
	results = append(results, 
		models.ChartSeries{Label: "MIXING", Target: 0, Actual: 0},
		models.ChartSeries{Label: "CUTTING", Target: 0, Actual: 0},
		pressingData, // Data Asli dari MySQL
		models.ChartSeries{Label: "FINISHING", Target: 0, Actual: 0},
	)

	c.JSON(http.StatusOK, results)
}

// 2. LEADER / PROCESS VIEW (Detail per Mesin)
func GetLeaderProcessView(c *gin.Context) {
	if !checkDB(c) { return }

	tanggal := c.Query("tanggal")
	// proses := c.Query("proses") // Sementara diabaikan karena tabelnya cuma satu (PRS)

	var results []models.ChartSeries

	// Grouping berdasarkan kolom 'mesin'
	query := `
		SELECT 
			mesin AS label,
			COALESCE(SUM(target_jam), 0) AS target,
			COALESCE(SUM(qty_actual), 0) AS actual,
			COALESCE(SUM(qty_ng), 0) AS actual_ng
		FROM vtrx_lwp_prs
		WHERE tanggal = ?
		GROUP BY mesin
		ORDER BY mesin
	`

	if err := database.MySQL.Raw(query, tanggal).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// 3. MACHINE DETAIL (Detail per Jam)
func GetMachineDetail(c *gin.Context) {
	if !checkDB(c) { return }

	tanggal := c.Query("tanggal")
	noMC := c.Query("no_mc")
	var results []models.ChartSeries

	// Menggunakan kolom 'jam_ke' yang sudah ada di tabel, jauh lebih simpel!
	// Kita asumsikan qty_actual adalah Total, jadi OK = Actual - NG.
	query := `
		SELECT 
			jam_ke AS label,
			COALESCE(SUM(target_jam), 0) AS target,
			COALESCE(SUM(qty_actual), 0) AS actual,
			COALESCE(SUM(qty_actual) - SUM(qty_ng), 0) AS actual_ok,
			COALESCE(SUM(qty_ng), 0) AS actual_ng,
			COALESCE(MAX(item_name), '-') AS extra_info
		FROM vtrx_lwp_prs
		WHERE tanggal = ? AND mesin = ?
		GROUP BY jam_ke
		ORDER BY jam_ke ASC
	`

	if err := database.MySQL.Raw(query, tanggal, noMC).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// 4. GET LIST MESIN (Untuk Dropdown)
func GetMachineList(c *gin.Context) {
	if !checkDB(c) { return }

	tanggal := c.Query("tanggal")
	if tanggal == "" {
		tanggal = time.Now().Format("2006-01-02")
	}

	var machines []string
	// Ambil list mesin unik yang aktif pada tanggal tersebut
	query := "SELECT DISTINCT mesin FROM vtrx_lwp_prs WHERE tanggal = ? AND mesin != '' ORDER BY mesin"

	if err := database.MySQL.Raw(query, tanggal).Scan(&machines).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, machines)
}
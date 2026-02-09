package controllers

import (
	"dashboard-project/backend/database"
	"dashboard-project/backend/models"
	"fmt"
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

// Clause WHERE untuk merakit tanggal dari kolom thn, bln, tgl
// Format DB: thn=2026, bln=2, tgl=9
const whereDateClause = "CONCAT(t.thn, '-', LPAD(t.bln, 2, '0'), '-', LPAD(t.tgl, 2, '0')) = ?"

// 1. MANAGER OVERVIEW
// Menampilkan total output pabrik vs total target
func GetManagerOverview(c *gin.Context) {
	if !checkDB(c) { return }

	tanggal := c.Query("tanggal") // Format: YYYY-MM-DD
	var results []models.ChartSeries

	var actualPressing models.ChartSeries
	
	// JOIN vtrx_lwp_prs (t) dengan v_stdlot (s)
	// Menggunakan COLLATE agar aman jika collation DB beda
	query := fmt.Sprintf(`
		SELECT 
			'PRESSING' AS label,
			COALESCE(SUM(s.tgtQtyPJam), 0) AS target,
			COALESCE(SUM(t.Total), 0) AS actual
		FROM vtrx_lwp_prs t
		LEFT JOIN v_stdlot s ON t.moldcode = s.moldCode COLLATE utf8mb4_unicode_ci
		WHERE %s
	`, whereDateClause)
	
	if err := database.MySQL.Raw(query, tanggal).Scan(&actualPressing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal Query Manager: " + err.Error()})
		return
	}
	actualPressing.Label = "PRESSING"

	// Data Dummy untuk departemen lain
	results = append(results, 
		models.ChartSeries{Label: "MIXING", Target: 0, Actual: 0},
		models.ChartSeries{Label: "CUTTING", Target: 0, Actual: 0},
		actualPressing, // Data Real Pressing
		models.ChartSeries{Label: "FINISHING", Target: 0, Actual: 0},
	)

	c.JSON(http.StatusOK, results)
}

// 2. LEADER / PROCESS VIEW (Detail per Mold/Mesin)
func GetLeaderProcessView(c *gin.Context) {
	if !checkDB(c) { return }

	tanggal := c.Query("tanggal")
	var results []models.ChartSeries

	// Grouping berdasarkan Mold
	// Label: Nama Mold (moldName) atau Kode Mold (moldcode)
	query := fmt.Sprintf(`
		SELECT 
			COALESCE(s.moldName, t.moldcode) AS label,
			COALESCE(SUM(s.tgtQtyPJam), 0) AS target,
			COALESCE(SUM(t.Total), 0) AS actual,
			COALESCE(SUM(t.NG), 0) AS actual_ng
		FROM vtrx_lwp_prs t
		LEFT JOIN v_stdlot s ON t.moldcode = s.moldCode COLLATE utf8mb4_unicode_ci
		WHERE %s
		GROUP BY t.moldcode, s.moldName
		ORDER BY label
	`, whereDateClause)

	if err := database.MySQL.Raw(query, tanggal).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal Query Leader: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// 3. MACHINE DETAIL (Detail per Jam untuk 1 Mold)
func GetMachineDetail(c *gin.Context) {
	if !checkDB(c) { return }

	tanggal := c.Query("tanggal")
	moldCode := c.Query("no_mc") // Parameter URL no_mc berisi moldcode
	var results []models.ChartSeries

	// Detail Per Jam
	query := fmt.Sprintf(`
		SELECT 
			CONCAT(LPAD(t.jam, 2, '0'), ':00') AS label,
			COALESCE(MAX(s.tgtQtyPJam), 0) AS target,
			COALESCE(SUM(t.Total), 0) AS actual,
			COALESCE(SUM(t.OK), 0) AS actual_ok,
			COALESCE(SUM(t.NG), 0) AS actual_ng,
			COALESCE(MAX(s.itemName), '-') AS extra_info
		FROM vtrx_lwp_prs t
		LEFT JOIN v_stdlot s ON t.moldcode = s.moldCode COLLATE utf8mb4_unicode_ci
		WHERE %s AND t.moldcode = ?
		GROUP BY t.jam
		ORDER BY t.jam ASC
	`, whereDateClause)

	if err := database.MySQL.Raw(query, tanggal, moldCode).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal Query Detail: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// 4. GET LIST MESIN (List Mold yang Aktif)
func GetMachineList(c *gin.Context) {
	if !checkDB(c) { return }

	tanggal := c.Query("tanggal")
	if tanggal == "" {
		tanggal = time.Now().Format("2006-01-02")
	}

	var machines []string
	
	// Mengambil daftar moldcode unik yang aktif hari ini
	query := fmt.Sprintf(`
		SELECT DISTINCT t.moldcode 
		FROM vtrx_lwp_prs t
		WHERE %s AND t.moldcode != '' 
		ORDER BY t.moldcode
	`, whereDateClause)

	if err := database.MySQL.Raw(query, tanggal).Scan(&machines).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal List Mesin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, machines)
}

// --- DEBUG TOOL ---
func CheckTableStructure(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Debug Tool Active"})
}
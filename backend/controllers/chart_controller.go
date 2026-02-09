package controllers

import (
	"graph/backend/database"
	"graph/backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetManagerOverview(c *gin.Context) {
	// Cek koneksi DB dulu
	if database.MySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL Offline (Mode Testing)"})
		return
	}

	tanggal := c.Query("tanggal")
	var results []models.ChartSeries

	query := `
		SELECT 
			t.proses AS label,
			COALESCE(SUM((TIME_TO_SEC(TIMEDIFF(t.SELESAI, t.MULAI)) / 3600.0) * s.tgtQtyPJam), 0) AS target,
			COALESCE(SUM(t.Total), 0) AS actual
		FROM vtrx_lwp_prs t
		LEFT JOIN v_stdlot s ON t.moldcode = s.moldCode COLLATE utf8mb4_unicode_ci
		WHERE t.tanggal = ?
		GROUP BY t.proses
	`

	if err := database.MySQL.Raw(query, tanggal).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func GetLeaderProcessView(c *gin.Context) {
	if database.MySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL Offline (Mode Testing)"})
		return
	}

	tanggal := c.Query("tanggal")
	proses := c.Query("proses")
	var results []models.ChartSeries

	query := `
		SELECT 
			t.noMC AS label,
			COALESCE(SUM((TIME_TO_SEC(TIMEDIFF(t.SELESAI, t.MULAI)) / 3600.0) * s.tgtQtyPJam), 0) AS target,
			COALESCE(SUM(t.Total), 0) AS actual
		FROM vtrx_lwp_prs t
		LEFT JOIN v_stdlot s ON t.moldcode = s.moldCode COLLATE utf8mb4_unicode_ci
		WHERE t.tanggal = ? AND t.proses = ?
		GROUP BY t.noMC
		ORDER BY t.noMC
	`

	if err := database.MySQL.Raw(query, tanggal, proses).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func GetMachineDetail(c *gin.Context) {
	if database.MySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL Offline (Mode Testing)"})
		return
	}

	tanggal := c.Query("tanggal")
	noMC := c.Query("no_mc")
	var results []models.ChartSeries

	query := `
	WITH RECURSIVE 
    jam_master AS (
        SELECT 0 AS jam_angka
        UNION ALL
        SELECT jam_angka + 1 FROM jam_master WHERE jam_angka < 23
    ),
    raw_data AS (
        SELECT 
            t.tanggal, t.nama AS operator, t.MULAI, t.SELESAI,
            s.itemName, s.tgtQtyPJam, j.jam_angka,
            ROUND(t.Total * (GREATEST(0, TIME_TO_SEC(TIMEDIFF(LEAST(t.SELESAI, MAKETIME(j.jam_angka + 1, 0, 0)), GREATEST(t.MULAI, MAKETIME(j.jam_angka, 0, 0))))) / NULLIF(TIME_TO_SEC(TIMEDIFF(t.SELESAI, t.MULAI)), 0))) AS allocated_total,
            ROUND(t.OK * (GREATEST(0, TIME_TO_SEC(TIMEDIFF(LEAST(t.SELESAI, MAKETIME(j.jam_angka + 1, 0, 0)), GREATEST(t.MULAI, MAKETIME(j.jam_angka, 0, 0))))) / NULLIF(TIME_TO_SEC(TIMEDIFF(t.SELESAI, t.MULAI)), 0))) AS allocated_ok,
            ROUND(t.NG * (GREATEST(0, TIME_TO_SEC(TIMEDIFF(LEAST(t.SELESAI, MAKETIME(j.jam_angka + 1, 0, 0)), GREATEST(t.MULAI, MAKETIME(j.jam_angka, 0, 0))))) / NULLIF(TIME_TO_SEC(TIMEDIFF(t.SELESAI, t.MULAI)), 0))) AS allocated_ng
        FROM vtrx_lwp_prs t
        LEFT JOIN v_stdlot s ON t.moldcode = s.moldCode COLLATE utf8mb4_unicode_ci
        CROSS JOIN jam_master j
        WHERE 
            t.tanggal = ? 
            AND t.noMC = ? 
            AND t.MULAI < MAKETIME(j.jam_angka + 1, 0, 0) 
            AND t.SELESAI > MAKETIME(j.jam_angka, 0, 0)
    )
    SELECT 
        CONCAT(LPAD(jam_angka, 2, '0'), ':00') AS label,
        COALESCE(MAX(tgtQtyPJam), 0) AS target,
        COALESCE(SUM(allocated_total), 0) AS actual,
        COALESCE(SUM(allocated_ok), 0) AS actual_ok,
        COALESCE(SUM(allocated_ng), 0) AS actual_ng,
        CONCAT(COALESCE(MAX(itemName), '-'), ' (', COALESCE(MAX(operator), '-'), ')') AS extra_info
    FROM raw_data
    GROUP BY jam_angka
    ORDER BY jam_angka ASC;
	`

	if err := database.MySQL.Raw(query, tanggal, noMC).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetMachineList mengambil daftar semua mesin yang unik dari database
func GetMachineList(c *gin.Context) {
	// Cek koneksi DB
	if database.MySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MySQL Offline"})
		return
	}

	var machines []string
	// Menggunakan DISTINCT untuk mendapatkan list mesin unik yang pernah beroperasi
	query := "SELECT DISTINCT noMC FROM vtrx_lwp_prs ORDER BY noMC"

	if err := database.MySQL.Raw(query).Scan(&machines).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, machines)
}
<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { auth } from '$lib/stores/auth';
    import Chart from 'chart.js/auto';
    import annotationPlugin from 'chartjs-plugin-annotation';

    Chart.register(annotationPlugin);

    let selectedDate = new Date().toISOString().split('T')[0];
    let isLoading = true;
    let isOffline = false;
    
    const API_URL = 'http://localhost:8080';

    let dashboardData: Record<string, { total: number; pending: number; completed: number }> = {
        mixing: { total: 0, pending: 0, completed: 0 },
        cutting: { total: 0, pending: 0, completed: 0 },
        pressing: { total: 0, pending: 0, completed: 0 },
        finishing: { total: 0, pending: 0, completed: 0 }
    };

    let charts: { [key: string]: { total?: Chart; status?: Chart } } = {};

    const divisions = [
        {
            id: 'mixing',
            name: 'Mixing',
            icon: 'fa-solid fa-blender',
            color: 'from-blue-500 to-cyan-500',
            bgColor: 'bg-blue-50',
            textColor: 'text-blue-700',
            description: 'Proses pencampuran bahan baku',
        },
        {
            id: 'cutting',
            name: 'Cutting',
            icon: 'fa-solid fa-scissors',
            color: 'from-purple-500 to-pink-500',
            bgColor: 'bg-purple-50',
            textColor: 'text-purple-700',
            description: 'Proses pemotongan material',
        },
        {
            id: 'pressing',
            name: 'Pressing',
            icon: 'fa-solid fa-square',
            color: 'from-orange-500 to-red-500',
            bgColor: 'bg-orange-50',
            textColor: 'text-orange-700',
            description: 'Proses pengepresan produk',
        },
        {
            id: 'finishing',
            name: 'Finishing',
            icon: 'fa-solid fa-star',
            color: 'from-green-500 to-emerald-500',
            bgColor: 'bg-green-50',
            textColor: 'text-green-700',
            description: 'Proses penyelesaian akhir',
        }
    ];

    async function fetchDashboardData() {
        isLoading = true;
        isOffline = false;
        try {
            const res = await fetch(`${API_URL}/api/chart/manager?tanggal=${selectedDate}`, {
                headers: { Authorization: `Bearer ${$auth.token}` }
            });

            if (res.status === 503) {
                isOffline = true;
            } else if (res.ok) {
                const result = await res.json();
                
                const newData: Record<string, { total: number; pending: number; completed: number }> = {
                    mixing: { total: 0, pending: 0, completed: 0 },
                    cutting: { total: 0, pending: 0, completed: 0 },
                    pressing: { total: 0, pending: 0, completed: 0 },
                    finishing: { total: 0, pending: 0, completed: 0 }
                };

                result.forEach((item: any) => {
                    let key = '';
                    const label = item.label ? item.label.toUpperCase() : '';

                    if (label.includes('PRS') || label.includes('PRESS')) key = 'pressing';
                    else if (label.includes('CT') || label.includes('CUT')) key = 'cutting';
                    else if (label.includes('MIX')) key = 'mixing';
                    else if (label.includes('FIN')) key = 'finishing';

                    if (key && newData[key]) {
                        newData[key] = {
                            total: item.target, 
                            completed: item.actual,
                            pending: Math.max(0, item.target - item.actual)
                        };
                    }
                });
                dashboardData = newData;
            }
        } catch (error) {
            console.error('Error fetching dashboard data:', error);
        } finally {
            isLoading = false;
            initializeCharts();
        }
    }

    function initializeCharts() {
        setTimeout(() => {
            divisions.forEach((division) => {
                const data = dashboardData[division.id];
                
                if (charts[division.id]?.total) charts[division.id].total?.destroy();
                if (charts[division.id]?.status) charts[division.id].status?.destroy();

                // Target standard (bisa disesuaikan per divisi jika perlu)
                const productionTarget = 35; 
                const notGoodTarget = 5; 
                
                // --- CHART 1: TOTAL ---
                const totalCanvasId = `chart-total-${division.id}`;
                const totalCanvas = document.getElementById(totalCanvasId) as HTMLCanvasElement;
                if (totalCanvas) {
                    const newChart = new Chart(totalCanvas, {
                        type: 'bar',
                        data: {
                            labels: ['Total Produksi'],
                            datasets: [
                                {
                                    label: 'Produk',
                                    data: [data.completed || 0, data.pending || 0],
                                    backgroundColor: '#3b82f6',
                                    borderRadius: 4,
                                    borderSkipped: false
                                }
                            ]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: true,
                            plugins: {
                                legend: { display: false },
                                annotation: {
                                    annotations: {
                                        targetLine: {
                                            type: 'line',
                                            yMin: productionTarget,
                                            yMax: productionTarget,
                                            borderColor: '#10b981',
                                            borderWidth: 2,
                                            borderDash: [5, 5],
                                            label: {
                                                content: `Target: ${productionTarget}`,
                                                display: true, // FIXED: Ganti enabled -> display
                                                position: 'end',
                                                backgroundColor: '#10b981',
                                                color: 'white',
                                                font: { size: 10 }
                                            }
                                        }
                                    }
                                }
                            },
                            scales: {
                                y: {
                                    beginAtZero: true,
                                    suggestedMax: productionTarget + 10, // Agar grafik tidak mentok atas
                                    ticks: { 
                                        font: { size: 10 },
                                        stepSize: 1, // FIXED: Paksa bilangan bulat
                                        precision: 0 
                                    },
                                    grid: { color: '#f0f0f0' }
                                },
                                x: { display: false }
                            }
                        }
                    });
                    if (!charts[division.id]) charts[division.id] = {};
                    charts[division.id].total = newChart;
                }

                // --- CHART 2: STATUS / NG ---
                const notGoodCanvasId = `chart-notgood-${division.id}`;
                const notGoodCanvas = document.getElementById(notGoodCanvasId) as HTMLCanvasElement;
                if (notGoodCanvas) {
                    const notGoodCount = Math.floor((data.pending || 0) * 0.4); 
                    const goodCount = (data.pending || 0) - notGoodCount;
                    const newChart = new Chart(notGoodCanvas, {
                        type: 'bar',
                        data: {
                            labels: ['NG'],
                            datasets: [{
                                label: 'Status',
                                data: [notGoodCount, goodCount],
                                backgroundColor: '#ef4444',
                                borderRadius: 4
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: true,
                            plugins: {
                                legend: { display: false },
                                annotation: {
                                    annotations: {
                                        targetLine: {
                                            type: 'line',
                                            yMin: notGoodTarget,
                                            yMax: notGoodTarget,
                                            borderColor: '#f59e0b',
                                            borderWidth: 2,
                                            borderDash: [5, 5],
                                            label: {
                                                content: `Max: ${notGoodTarget}`,
                                                display: true, // FIXED
                                                position: 'end',
                                                backgroundColor: '#f59e0b',
                                                color: 'white',
                                                font: { size: 10 }
                                            }
                                        }
                                    }
                                }
                            },
                            scales: {
                                y: {
                                    beginAtZero: true,
                                    suggestedMax: notGoodTarget + 5,
                                    ticks: { 
                                        font: { size: 10 },
                                        stepSize: 1, 
                                        precision: 0 
                                    },
                                    grid: { color: '#f0f0f0' }
                                },
                                x: { display: false }
                            }
                        }
                    });
                    if (!charts[division.id]) charts[division.id] = {};
                    charts[division.id].status = newChart;
                }
            });
        }, 100);
    }

    function getStatusColor(divisionId: string) {
        if (divisionId === 'cutting') return { badge: 'bg-red-500', label: 'Problem' };
        if (divisionId === 'mixing') return { badge: 'bg-green-500', label: 'Lancar' };
        if (divisionId === 'pressing') return { badge: 'bg-yellow-500', label: 'Sedang Berjalan' };
        if (divisionId === 'finishing') return { badge: 'bg-green-500', label: 'Lancar' };
        return { badge: 'bg-slate-500', label: 'Unknown' };
    }

    function handleDetailClick(divisionId: string) {
        // Redirect ke halaman detail proses dengan tanggal yang dipilih
        goto(`/manager/prs-ldr?process=${divisionId.toUpperCase()}&tanggal=${selectedDate}`);
    }

    onMount(() => {
        if ($auth.token) fetchDashboardData();
    });
</script>

<div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-slate-100 p-8">
    <div class="max-w-7xl mx-auto">
        <div class="mb-8 flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
            <div>
                <h1 class="text-4xl font-bold text-slate-900 mb-2">Dashboard Manager</h1>
                {#if isOffline}
                    <div class="bg-amber-100 border-l-4 border-amber-500 text-amber-700 px-3 py-2 rounded text-sm flex items-center gap-2">
                        <i class="fa-solid fa-triangle-exclamation"></i>
                        <span>Mode Offline: Database produksi tidak terhubung.</span>
                    </div>
                {:else}
                    <p class="text-slate-600 text-lg">Pantau performa divisi produksi secara real-time</p>
                {/if}
            </div>

            <div class="bg-white p-2 rounded-lg shadow-sm border border-slate-200">
                <label class="block text-xs font-bold text-slate-500 px-1 mb-1">Pilih Tanggal</label>
                <input 
                    type="date" 
                    bind:value={selectedDate} 
                    on:change={fetchDashboardData}
                    class="border-none outline-none text-slate-700 font-bold bg-transparent cursor-pointer focus:ring-0"
                />
            </div>
        </div>

        {#if isLoading}
            <div class="flex justify-center items-center min-h-64">
                <div class="text-center">
                    <div class="inline-block animate-spin">
                        <i class="fa-solid fa-spinner text-indigo-600 text-4xl"></i>
                    </div>
                    <p class="text-slate-600 mt-4">Memuat data...</p>
                </div>
            </div>
        {:else}
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-10">
                {#each divisions as division (division.id)}
                    {@const status = getStatusColor(division.id)}
                    {@const data = dashboardData[division.id]}
                    <div class="group relative bg-white rounded-2xl shadow-lg hover:shadow-2xl transition-all duration-300 overflow-hidden border border-slate-100 hover:border-slate-200">
                        <div class="absolute top-0 left-0 right-0 h-1.5 bg-gradient-to-r {division.color}"></div>
                        <div class="p-8 relative z-10">
                            <div class="flex items-center gap-4 mb-8">
                                <div class="{division.bgColor} p-4 rounded-xl group-hover:scale-110 transition-transform shadow-md">
                                    <i class="{division.icon} {division.textColor} text-2xl"></i>
                                </div>
                                <div class="flex-1">
                                    <h3 class="text-xl font-bold text-slate-800">{division.name}</h3>
                                    <p class="text-sm text-slate-500 mt-1">{division.description}</p>
                                </div>
                            </div>
                            <div class="mb-8 flex items-center gap-3 px-4 py-3 bg-slate-50 rounded-lg border border-slate-200">
                                <span class="{status.badge} w-3 h-3 rounded-full animate-pulse"></span>
                                <span class="text-slate-700 text-sm font-bold flex-1">{status.label}</span>
                                <span class="text-slate-600 text-sm font-semibold px-3 py-1 bg-white rounded-md">Total: <span class="text-slate-900 font-bold">{data.total}</span></span>
                            </div>
                            <div class="grid grid-cols-2 gap-6">
                                <div class="flex flex-col items-center p-4 bg-blue-50 rounded-lg border border-blue-200">
                                    <p class="text-xs text-blue-700 font-bold mb-3">Total Output</p>
                                    <canvas id="chart-total-{division.id}" style="max-height: 120px;"></canvas>
                                </div>
                                <div class="flex flex-col items-center p-4 bg-red-50 rounded-lg border border-red-200">
                                    <p class="text-xs text-red-700 font-bold mb-3">Total NG</p>
                                    <canvas id="chart-notgood-{division.id}" style="max-height: 120px;"></canvas>
                                </div>
                            </div>
                        </div>
                        <div class="px-8 py-4 bg-slate-50 border-t border-slate-100">
                            <button on:click={() => handleDetailClick(division.id)} class="w-full py-2 px-4 rounded-lg text-sm font-semibold text-white bg-slate-700 hover:bg-slate-800 transition-all duration-300">
                                Lihat Detail
                            </button>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
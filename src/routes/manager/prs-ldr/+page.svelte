<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { goto } from '$app/navigation';
    import Chart from 'chart.js/auto';
    import annotationPlugin from 'chartjs-plugin-annotation';
    import { auth } from '$lib/stores/auth';
    import { page } from '$app/stores';

    Chart.register(annotationPlugin);

    let machineData: any[] = [];
    let isLoading = true; 
    let charts: Record<string, any> = {};
    const API_URL = 'http://localhost:8080';
    let processName = 'PRS'; 
    let selectedDate = new Date().toISOString().split('T')[0];

    $: processName = $page.url.searchParams.get('process') || 'PRESSING';
    $: urlDate = $page.url.searchParams.get('tanggal');

    // Jika ada tanggal di URL, gunakan itu
    $: if (urlDate) selectedDate = urlDate;

    function goBack() {
        if ($auth.user?.role === 'MANAGER') {
            goto('/manager');
        }
    }

    function selectMachine(machine: any) {
        goto(`/manager/mesin-detail?no_mc=${machine.id}&tanggal=${selectedDate}`);
    }

    function getMachineStatusColor(machine: any) {
        if (machine.isProblem) {
            return {
                bg: 'bg-red-500 hover:bg-red-600',
                badge: 'bg-red-700',
                text: 'text-white',
                indicator: '🔴'
            };
        }
        return {
            bg: 'bg-green-500 hover:bg-green-600',
            badge: 'bg-green-700',
            text: 'text-white',
            indicator: '🟢'
        };
    }

    async function fetchMachineStatus() {
        isLoading = true;
        machineData = []; // Reset data
        try {
            console.log(`Fetching Process: ${processName} Date: ${selectedDate}`);
            
            const res = await fetch(`${API_URL}/api/chart/process?tanggal=${selectedDate}&proses=${processName}`, {
                headers: { Authorization: `Bearer ${$auth.token}` }
            });
            
            if (res.ok) {
                const data = await res.json();
                console.log("Data diterima:", data); // DEBUG: Cek di Console Browser

                if (Array.isArray(data)) {
                    machineData = data.map((item: any) => {
                        // FIX: Handle omitempty / null / undefined -> 0
                        const target = Math.round(item.target || 0);
                        const actual = Math.round(item.actual || 0);
                        const ng = Math.round(item.actual_ng || 0);

                        const achievement = target > 0 ? (actual / target) * 100 : 0;
                        const ngRate = (actual + ng) > 0 ? (ng / (actual + ng)) * 100 : 0;

                        return {
                            id: item.label,
                            name: `Mesin ${item.label}`,
                            target: target,
                            completed: actual,
                            notGood: ng,
                            isProblem: achievement < 80 || ngRate > 5 
                        };
                    });
                }
            }
        } catch (e) {
            console.error("Gagal ambil data mesin:", e);
        } finally {
            isLoading = false;
            // Beri sedikit delay agar DOM render dulu sebelum Chart.js init
            setTimeout(initializeCharts, 100);
        }
    }

    function initializeCharts() {
        // Hapus chart lama
        Object.keys(charts).forEach(machineId => {
            if (charts[machineId]?.total) charts[machineId].total.destroy();
            if (charts[machineId]?.notGood) charts[machineId].notGood.destroy();
        });
        charts = {};

        machineData.forEach((machine) => {
            const machineTarget = machine.target || 10; // Default skala jika 0
            const maxScale = Math.max(machineTarget, machine.completed) + 10;

            // --- Chart Total Produksi ---
            const totalCanvasId = `chart-total-${machine.id}`;
            const totalCanvas = document.getElementById(totalCanvasId) as HTMLCanvasElement;
            
            if (totalCanvas) {
                const newChart = new Chart(totalCanvas, {
                    type: 'bar',
                    data: {
                        labels: ['Produksi'],
                        datasets: [{
                            label: 'Produksi',
                            data: [machine.completed],
                            backgroundColor: '#3b82f6',
                            borderColor: '#1e40af',
                            borderWidth: 1,
                            borderRadius: 6
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: { display: false },
                            annotation: {
                                annotations: {
                                    targetLine: {
                                        type: 'line',
                                        yMin: machine.target,
                                        yMax: machine.target,
                                        borderColor: '#10b981',
                                        borderWidth: 2,
                                        borderDash: [5, 5],
                                        label: {
                                            content: `Target: ${machine.target}`,
                                            display: true,
                                            position: 'end',
                                            backgroundColor: '#10b981',
                                            color: 'white',
                                            font: { size: 9, weight: 'bold' }
                                        }
                                    }
                                }
                            }
                        },
                        scales: {
                            y: {
                                beginAtZero: true,
                                max: maxScale, 
                                ticks: { font: { size: 9 }, stepSize: 1 },
                                display: true
                            },
                            x: { display: false }
                        }
                    }
                });
                if (!charts[machine.id]) charts[machine.id] = {};
                charts[machine.id].total = newChart;
            }

            // --- Chart Not Good ---
            const notGoodCanvasId = `chart-notgood-${machine.id}`;
            const notGoodCanvas = document.getElementById(notGoodCanvasId) as HTMLCanvasElement;
            
            if (notGoodCanvas) {
                const maxNG = Math.ceil(machineTarget * 0.05) || 5;
                const newChart = new Chart(notGoodCanvas, {
                    type: 'bar',
                    data: {
                        labels: ['NG'],
                        datasets: [{
                            label: 'NG',
                            data: [machine.notGood],
                            backgroundColor: '#ef4444',
                            borderColor: '#b91c1c',
                            borderWidth: 1,
                            borderRadius: 6
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: { legend: { display: false } },
                        scales: {
                            y: {
                                beginAtZero: true,
                                suggestedMax: Math.max(machine.notGood, maxNG) + 5, 
                                ticks: { font: { size: 9 }, stepSize: 1 },
                                display: true
                            },
                            x: { display: false }
                        }
                    }
                });
                if (!charts[machine.id]) charts[machine.id] = {};
                charts[machine.id].notGood = newChart;
            }
        });
    }

    $: pressMachines = machineData.filter(m => !['20A', '21A', '22A', '19A', '18A', '17A'].includes(m.id));
    $: injectMachines = machineData.filter(m => ['20A', '21A', '22A', '19A', '18A', '17A'].includes(m.id));

    onMount(() => {
        if ($auth.token) fetchMachineStatus();
        window.addEventListener('resize', handleResize);
    });

    function handleResize() {
        setTimeout(initializeCharts, 300);
    }

    onDestroy(() => {
        Object.keys(charts).forEach(machineId => {
            if (charts[machineId]?.total) charts[machineId].total.destroy();
            if (charts[machineId]?.notGood) charts[machineId].notGood.destroy();
        });
        if (typeof window !== 'undefined') window.removeEventListener('resize', handleResize);
    });
</script>

<div class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 p-4 md:p-6">
    <div class="max-w-7xl mx-auto">
        <div class="mb-6 flex flex-col sm:flex-row items-center justify-between gap-4">
            <div>
                {#if $auth.user?.role === 'MANAGER'}
                    <button on:click={goBack} class="mb-3 inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-white text-slate-700 hover:bg-slate-50 transition-all duration-200 border border-slate-200 shadow-sm hover:shadow">
                        <i class="fa-solid fa-arrow-left"></i>
                        Kembali
                    </button>
                {/if}
                <h1 class="text-2xl md:text-3xl font-bold text-slate-800">Dashboard: {processName}</h1>
            </div>

            <div class="bg-white p-2 rounded-lg shadow-sm border border-slate-200 flex items-center gap-2">
                <span class="text-xs font-bold text-slate-500 uppercase px-2">Tanggal:</span>
                <input 
                    type="date" 
                    bind:value={selectedDate} 
                    on:change={fetchMachineStatus}
                    class="border-none outline-none text-slate-800 font-bold bg-transparent cursor-pointer focus:ring-0 text-sm"
                />
            </div>
        </div>

        {#if isLoading}
            <div class="flex justify-center items-center min-h-64 bg-white rounded-2xl shadow-lg border border-slate-200">
                <div class="text-center">
                    <i class="fa-solid fa-spinner animate-spin text-indigo-600 text-4xl"></i>
                    <p class="text-slate-600 mt-4">Memuat data...</p>
                </div>
            </div>
        {:else if machineData.length === 0}
            <div class="flex flex-col justify-center items-center min-h-64 bg-white rounded-2xl shadow-lg border border-slate-200 p-8">
                <div class="bg-slate-100 p-4 rounded-full mb-4">
                    <i class="fa-solid fa-database text-slate-400 text-4xl"></i>
                </div>
                <h3 class="text-xl font-bold text-slate-700">Tidak Ada Data</h3>
                <p class="text-slate-500 text-center mt-2 max-w-md">
                    Tidak ditemukan data produksi untuk proses <b>{processName}</b> pada tanggal <b>{selectedDate}</b>.
                </p>
                <p class="text-xs text-slate-400 mt-4">Pastikan nama proses di database sesuai (Misal: PRESSING vs PRS)</p>
            </div>
        {:else}
            <div class="bg-white rounded-2xl shadow-lg p-4 md:p-6 border border-slate-200">
                <h2 class="text-lg font-bold text-slate-800 mb-6">Status Mesin</h2>
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div>
                        <h3 class="text-xs font-bold text-slate-600 mb-4 uppercase tracking-wider flex items-center gap-2">
                            <span class="w-2 h-2 bg-purple-500 rounded-full"></span> Mesin Group A
                        </h3>
                        <div class="space-y-4">
                            {#each injectMachines as machine}
                                {@const status = getMachineStatusColor(machine)}
                                <div class="bg-gradient-to-br from-slate-50 to-slate-100 rounded-xl border border-slate-200 overflow-hidden hover:shadow-xl transition-all duration-300 transform hover:-translate-y-1">
                                    <button on:click={() => selectMachine(machine)} class="w-full flex items-center gap-3 p-4 hover:bg-slate-50/50 transition-colors">
                                        <div class="relative">
                                            <div class="w-10 h-10 flex items-center justify-center rounded-full {status.badge} text-white font-bold text-sm flex-shrink-0 shadow">{machine.id}</div>
                                            <div class="absolute -top-1 -right-1">{status.indicator}</div>
                                        </div>
                                        <div class="flex-1 text-left min-w-0">
                                            <div class="font-semibold text-slate-800 truncate">{machine.name}</div>
                                            <div class="text-xs text-slate-600 flex items-center gap-1">
                                                <span class="w-2 h-2 bg-blue-500 rounded-full"></span>
                                                {machine.completed} / {machine.target}
                                            </div>
                                        </div>
                                        <i class="fa-solid fa-arrow-right text-slate-400"></i>
                                    </button>
                                    <div class="px-4 pb-4 pt-2 border-t border-slate-200/50 space-y-4">
                                        <div class="h-32"><canvas id="chart-total-{machine.id}"></canvas></div>
                                        <div class="h-32"><canvas id="chart-notgood-{machine.id}"></canvas></div>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>
                    <div>
                        <h3 class="text-xs font-bold text-slate-600 mb-4 uppercase tracking-wider flex items-center gap-2">
                            <span class="w-2 h-2 bg-blue-500 rounded-full"></span> Mesin Group B
                        </h3>
                        <div class="space-y-4">
                            {#each pressMachines as machine}
                                {@const status = getMachineStatusColor(machine)}
                                <div class="bg-gradient-to-br from-slate-50 to-slate-100 rounded-xl border border-slate-200 overflow-hidden hover:shadow-xl transition-all duration-300 transform hover:-translate-y-1">
                                    <button on:click={() => selectMachine(machine)} class="w-full flex items-center gap-3 p-4 hover:bg-slate-50/50 transition-colors">
                                        <div class="relative">
                                            <div class="w-10 h-10 flex items-center justify-center rounded-full {status.badge} text-white font-bold text-sm flex-shrink-0 shadow">{machine.id}</div>
                                            <div class="absolute -top-1 -right-1">{status.indicator}</div>
                                        </div>
                                        <div class="flex-1 text-left min-w-0">
                                            <div class="font-semibold text-slate-800 truncate">{machine.name}</div>
                                            <div class="text-xs text-slate-600 flex items-center gap-1">
                                                <span class="w-2 h-2 bg-blue-500 rounded-full"></span>
                                                {machine.completed} / {machine.target}
                                            </div>
                                        </div>
                                        <i class="fa-solid fa-arrow-right text-slate-400"></i>
                                    </button>
                                    <div class="px-4 pb-4 pt-2 border-t border-slate-200/50 space-y-4">
                                        <div class="h-32"><canvas id="chart-total-{machine.id}"></canvas></div>
                                        <div class="h-32"><canvas id="chart-notgood-{machine.id}"></canvas></div>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>
                </div>
            </div>
        {/if}
    </div>
</div>

<style>
    :global(body) { background-color: #f8fafc; font-family: 'Inter', sans-serif; }
    canvas { width: 100% !important; height: 100% !important; }
</style>
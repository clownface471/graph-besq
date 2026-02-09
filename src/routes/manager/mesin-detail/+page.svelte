<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import Chart from 'chart.js/auto';
    import { auth } from '$lib/stores/auth';
    import { goto } from '$app/navigation';
    import annotationPlugin from 'chartjs-plugin-annotation';
    Chart.register(annotationPlugin);

    let canvasTotal: HTMLCanvasElement;
    let canvasOK: HTMLCanvasElement;
    let canvasNG: HTMLCanvasElement;
    let chartTotal: Chart;
    let chartOK: Chart;
    let chartNG: Chart;
    
    // Ambil param dari URL
    $: noMC = $page.url.searchParams.get('no_mc') || '';
    $: selectedDate = $page.url.searchParams.get('tanggal') || new Date().toISOString().split('T')[0];
    
    let filters = {
        tanggal: selectedDate,
        mesin: noMC
    };

    $: {
        filters.tanggal = selectedDate;
        filters.mesin = noMC;
    }

    let itemsProduced = "";
    let isLoading = false;
    // Array untuk menampung list mesin dari DB
    let machineList: string[] = []; 
    
    const API_URL = 'http://localhost:8080';

    function goBack() {
        history.back();
    }

    function updateFilters() {
        goto(`/manager/mesin-detail?no_mc=${filters.mesin}&tanggal=${filters.tanggal}`, { replaceState: true });
    }

    // --- FUNGSI BARU: Fetch List Mesin Aktual ---
    async function fetchMachineList() {
        try {
            const res = await fetch(`${API_URL}/api/chart/machines`, {
                headers: { Authorization: `Bearer ${$auth.token}` }
            });
            if (res.ok) {
                machineList = await res.json();
                
                // Jika mesin di URL tidak ada di list database, peringatkan atau biarkan saja
                // (User mungkin mengakses link lama)
                if (machineList.length > 0 && !noMC) {
                    // Jika tidak ada mesin terpilih, pilih yang pertama
                    filters.mesin = machineList[0];
                    updateFilters();
                }
            } else {
                console.warn("Gagal load list mesin dari DB, menggunakan fallback.");
                generateFallbackMachines();
            }
        } catch (e) {
            console.error("Error fetching machines:", e);
            generateFallbackMachines();
        }
    }

    function generateFallbackMachines() {
        // Fallback jika offline: Generate 01A - 25A
        machineList = Array.from({length: 25}, (_, i) => `${(i + 1).toString().padStart(2, '0')}A`);
    }

    async function loadChartData() {
        if (!filters.mesin) return; // Jangan load jika belum ada mesin terpilih
        isLoading = true;
        try {
            const res = await fetch(`${API_URL}/api/chart/machine?tanggal=${filters.tanggal}&no_mc=${filters.mesin}`, {
                 headers: { Authorization: `Bearer ${$auth.token}` }
            });
            
            // Handle jika offline/maintenance
            if (res.status === 503) {
                // Bisa handle state offline disini jika mau
                itemsProduced = "Offline Mode";
                renderCharts([]); // Render kosong
                return;
            }

            const data = await res.json();
            
            const validItem = data.find((d: any) => d.extra_info && d.extra_info !== '- (-)' && d.extra_info !== '-');
            itemsProduced = validItem ? validItem.extra_info : "-";

            renderCharts(data);
        } catch (error) {
            console.error("Error fetching machine detail:", error);
        } finally {
            isLoading = false;
        }
    }

    function renderCharts(data: any[]) {
        const labels = data.map((d: any) => d.label);
        const totalVals = data.map((d: any) => d.actual); 
        const okVals = data.map((d: any) => d.actual_ok);
        const ngVals = data.map((d: any) => d.actual_ng);

        const targetTotal = 30; 
        const targetNG = 5;

        // --- CHART 1: Total Output ---
        if (chartTotal) chartTotal.destroy();
        chartTotal = new Chart(canvasTotal, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'Output',
                        data: totalVals,
                        backgroundColor: '#4f46e5',
                        borderRadius: 4
                    },
                    {
                        label: 'Target',
                        type: 'line',
                        data: Array(labels.length).fill(targetTotal),
                        borderColor: '#10b981',
                        borderWidth: 2,
                        borderDash: [5, 5],
                        pointRadius: 0
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { 
                        beginAtZero: true, 
                        title: { display: true, text: 'Pcs' },
                        ticks: { stepSize: 1, precision: 0 },
                        suggestedMax: targetTotal + 10
                    },
                    x: { display: false }
                },
                plugins: {
                    legend: { display: false },
                    title: { display: true, text: 'Total Output' }
                }
            }
        });

        // --- CHART 2: OK Output ---
        if (chartOK) chartOK.destroy();
        chartOK = new Chart(canvasOK, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'OK',
                        data: okVals,
                        backgroundColor: '#10b981',
                        borderRadius: 4
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { 
                        beginAtZero: true, 
                        title: { display: true, text: 'Pcs' },
                        ticks: { stepSize: 1, precision: 0 },
                        suggestedMax: targetTotal + 10
                    },
                    x: { display: false }
                },
                plugins: {
                    legend: { display: false },
                    title: { display: true, text: 'Barang OK' }
                }
            }
        });

        // --- CHART 3: NG ---
        if (chartNG) chartNG.destroy();
        chartNG = new Chart(canvasNG, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'NG',
                        data: ngVals,
                        backgroundColor: '#e11d48',
                        borderRadius: 4
                    },
                    {
                        label: 'Max Limit',
                        type: 'line',
                        data: Array(labels.length).fill(targetNG),
                        borderColor: '#f59e0b',
                        borderWidth: 2,
                        borderDash: [5, 5],
                        pointRadius: 0
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: { 
                        beginAtZero: true, 
                        title: { display: true, text: 'Pcs' },
                        ticks: { stepSize: 1, precision: 0 },
                        suggestedMax: targetNG + 5
                    },
                    x: { display: false }
                },
                plugins: {
                    legend: { display: false },
                    title: { display: true, text: 'Barang NG' }
                }
            }
        });
    }

    // Trigger loadChartData setiap kali parameter URL berubah
    $: if (noMC && selectedDate) {
        loadChartData();
    }

    onMount(() => {
        fetchMachineList(); // Ambil list mesin dulu
        loadChartData();    // Lalu load data chart
    });
</script>

<div class="p-6 max-w-7xl mx-auto space-y-6">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
            <button on:click={goBack} class="text-sm text-indigo-600 hover:underline flex items-center gap-1 mb-1">
                <i class="fa-solid fa-arrow-left"></i> Kembali
            </button>
            <h1 class="text-2xl font-bold text-slate-800">Detail Monitoring Mesin</h1>
        </div>
        
        <div class="bg-white p-3 rounded-xl shadow-sm border border-slate-200 flex flex-wrap gap-3 items-center">
            
            <div class="flex flex-col">
                <label class="text-[10px] font-bold text-slate-400 uppercase">Tanggal</label>
                <input 
                    type="date" 
                    bind:value={filters.tanggal} 
                    on:change={updateFilters}
                    class="text-sm font-bold text-slate-700 bg-transparent outline-none border-b border-transparent focus:border-indigo-500 transition-colors"
                >
            </div>

            <div class="h-8 w-px bg-slate-200 mx-1"></div>

            <div class="flex flex-col min-w-[100px]">
                <label class="text-[10px] font-bold text-slate-400 uppercase">Pilih Mesin</label>
                <div class="relative">
                    <select 
                        bind:value={filters.mesin} 
                        on:change={updateFilters}
                        class="w-full text-sm font-bold text-slate-700 bg-transparent outline-none appearance-none cursor-pointer pr-6"
                    >
                        {#if machineList.length === 0}
                            <option value="">Loading...</option>
                        {:else}
                            {#each machineList as mc}
                                <option value={mc}>{mc}</option>
                            {/each}
                        {/if}
                    </select>
                    <i class="fa-solid fa-chevron-down absolute right-0 top-1 text-xs text-slate-400 pointer-events-none"></i>
                </div>
            </div>
        </div>
    </div>
    
    <div class="bg-white p-4 rounded-xl shadow-sm border border-blue-100 flex items-center gap-4">
        <div class="p-3 bg-blue-50 rounded-lg text-blue-600">
            <i class="fa-solid fa-industry text-xl"></i>
        </div>
        <div>
            <span class="text-xs text-slate-500 font-bold uppercase">Sedang Diproduksi</span>
            <p class="text-lg font-bold text-slate-800">{itemsProduced}</p>
        </div>
    </div>

    {#if isLoading}
        <div class="flex justify-center items-center h-64 bg-white rounded-lg shadow">
            <div class="text-center">
                <i class="fa-solid fa-spinner animate-spin text-4xl text-blue-500"></i>
                <p class="mt-4 text-gray-500 font-medium">Memuat data grafik...</p>
            </div>
        </div>
    {:else}
        <div class="space-y-6">
            <div class="bg-white p-4 rounded-xl shadow border border-slate-200">
                <div class="h-[350px]">
                    <canvas bind:this={canvasTotal}></canvas>
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div class="bg-white p-4 rounded-xl shadow border border-slate-200">
                    <div class="h-[250px]">
                        <canvas bind:this={canvasOK}></canvas>
                    </div>
                </div>

                <div class="bg-white p-4 rounded-xl shadow border border-slate-200">
                    <div class="h-[250px]">
                        <canvas bind:this={canvasNG}></canvas>
                    </div>
                </div>
            </div>
        </div>
    {/if}
</div>

<style>
    :global(body) { background-color: #f8fafc; font-family: 'Inter', sans-serif; }
    canvas { width: 100% !important; height: 100% !important; }
</style>
<script lang="ts">
    // @ts-nocheck
    import { auth, login, logout, type User } from '$lib/stores/auth';
    import Swal from 'sweetalert2';
    import { goto } from '$app/navigation';

    // --- STATE (Svelte 5 Runes) ---
    let username = $state('');
    let password = $state('');
    let isLoading = $state(false);
    let errorMessage = $state('');

    let stats = $state({
        totalEmployees: 0,
        totalOutput: 0,
        rejectRate: '0%',
        activeShift: 'N/A',
        cuttingOutput: 0,
        pressingOutput: 0,
        finishingOutput: 0, 
        activities: [] as { user: string; action: string; time: string; status: string }[]
    });

    const API_URL = 'http://localhost:8080';

    // --- LIFECYCLE & EFFECTS ---
    $effect(() => {
        if ($auth.isLoggedIn) {
            fetchStats();
        }
    });

    // --- ASYNC FUNCTIONS ---
    async function fetchStats() {
        // Backend saat ini belum punya endpoint stats, kita mock dulu agar tidak error
        // Nanti bisa dibuatkan endpointnya jika perlu
        return; 
    }

    async function handleLogin(e: Event) {
        e.preventDefault();
        isLoading = true;
        errorMessage = '';

        try {
            const response = await fetch(`${API_URL}/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });

            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Login failed');
            }

            // Backend kita langsung mengembalikan role & department, jadi tidak perlu fetch profile lagi
            // { token: "...", role: "MANAGER", department: "..." }
            
            const userData: User = {
                username: username,
                role: data.role,
                department: data.department
            };

            auth.login(data.token, userData);

            // --- LOGIKA REDIRECT (Sesuai Struktur Graph) ---
            const role = data.role;
            const dept = data.department;

            if (role === 'MANAGER') {
                goto('/manager');
            } else if (role === 'LEADER') {
                goto(`/manager/prs-ldr?process=${dept || 'PRESSING'}`);
            } else if (role === 'OPERATOR') {
                goto('/manager/mesin-detail?no_mc=01A');
            } else {
                goto('/'); 
            }
        } catch (error) {
            errorMessage = error instanceof Error ? error.message : 'An error occurred during login';
            console.error('Login error:', error);
        } finally {
            isLoading = false;
        }
    }

    async function handleLogout() {
        const result = await Swal.fire({
            title: 'Yakin ingin keluar?',
            text: 'Anda harus login ulang untuk mengakses aplikasi.',
            icon: 'warning',
            showCancelButton: true,
            confirmButtonText: 'Ya, Keluar',
            cancelButtonText: 'Batal'
        });

        if (result.isConfirmed) {
            auth.logout();
        }
    }

    // --- STATIC DATA (HANYA PRESSING & CUTTING) ---
    // Data dummy untuk tampilan dashboard "Home" (jika user memaksa masuk sini)
    let deptSummary = [
        {
            id: 'pressing',
            name: 'Pressing',
            icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
            totalStaff: 12,
            activeNow: 10,
            avgEfficiency: 88,
            target: 2000,
            color: 'amber'
        },
        {
            id: 'cutting',
            name: 'Cutting',
            icon: 'M14.121 14.121L19 19m-7-7l7-7m-7 7l-2.879 2.879M12 12L9.121 9.121m0 5.758a3 3 0 10-4.243 4.243 3 3 0 004.243-4.243zm0-5.758a3 3 0 10-4.243-4.243 3 3 0 004.243 4.243z',
            totalStaff: 8,
            activeNow: 8,
            avgEfficiency: 95,
            target: 2500,
            color: 'blue'
        }
    ];

    // --- UI HELPER FUNCTIONS ---
    function getStatusColor(status: string) {
        if (status === 'Completed') return 'bg-emerald-100 text-emerald-700 ring-emerald-600/20';
        if (status === 'Pending') return 'bg-slate-100 text-slate-700 ring-slate-600/20';
        return 'bg-amber-100 text-amber-700 ring-amber-600/20';
    }

    function getDeptOutput(deptId: string) {
        if (deptId === 'pressing') return stats.pressingOutput;
        if (deptId === 'cutting') return stats.cuttingOutput;
        return 0;
    }

    function getCardColor(color: string) {
        const map = {
            amber: 'bg-amber-50 border-amber-100 hover:border-amber-300 text-amber-900',
            blue: 'bg-blue-50 border-blue-100 hover:border-blue-300 text-blue-900',
            emerald: 'bg-emerald-50 border-emerald-100 hover:border-emerald-300 text-emerald-900'
        };
        return map[color] || '';
    }

    function getIconBg(color: string) {
        const map = {
            amber: 'bg-amber-100 text-amber-600',
            blue: 'bg-blue-100 text-blue-600',
            emerald: 'bg-emerald-100 text-emerald-600'
        };
        return map[color] || '';
    }

    function getBarColor(color: string) {
        const map = {
            amber: 'bg-amber-500',
            blue: 'bg-blue-500',
            emerald: 'bg-emerald-500'
        };
        return map[color] || '';
    }
</script>

<div class="font-sans text-slate-800 bg-gray-100 min-h-screen">
    {#if !$auth.isLoggedIn}
        <div class="min-h-screen flex items-center justify-center p-4">
            <div
                class="border w-full max-w-md overflow-hidden shadow-xl bg-white rounded-2xl sm:rounded-3xl p-4 sm:p-6 md:p-8 transition-all duration-300 hover:shadow-2xl"
            >
                <div class="mb-6 sm:mb-8">
                    <div class="flex justify-center mb-2">
                        <span
                            class="bg-blue-100 text-[#0065F8] text-xs sm:text-[10px] font-bold px-2.5 py-1 rounded-full uppercase tracking-widest"
                        >
                            Official Portal
                        </span>
                    </div>
                    <h1 class="font-bold text-xl sm:text-2xl text-center text-gray-800 font-[Poppins] tracking-tight">
                        Besq User Login
                    </h1>
                    <div class="text-[#0065F8] text-5xl sm:text-6xl md:text-7xl text-center mt-4 sm:mt-6">
                        <i class="fas fa-circle-user text-blue-600"></i>
                    </div>
                </div>

                <form onsubmit={handleLogin} class="space-y-4 sm:space-y-5">
                    {#if errorMessage}
                        <div class="bg-red-50 border border-red-200 text-red-700 px-3 sm:px-4 py-2.5 sm:py-3 rounded-lg text-sm">
                            {errorMessage}
                        </div>
                    {/if}

                    <div>
                        <label
                            for="username"
                            class="text-xs font-bold text-gray-400 uppercase tracking-widest ml-1 block mb-1">Username</label
                        >
                        <input
                            id="username"
                            type="text"
                            placeholder="Masukkan Username"
                            bind:value={username}
                            required
                            disabled={isLoading}
                            class="w-full border border-gray-200 bg-gray-50 rounded-lg sm:rounded-xl py-2.5 sm:py-3 px-3 sm:px-4 outline-none focus:bg-white focus:border-[#0065F8] focus:ring-2 sm:focus:ring-4 focus:ring-blue-100 transition-all disabled:opacity-50 text-sm sm:text-base"
                        />
                    </div>

                    <div>
                        <label
                            for="password"
                            class="text-xs font-bold text-gray-400 uppercase tracking-widest ml-1 block mb-1"
                            >Password</label
                        >
                        <input
                            id="password"
                            type="password"
                            placeholder="••••••••"
                            bind:value={password}
                            required
                            disabled={isLoading}
                            class="w-full border border-gray-200 bg-gray-50 rounded-lg sm:rounded-xl py-2.5 sm:py-3 px-3 sm:px-4 outline-none focus:bg-white focus:border-[#0065F8] focus:ring-2 sm:focus:ring-4 focus:ring-blue-100 transition-all disabled:opacity-50 text-sm sm:text-base"
                        />
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading}
                        class="bg-[#0065F8] shadow-lg shadow-blue-300/50 transition-all duration-300 hover:bg-[#004bbd] hover:shadow-blue-400/50 hover:-translate-y-0.5 text-white font-bold py-3 sm:py-3.5 px-4 rounded-xl w-full mt-6 sm:mt-8 flex justify-center items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed text-sm sm:text-base"
                    >
                        {#if isLoading}
                            <span
                                class="animate-spin border-2 border-white border-t-transparent rounded-full h-4 w-4 sm:h-5 sm:w-5"
                            ></span>
                            <span>Logging in...</span>
                        {:else}
                            <span>Login</span>
                        {/if}
                    </button>
                </form>
            </div>
        </div>
    {:else}
        <div class="min-h-screen bg-slate-50 pb-8 sm:pb-12 relative">
            <div
                class="absolute inset-0 z-0 opacity-[0.03] pointer-events-none"
                style="background-image: radial-gradient(#64748b 1px, transparent 1px); background-size: 20px 20px;"
            ></div>

            <main class="max-w-7xl mx-auto px-3 sm:px-4 md:px-5 lg:px-6 xl:px-8 pt-4 sm:pt-6 space-y-4 sm:space-y-6 relative z-10">
                <div
                    class="bg-white rounded-xl sm:rounded-2xl shadow-sm border border-slate-200 p-4 sm:p-5 md:p-6 flex flex-col md:flex-row items-center justify-between gap-4 sm:gap-6"
                >
                    <div class="flex items-center gap-3 sm:gap-4 md:gap-5 w-full md:w-auto">
                        <img
                            src="https://i.pravatar.cc/300?u={$auth.user?.username}"
                            alt="User Avatar"
                            class="w-12 h-12 sm:w-14 sm:h-14 md:w-16 md:h-16 rounded-full object-cover border-2 border-white shadow-md ring-2 ring-slate-100"
                        />
                        <div class="flex-1 min-w-0">
                            <h2 class="text-lg sm:text-xl font-bold text-slate-800 truncate">Halo, {$auth.user?.username}!</h2>
                            <div class="flex flex-wrap items-center gap-1.5 sm:gap-2 text-xs sm:text-sm text-slate-500 mt-1">
                                <span class="font-medium bg-slate-100 px-1.5 sm:px-2 py-0.5 rounded text-slate-600 truncate"
                                    >{$auth.user?.username}</span
                                >
                                <span class="hidden sm:inline">•</span>
                                <span class="capitalize truncate">{$auth.user?.role}</span>
                                <button onclick={handleLogout} class="ml-1.5 sm:ml-3 md:ml-4 text-rose-500 hover:underline font-bold text-xs sm:text-sm"
                                    >Logout</button
                                >
                            </div>
                        </div>
                    </div>

                    <div
                        class="w-full md:w-auto grid grid-cols-3 gap-3 sm:gap-4 md:gap-6 border-t md:border-t-0 border-slate-100 pt-3 sm:pt-4 md:pt-0"
                    >
                        <div class="text-center">
                            <p class="text-[10px] xs:text-xs text-slate-400 font-bold uppercase tracking-wider truncate">Total Output</p>
                            <p class="text-lg sm:text-xl md:text-2xl font-bold text-slate-800 mt-0.5">
                                {stats.totalOutput.toLocaleString('id-ID')}
                                <span class="text-xs sm:text-sm font-normal text-slate-500">Lot</span>
                            </p>
                        </div>
                        <div class="text-center">
                            <p class="text-[10px] xs:text-xs text-slate-400 font-bold uppercase tracking-wider truncate">Tingkat Reject</p>
                            <p class="text-lg sm:text-xl md:text-2xl font-bold text-rose-600 mt-0.5">{stats.rejectRate}</p>
                        </div>
                        <div class="text-center">
                            <p class="text-[10px] xs:text-xs text-slate-400 font-bold uppercase tracking-wider truncate">Shift Aktif</p>
                            <p class="text-lg sm:text-xl md:text-2xl font-bold text-blue-600 mt-0.5">{stats.activeShift}</p>
                        </div>
                    </div>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-5 md:gap-6">
                    {#each deptSummary as dept}
                        {@const actual = getDeptOutput(dept.id)}
                        <div
                            class={`rounded-xl sm:rounded-2xl p-4 sm:p-5 md:p-6 border transition-all duration-300 hover:shadow-md sm:hover:shadow-lg hover:-translate-y-0.5 sm:hover:-translate-y-1 ${getCardColor(dept.color)}`}
                        >
                            <div class="flex justify-between items-start mb-3 sm:mb-4">
                                <div class={`p-2 sm:p-3 rounded-lg sm:rounded-xl ${getIconBg(dept.color)}`}>
                                    <svg class="w-5 h-5 sm:w-6 sm:h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d={dept.icon}
                                        />
                                    </svg>
                                </div>
                                <div class="text-right">
                                    <span class="block text-xl sm:text-2xl font-bold">{dept.avgEfficiency}%</span>
                                    <span class="text-xs font-semibold opacity-70">Efisiensi</span>
                                </div>
                            </div>
                            <h3 class="text-base sm:text-lg font-bold mb-1">{dept.name}</h3>
                            <div class="space-y-2 sm:space-y-3 mt-3 sm:mt-4">
                                <div class="flex justify-between text-xs font-bold uppercase opacity-60">
                                    <span>Progress Lot</span>
                                    <span>{actual} / {dept.target}</span>
                                </div>
                                <div class="w-full h-1.5 sm:h-2 bg-white/50 rounded-full overflow-hidden">
                                    <div
                                        class={`h-full rounded-full ${getBarColor(dept.color)}`}
                                        style={`width: ${Math.min((actual / dept.target) * 100, 100)}%`}
                                    ></div>
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            </main>
        </div>
    {/if}
</div>

<style>
    /* Custom Scrollbar */
    .overflow-x-auto::-webkit-scrollbar {
        height: 6px;
    }
    .overflow-x-auto::-webkit-scrollbar-track {
        background: #f1f5f9;
        border-radius: 10px;
    }
    .overflow-x-auto::-webkit-scrollbar-thumb {
        background-color: #cbd5e1;
        border-radius: 10px;
    }
    .overflow-x-auto::-webkit-scrollbar-thumb:hover {
        background-color: #94a3b8;
    }
    
    /* Responsive Font Sizes */
    @media (max-width: 640px) {
        .min-w-0 {
            min-width: 0;
        }
    }
    
    /* Touch-friendly targets */
    @media (max-width: 768px) {
        button, 
        input, 
        [role="button"] {
            min-height: 44px;
        }
    }
</style>
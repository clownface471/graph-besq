<script lang="ts">
    import './layout.css';
    import { auth } from '$lib/stores/auth';
    import { browser } from '$app/environment';
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';
    import { onMount } from 'svelte';
    import Swal from 'sweetalert2';

    let { children } = $props();
    
    let sidebarOpen = $state(false);
    let isMobile = $state(false);
    let isChecking = $state(true); 
    let isAuthorized = $state(false);

    const hideSidebarRoutes: string[] = []; 
    let showSidebar = $derived($auth.isLoggedIn && !hideSidebarRoutes.includes($page.url.pathname));

    // --- 1. INISIALISASI SESI SAAT APLIKASI DIMUAT ---
    onMount(() => {
        auth.initialize();
    });

    // --- 2. LOGIKA SECURITY CHECK ---
    $effect(() => {
        if (!browser) return;

        const pathname = $page.url.pathname;
        const user = $auth.user;
        const isLoggedIn = $auth.isLoggedIn;

        // A. Jika di Halaman Login (root)
        if (pathname === '/') {
            if (isLoggedIn && user) {
                redirectBasedOnRole(user.role, user.department);
            } else {
                isAuthorized = true; 
            }
            isChecking = false;
            return;
        }

        // B. Jika Halaman Lain tapi BELUM Login
        if (!isLoggedIn) {
            // Tunggu sebentar untuk memastikan initialize sudah jalan
            const token = localStorage.getItem('auth_token');
            if (!token) {
                window.location.href = '/';
                return;
            }
        }

        // C. Cek Permission Role
        if (user) {
            let allowed = false;

            // URUTAN PENTING: Cek yang paling spesifik dulu!
            if (pathname.includes('mesin-detail')) { 
                // Level 3: Semua role boleh
                allowed = true; 
            } else if (pathname.includes('prs-ldr')) { 
                // Level 2: Manager & Leader
                allowed = user.role === 'MANAGER' || user.role === 'LEADER';
            } else if (pathname === '/manager') { 
                // Level 1: Hanya Manager
                allowed = user.role === 'MANAGER';
            } else if (pathname.startsWith('/manager')) {
                // Fallback untuk route lain di bawah manager
                allowed = user.role === 'MANAGER';
            } else {
                allowed = true; // Halaman umum
            }
            
            if (!allowed) {
                Swal.fire({
                    icon: 'error',
                    title: 'Akses Ditolak',
                    text: `Role ${user.role} tidak diizinkan mengakses halaman ini.`,
                    timer: 2000,
                    showConfirmButton: false
                }).then(() => {
                    redirectBasedOnRole(user.role, user.department);
                });
                isAuthorized = false;
            } else {
                isAuthorized = true;
            }
        }

        isChecking = false;
        
        isMobile = window.innerWidth < 768;
        if (!isMobile) sidebarOpen = false;
    });

    function redirectBasedOnRole(role: string, dept: string = '') {
        if (role === 'MANAGER') {
            goto('/manager');
        } else if (role === 'LEADER') {
            goto(`/manager/prs-ldr?process=${dept || 'PRESSING'}`);
        } else if (role === 'OPERATOR') {
            goto(`/manager/mesin-detail?no_mc=01A`);
        } else {
            goto('/');
        }
    }

    function toggleSidebar() {
        sidebarOpen = !sidebarOpen;
    }

    function getInitials(name: string): string {
        return name ? name.substring(0, 2).toUpperCase() : 'U';
    }
    
    async function handleLogoutConfirm() {
        const result = await Swal.fire({
            title: "Logout?",
            text: "Anda akan keluar dari sesi ini.",
            icon: "warning",
            showCancelButton: true,
            confirmButtonColor: '#e11d48',
            confirmButtonText: "Ya, Keluar",
            cancelButtonText: "Batal"
        });

        if (result.isConfirmed) {
            auth.logout();
            goto('/');
        }
    }
</script>

<svelte:head>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
</svelte:head>

{#if isChecking}
    <div class="min-h-screen flex items-center justify-center bg-slate-50">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
    </div>

{:else if isAuthorized}
    {#if showSidebar} 
        <div class="md:hidden fixed top-0 left-0 right-0 bg-white/95 backdrop-blur-xl border-b border-slate-100 shadow-md z-50 px-4 py-3">
            <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                    <button onclick={toggleSidebar} aria-label="Toggle Sidebar" class="p-2.5 rounded-lg hover:bg-slate-100">
                        <svg class="w-6 h-6 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                        </svg>
                    </button>
                    <span class="font-bold text-slate-800">BESQ MONITOR</span>
                </div>
                <div class="w-8 h-8 rounded-full bg-indigo-600 flex items-center justify-center text-white text-xs font-bold">
                    {getInitials($auth.user?.username || '')}
                </div>
            </div>
        </div>

        {#if sidebarOpen && isMobile}
            <div class="fixed inset-0 bg-black/40 z-40 md:hidden" 
                 role="button" 
                 tabindex="0"
                 onclick={() => sidebarOpen = false}
                 onkeydown={(e) => { if(e.key === 'Escape') sidebarOpen = false; }}>
            </div>
        {/if}

        <aside class="fixed left-0 top-0 h-screen w-64 bg-white border-r border-slate-200 z-50 transition-transform duration-300 md:translate-x-0 pt-16 md:pt-0 flex flex-col"
            class:translate-x-0={sidebarOpen}
            class:-translate-x-full={!sidebarOpen && isMobile}
        >
            <div class="hidden md:flex p-6 border-b border-slate-100 items-center gap-3">
                <div class="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center text-white font-bold">B</div>
                <span class="font-bold text-lg text-slate-800">BESQ Portal</span>
            </div>

            <nav class="flex-1 p-4 space-y-1 overflow-y-auto">
                {#if $auth.user?.role === 'MANAGER'}
                    <a href="/manager" 
                       class="flex items-center gap-3 px-4 py-3 rounded-lg text-slate-600 hover:bg-slate-50 font-medium"
                       class:text-indigo-600={$page.url.pathname === '/manager'} 
                       class:bg-indigo-50={$page.url.pathname === '/manager'}>
                        <i class="fa-solid fa-chart-pie w-5 text-center"></i>
                        Dashboard Manager
                    </a>
                {/if}

                {#if $auth.user?.role === 'MANAGER' || $auth.user?.role === 'LEADER'}
                     <div class="mt-4 mb-2 px-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Produksi</div>
                     <a href="/manager/prs-ldr?process=PRESSING" 
                       class="flex items-center gap-3 px-4 py-3 rounded-lg text-slate-600 hover:bg-slate-50 font-medium"
                       class:text-indigo-600={$page.url.pathname.includes('prs-ldr')} 
                       class:bg-indigo-50={$page.url.pathname.includes('prs-ldr')}>
                        <i class="fa-solid fa-industry w-5 text-center"></i>
                        Pressing Dept
                    </a>
                {/if}

                 <div class="mt-4 mb-2 px-4 text-xs font-bold text-slate-400 uppercase tracking-wider">Mesin</div>
                 <a href="/manager/mesin-detail?no_mc=01A" 
                   class="flex items-center gap-3 px-4 py-3 rounded-lg text-slate-600 hover:bg-slate-50 font-medium"
                   class:text-indigo-600={$page.url.pathname.includes('mesin-detail')} 
                   class:bg-indigo-50={$page.url.pathname.includes('mesin-detail')}>
                    <i class="fa-solid fa-microchip w-5 text-center"></i>
                    Detail Mesin
                </a>
            </nav>

            <div class="p-4 border-t border-slate-100">
                <button onclick={handleLogoutConfirm} class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-rose-600 hover:bg-rose-50 font-medium transition-colors">
                    <i class="fa-solid fa-right-from-bracket w-5 text-center"></i>
                    Keluar
                </button>
            </div>
        </aside>
    {/if}

    <div class="min-h-screen bg-slate-50 transition-all duration-300"
         class:md:ml-64={!isMobile && showSidebar} 
         class:pt-16={isMobile && showSidebar}
    >
        <div class={showSidebar ? "p-4 sm:p-6" : "p-0"}>
            {@render children()}
        </div>
    </div>

{:else}
    <div class="min-h-screen bg-slate-50"></div>
{/if}

<style>
    nav::-webkit-scrollbar { width: 4px; }
    nav::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 10px; }
</style>
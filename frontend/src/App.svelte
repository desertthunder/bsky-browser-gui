<script lang="ts">
  import "@fontsource-variable/jetbrains-mono";
  import "@fontsource-variable/geist";
  import "@fontsource-variable/lora";
  import { onMount } from "svelte";
  import { fade, slide } from "svelte/transition";
  import { GetVersion } from "../wailsjs/go/main/App";
  import { Login, Logout, Whoami, IsAuthenticated } from "../wailsjs/go/main/AuthService";
  import { Refresh, IsIndexing } from "../wailsjs/go/main/IndexService";
  import { Search, CountPosts } from "../wailsjs/go/main/SearchService";
  import { EventsOn } from "../wailsjs/runtime/runtime";
  import SearchBar from "./lib/components/SearchBar.svelte";
  import DataTable from "./lib/components/DataTable.svelte";
  import LogViewer from "./lib/components/LogViewer.svelte";
  import Toaster from "./lib/components/Toast.svelte";
  import { toaster } from "./lib/stores/toast.svelte";
  import EmptyState from "./lib/components/EmptyState.svelte";
  import ProgressBar from "./lib/components/ProgressBar.svelte";
  import PostDetailPanel from "./lib/components/PostDetailPanel.svelte";
  import type { main } from "../wailsjs/go/models";
  import type { IndexStats } from "./lib/types";

  type AuthInfo = { handle: string; did: string };
  const SEARCH_DEBOUNCE_MS = 200;

  let handle = $state("");
  let isLoading = $state(false);
  let status = $state("");
  let isLoggedIn = $state(false);
  let authInfo = $state<AuthInfo | null>(null);
  let isIndexing = $state(false);
  let refreshLimit = $state(0);
  let indexStats = $state<IndexStats>({ fetched: 0, inserted: 0, errors: 0, total: 0 });
  let showProgress = $state(false);
  let searchQuery = $state("");
  let searchSource = $state("");
  let searchResults = $state<main.SearchResult[]>([]);
  let totalSearchResults = $state(0);
  let totalPosts = $state(0);
  let sortColumn = $state("created_at");
  let sortDirection = $state<"asc" | "desc">("desc");
  let isSearching = $state(false);
  let showLogs = $state(false);
  let selectedPost = $state<main.SearchResult | null>(null);
  let currentPage = $state(1);
  let pageSize = $state(25);
  let appVersion = $state("dev");
  let showAbout = $state(false);

  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null;
  let activeSearchToken = 0;
  let hasSearchFilters = $derived(searchQuery.trim().length > 0 || searchSource !== "");

  onMount(() => {
    document.addEventListener("keydown", handleGlobalKeydown);
    void loadVersion();

    void checkAuthStatus().then(() => {
      EventsOn("index:started", () => {
        isIndexing = true;
        showProgress = true;
        indexStats = { fetched: 0, inserted: 0, errors: 0, total: 0 };
      });

      EventsOn("index:progress", (stats: any) => {
        indexStats = stats;
      });

      EventsOn("index:done", (result: any) => {
        isIndexing = false;
        indexStats.total = result.total || 0;
        void loadPosts();

        if (result.errors > 0) {
          const inserted = Math.max((result.total || 0) - result.errors, 0);
          toaster.warning(`Indexed ${inserted} posts with ${result.errors} errors`);
        } else {
          toaster.success(`Indexed ${result.total} posts successfully`);
        }

        setTimeout(() => {
          showProgress = false;
        }, 3000);
      });

      void IsIndexing().then((indexing) => {
        isIndexing = indexing;
        if (isIndexing) {
          showProgress = true;
        }
      });

      void loadPosts();
    });

    return () => {
      clearSearchDebounce();
      document.removeEventListener("keydown", handleGlobalKeydown);
    };
  });

  async function loadVersion() {
    try {
      appVersion = await GetVersion();
    } catch (err) {
      console.error("Failed to load version:", err);
    }
  }

  async function checkAuthStatus() {
    try {
      isLoggedIn = await IsAuthenticated();
      if (isLoggedIn) {
        const auth = await Whoami(false);
        if (auth) {
          authInfo = { handle: auth.handle, did: auth.did };
          status = `Logged in as @${auth.handle}`;
        }
      } else {
        status = "Please log in to continue";
      }
    } catch (err) {
      status = "Failed to check authentication status";
      toaster.error("Failed to check authentication status");
    }
  }

  async function handleLogin() {
    if (!handle.trim()) {
      status = "Please enter your Bluesky handle";
      toaster.warning("Please enter your Bluesky handle");
      return;
    }

    isLoading = true;
    status = "Opening browser for authentication...";

    try {
      await Login(handle.trim());
      status = "Login successful!";
      toaster.success("Login successful!");
      await checkAuthStatus();
    } catch (err) {
      status = `Login failed: ${err}`;
      toaster.error(`Login failed: ${err}`);
    } finally {
      isLoading = false;
    }
  }

  async function handleRefresh() {
    if (isIndexing) return;

    try {
      const sanitizedLimit = Math.max(0, Math.trunc(Number(refreshLimit) || 0));
      refreshLimit = sanitizedLimit;
      await Refresh(sanitizedLimit);
    } catch (err) {
      status = `Refresh failed: ${err}`;
      toaster.error(`Refresh failed: ${err}`);
    }
  }

  async function handleLogout() {
    try {
      await Logout();
      isLoggedIn = false;
      authInfo = null;
      searchResults = [];
      totalPosts = 0;
      totalSearchResults = 0;
      searchQuery = "";
      searchSource = "";
      handle = "";
      selectedPost = null;
      currentPage = 1;
      status = "Please log in to continue";
      toaster.success("Logged out");
    } catch (err) {
      toaster.error(`Logout failed: ${err}`);
    }
  }

  async function loadPosts() {
    try {
      totalPosts = await CountPosts();
      performSearch(searchQuery, searchSource, true, 1);
    } catch (err) {
      console.error("Failed to load posts:", err);
      toaster.error("Failed to load posts");
    }
  }

  function clearSearchDebounce() {
    if (searchDebounceTimer !== null) {
      clearTimeout(searchDebounceTimer);
      searchDebounceTimer = null;
    }
  }

  async function runSearch(query: string, source: string, page: number, searchToken: number) {
    isSearching = true;
    try {
      const searchPage = await Search(query.trim(), source, page, pageSize, sortColumn, sortDirection);
      if (searchToken !== activeSearchToken) {
        return;
      }

      searchResults = searchPage.results;
      totalSearchResults = searchPage.total;
      currentPage = searchPage.page;
      if (selectedPost && !searchPage.results.some((post) => post.uri === selectedPost?.uri)) {
        selectedPost = null;
      }
    } catch (err) {
      if (searchToken !== activeSearchToken) {
        return;
      }

      console.error("Search failed:", err);
      toaster.error("Search failed");
    } finally {
      if (searchToken === activeSearchToken) {
        isSearching = false;
      }
    }
  }

  function performSearch(query: string, source: string, immediate = false, page = 1) {
    clearSearchDebounce();
    const searchToken = ++activeSearchToken;

    const executeSearch = () => {
      searchDebounceTimer = null;
      void runSearch(query, source, page, searchToken);
    };

    if (immediate) {
      executeSearch();
      return;
    }

    searchDebounceTimer = setTimeout(executeSearch, SEARCH_DEBOUNCE_MS);
  }

  function handleSort(column: string) {
    if (sortColumn === column) {
      sortDirection = sortDirection === "asc" ? "desc" : "asc";
    } else {
      sortColumn = column;
      sortDirection = "desc";
    }
    performSearch(searchQuery, searchSource, true, currentPage);
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" && !isLoading) {
      void handleLogin();
    }
  }

  function handleGlobalKeydown(event: KeyboardEvent) {
    if (event.key === "Escape" && showAbout) {
      showAbout = false;
      return;
    }

    if ((event.metaKey || event.ctrlKey) && event.key === "k") {
      event.preventDefault();
      const searchInput = document.getElementById("search-posts") as HTMLInputElement | null;
      if (searchInput) {
        searchInput.focus();
        searchInput.select();
      }
    }

    if ((event.metaKey || event.ctrlKey) && event.shiftKey && event.key.toLowerCase() === "r") {
      event.preventDefault();
      if (!isIndexing) {
        void handleRefresh();
      }
    }

    if ((event.metaKey || event.ctrlKey) && event.key === "l") {
      event.preventDefault();
      showLogs = !showLogs;
    }
  }

  function clearSearchFilters() {
    searchQuery = "";
    searchSource = "";
    performSearch("", "", true, 1);
  }

  function handleSearchInput(query: string, source: string) {
    currentPage = 1;
    performSearch(query, source, false, 1);
  }

  function handlePageChange(page: number) {
    if (page === currentPage) {
      return;
    }

    performSearch(searchQuery, searchSource, true, page);
  }
</script>

<Toaster />

<main class="text-bright flex min-h-screen flex-col bg-black">
  {#if !isLoggedIn}
    <!-- Login View -->
    <div class="flex flex-1 items-center justify-center p-4" transition:fade={{ duration: 300 }}>
      <div class="w-full max-w-md">
        <div class="mb-8 text-center">
          <h1 class="mb-2 font-serif text-4xl">bsky-browser</h1>
          <p class="text-muted font-mono text-sm">Search your Bluesky bookmarks and likes</p>
        </div>

        <div class="bg-surface border-outline rounded-lg border p-6">
          <div class="space-y-4">
            <div>
              <label for="handle" class="text-muted mb-2 block font-sans text-sm">Bluesky Handle</label>
              <div class="relative">
                <input
                  id="handle"
                  type="text"
                  placeholder="username.bsky.social"
                  bind:value={handle}
                  onkeydown={handleKeydown}
                  disabled={isLoading}
                  class="border-outline text-bright w-full rounded border bg-black px-4 py-2 pr-10 font-mono text-sm placeholder-[#333] focus:border-[#333] focus:outline-none disabled:opacity-50" />

                {#if isLoading}
                  <span class="text-muted pointer-events-none absolute inset-y-0 right-3 flex items-center">
                    <i class="i-ri-loader-4-line animate-spin text-base"></i>
                  </span>
                {/if}
              </div>
            </div>

            <button
              onclick={() => void handleLogin()}
              disabled={isLoading || !handle.trim()}
              class="bg-surface border-outline hover:bg-outline text-bright w-full rounded border px-4 py-2 font-sans transition-colors disabled:cursor-not-allowed disabled:opacity-50">
              {#if isLoading}
                <span class="flex items-center justify-center gap-2">
                  <i class="i-ri-loader-4-line animate-spin text-base"></i>
                  <span>Authenticating...</span>
                </span>
              {:else}
                Login with Bluesky
              {/if}
            </button>
          </div>

          {#if status}
            <div class="border-outline mt-4 rounded border bg-black p-3" transition:slide={{ duration: 200 }}>
              <p class="text-muted font-mono text-xs">{status}</p>
            </div>
          {/if}

          <div class="mt-4 flex items-center justify-between border-t border-white/6 pt-4">
            <p class="text-muted font-mono text-xs tracking-[0.14em] uppercase">Version {appVersion}</p>
            <button
              type="button"
              onclick={() => (showAbout = true)}
              class="text-muted hover:text-bright font-sans text-xs transition-colors">
              About
            </button>
          </div>
        </div>
      </div>
    </div>
  {:else}
    <!-- Main View -->
    <div class="flex flex-1 flex-col">
      <!-- Header -->
      <header class="border-secondary bg-surface border-b px-6 py-4">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="font-serif text-xl">bsky-browser</h1>
            <p class="text-muted font-mono text-xs">@{authInfo?.handle} · {totalPosts} posts indexed</p>
          </div>

          <div class="flex flex-wrap items-center justify-end gap-3">
            <button
              onclick={() => (showAbout = true)}
              class="bg-surface border-outline hover:bg-outline text-bright rounded border px-3 py-2 font-mono text-xs transition-colors">
              <span class="flex items-center gap-2">
                <i class="i-ri-information-line"></i>
                <span>{appVersion}</span>
              </span>
            </button>

            <button
              onclick={() => (showLogs = !showLogs)}
              class="bg-surface border-outline hover:bg-outline text-bright rounded border px-3 py-2 font-mono text-xs transition-colors {showLogs
                ? 'bg-[#333]'
                : ''}">
              {#if showLogs}
                <span class="flex items-center gap-2">
                  <i class="i-ri-eye-off-line"></i>
                  <span>Hide Logs</span>
                </span>
              {:else}
                <span class="flex items-center gap-2">
                  <i class="i-ri-eye-line"></i>
                  <span>Show Logs</span>
                </span>
              {/if}
            </button>

            <div class="flex items-center gap-2">
              <label for="refreshLimit" class="text-muted font-sans text-xs">Limit:</label>
              <input
                id="refreshLimit"
                type="number"
                min="0"
                bind:value={refreshLimit}
                disabled={isIndexing}
                class="border-outline text-bright w-20 rounded border bg-black px-2 py-1 font-mono text-sm focus:border-[#333] focus:outline-none disabled:opacity-50" />
            </div>

            <button
              onclick={() => void handleRefresh()}
              disabled={isIndexing}
              class="bg-surface border-outline hover:bg-outline text-bright rounded border px-4 py-2 font-sans transition-colors disabled:cursor-not-allowed disabled:opacity-50">
              {#if isIndexing}
                <span class="animate-pulse">Refreshing...</span>
              {:else}
                <span class="flex items-center gap-2">
                  <i class="i-ri-refresh-line"></i>
                  <span>Refresh</span>
                </span>
              {/if}
            </button>

            <button
              onclick={() => void handleLogout()}
              class="bg-surface border-outline hover:bg-outline text-bright rounded border px-4 py-2 font-sans transition-colors">
              <span class="flex items-center gap-2">
                <i class="i-ri-logout-box-r-line"></i>
                <span>Logout</span>
              </span>
            </button>
          </div>
        </div>
      </header>

      <div class="border-outline border-b px-6 py-4">
        <SearchBar bind:query={searchQuery} bind:source={searchSource} bind:pageSize onSearch={handleSearchInput} />
      </div>

      <main class="flex-1 overflow-hidden p-6">
        {#if isSearching}
          <div class="flex h-full items-center justify-center">
            <span class="text-muted animate-pulse font-sans">Searching...</span>
          </div>
        {:else if totalPosts === 0}
          <EmptyState onRefresh={handleRefresh} />
        {:else if searchResults.length === 0}
          <div
            class="border-outline bg-surface flex h-full flex-col items-center justify-center rounded-[1.25rem] border px-8 py-12 text-center shadow-[0_18px_60px_rgba(0,0,0,0.35)]">
            <div class="text-muted mb-5 rounded-full border border-white/8 bg-black/70 p-4">
              <i class="i-ri-search-eye-line text-2xl"></i>
            </div>
            <h2 class="text-bright font-serif text-2xl">No matching posts</h2>
            <p class="text-muted mt-3 max-w-xl font-sans leading-6">
              {#if hasSearchFilters}
                No posts matched your current search. Try a different query, switch the source filter, or clear the
                current filters to browse recent posts again.
              {:else}
                No posts are available for this view right now. Refresh your index and try again.
              {/if}
            </p>
            {#if hasSearchFilters}
              <button
                type="button"
                onclick={clearSearchFilters}
                class="bg-surface border-outline hover:bg-outline text-bright mt-6 rounded-lg border px-5 py-2.5 font-sans text-sm transition-colors">
                Clear Search
              </button>
            {/if}
          </div>
        {:else}
          <div class="flex h-full min-h-0 flex-col gap-6 xl:flex-row">
            <div class="min-h-0 min-w-0 flex-1">
              <DataTable
                posts={searchResults}
                totalPosts={totalSearchResults}
                {currentPage}
                {pageSize}
                {sortColumn}
                {sortDirection}
                selectedPostURI={selectedPost?.uri ?? null}
                onSort={handleSort}
                onPageChange={handlePageChange}
                onOpenPost={(post) => {
                  selectedPost = post;
                }} />
            </div>

            {#if selectedPost}
              <div class="min-h-88 xl:h-full" transition:slide={{ axis: "x", duration: 220 }}>
                <PostDetailPanel
                  post={selectedPost}
                  onClose={() => {
                    selectedPost = null;
                  }} />
              </div>
            {/if}
          </div>
        {/if}
      </main>

      {#if showLogs}
        <div transition:slide={{ duration: 300 }}>
          <LogViewer visible={showLogs} />
        </div>
      {/if}

      {#if showProgress}
        <ProgressBar {isIndexing} {indexStats} />
      {/if}
    </div>
  {/if}
</main>

{#if showAbout}
  <div class="bg-surface/20 fixed inset-0 z-50 flex items-center justify-center p-4 backdrop-blur-sm">
    <button type="button" class="absolute inset-0" aria-label="Close about dialog" onclick={() => (showAbout = false)}>
    </button>

    <section
      class="border-outline bg-surface relative z-10 flex w-full max-w-lg flex-col gap-6 rounded-3xl border p-6 shadow-[0_24px_80px_rgba(0,0,0,0.55)]"
      transition:fade={{ duration: 180 }}>
      <div class="flex items-start justify-between gap-4">
        <div>
          <p class="text-muted font-mono text-xs tracking-[0.18em] uppercase">About</p>
          <h2 class="mt-2 font-serif text-3xl">bsky-browser</h2>
          <p class="text-muted mt-2 font-sans text-sm">Desktop search for your indexed Bluesky bookmarks and likes.</p>
        </div>

        <button
          type="button"
          class="text-muted hover:text-bright flex items-center rounded-full border border-white/8 p-2 transition-colors"
          aria-label="Close about dialog"
          onclick={() => (showAbout = false)}>
          <i class="i-ri-close-line text-lg"></i>
        </button>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="rounded-2xl border border-white/8 bg-black/50 p-4">
          <p class="text-muted font-mono text-[11px] tracking-[0.16em] uppercase">Version</p>
          <p class="mt-2 font-mono text-lg text-white">{appVersion}</p>
        </div>

        <div class="rounded-2xl border border-white/8 bg-black/50 p-4">
          <p class="text-muted font-mono text-[11px] tracking-[0.16em] uppercase">Refresh Shortcut</p>
          <p class="mt-2 font-mono text-lg text-white">Cmd/Ctrl + Shift + R</p>
        </div>
      </div>

      <div class="rounded-2xl border border-white/8 bg-black/50 p-4">
        <p class="text-muted font-mono text-[11px] tracking-[0.16em] uppercase">Included</p>
        <p class="text-bright mt-2 font-sans text-sm leading-6">
          OAuth, Keyword Search, Rich Text Post Rendering, and Live Log Viewer.
        </p>
      </div>
    </section>
  </div>
{/if}

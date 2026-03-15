<script lang="ts">
  import "@fontsource-variable/jetbrains-mono";
  import "@fontsource-variable/geist";
  import "@fontsource-variable/lora";
  import { onMount } from "svelte";
  import { fade, slide } from "svelte/transition";
  import { Login, Whoami, IsAuthenticated } from "../wailsjs/go/main/AuthService";
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
  import type { main } from "../wailsjs/go/models";
  import type { IndexStats } from "./lib/types";

  type AuthInfo = { handle: string; did: string };

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
  let totalPosts = $state(0);
  let sortColumn = $state("created_at");
  let sortDirection = $state<"asc" | "desc">("desc");
  let isSearching = $state(false);
  let showLogs = $state(false);

  onMount(() => {
    document.addEventListener("keydown", handleGlobalKeydown);

    checkAuthStatus().then(() => {
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
        loadPosts();
        toaster.success(`Indexed ${result.total} posts successfully`);
        setTimeout(() => {
          showProgress = false;
        }, 3000);
      });

      IsIndexing().then((indexing) => {
        isIndexing = indexing;
        if (isIndexing) {
          showProgress = true;
        }
      });

      loadPosts();
    });

    return () => {
      document.removeEventListener("keydown", handleGlobalKeydown);
    };
  });

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
      await Refresh(refreshLimit);
    } catch (err) {
      status = `Refresh failed: ${err}`;
      toaster.error(`Refresh failed: ${err}`);
    }
  }

  async function loadPosts() {
    try {
      totalPosts = await CountPosts();
      await performSearch(searchQuery, searchSource);
    } catch (err) {
      console.error("Failed to load posts:", err);
      toaster.error("Failed to load posts");
    }
  }

  async function performSearch(query: string, source: string) {
    isSearching = true;
    try {
      const results = await Search(query.trim(), source);
      searchResults = sortResults(results);
    } catch (err) {
      console.error("Search failed:", err);
      searchResults = [];
      toaster.error("Search failed");
    } finally {
      isSearching = false;
    }
  }

  function sortResults(results: main.SearchResult[]): main.SearchResult[] {
    return [...results].sort((a, b) => {
      let aVal: any = a[sortColumn as keyof main.SearchResult];
      let bVal: any = b[sortColumn as keyof main.SearchResult];

      if (sortColumn === "created_at" || sortColumn === "indexed_at") {
        aVal = aVal ? new Date(aVal).getTime() : 0;
        bVal = bVal ? new Date(bVal).getTime() : 0;
      }

      if (typeof aVal === "number" && typeof bVal === "number") {
        return sortDirection === "asc" ? aVal - bVal : bVal - aVal;
      }

      const aStr = String(aVal || "").toLowerCase();
      const bStr = String(bVal || "").toLowerCase();

      if (sortDirection === "asc") {
        return aStr < bStr ? -1 : aStr > bStr ? 1 : 0;
      } else {
        return aStr > bStr ? -1 : aStr < bStr ? 1 : 0;
      }
    });
  }

  function handleSort(column: string) {
    if (sortColumn === column) {
      sortDirection = sortDirection === "asc" ? "desc" : "asc";
    } else {
      sortColumn = column;
      sortDirection = "desc";
    }
    searchResults = sortResults(searchResults);
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" && !isLoading) {
      handleLogin();
    }
  }

  function handleGlobalKeydown(event: KeyboardEvent) {
    if ((event.metaKey || event.ctrlKey) && event.key === "k") {
      event.preventDefault();
      const searchInput = document.getElementById("search-posts") as HTMLInputElement | null;
      if (searchInput) {
        searchInput.focus();
        searchInput.select();
      }
    }

    if ((event.metaKey || event.ctrlKey) && event.key === "r") {
      event.preventDefault();
      if (!isIndexing) {
        handleRefresh();
      }
    }

    if ((event.metaKey || event.ctrlKey) && event.key === "l") {
      event.preventDefault();
      showLogs = !showLogs;
    }
  }
</script>

<Toaster />

<main class="min-h-screen bg-black text-bright flex flex-col">
  {#if !isLoggedIn}
    <!-- Login View -->
    <div class="flex-1 flex items-center justify-center p-4" transition:fade={{ duration: 300 }}>
      <div class="w-full max-w-md">
        <div class="text-center mb-8">
          <h1 class="font-serif text-4xl mb-2">bsky-browser</h1>
          <p class="font-mono text-muted text-sm">Search your Bluesky bookmarks and likes</p>
        </div>

        <div class="bg-surface border border-outline rounded-lg p-6">
          <div class="space-y-4">
            <div>
              <label for="handle" class="block font-sans text-sm text-muted mb-2"> Bluesky Handle </label>
              <input
                id="handle"
                type="text"
                placeholder="username.bsky.social"
                bind:value={handle}
                onkeydown={handleKeydown}
                disabled={isLoading}
                class="w-full bg-black border border-outline rounded px-4 py-2 font-mono text-sm text-bright placeholder-[#333] focus:outline-none focus:border-[#333] disabled:opacity-50" />
            </div>

            <button
              onclick={handleLogin}
              disabled={isLoading || !handle.trim()}
              class="w-full bg-surface border border-outline hover:bg-outline text-bright font-sans py-2 px-4 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
              {#if isLoading}
                <span class="animate-pulse">Authenticating...</span>
              {:else}
                Login with Bluesky
              {/if}
            </button>
          </div>

          {#if status}
            <div class="mt-4 p-3 bg-black border border-outline rounded" transition:slide={{ duration: 200 }}>
              <p class="font-mono text-xs text-muted">{status}</p>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {:else}
    <!-- Main View -->
    <div class="flex-1 flex flex-col">
      <!-- Header -->
      <header class="border-b border-outline bg-surface px-6 py-4">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="font-serif text-xl">bsky-browser</h1>
            <p class="font-mono text-xs text-muted">@{authInfo?.handle} · {totalPosts} posts indexed</p>
          </div>

          <div class="flex items-center gap-3">
            <button
              onclick={() => (showLogs = !showLogs)}
              class="font-mono text-xs px-3 py-2 rounded bg-surface border border-outline hover:bg-outline text-bright transition-colors {showLogs
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
              <label for="refreshLimit" class="font-sans text-xs text-muted">Limit:</label>
              <input
                id="refreshLimit"
                type="number"
                min="0"
                bind:value={refreshLimit}
                disabled={isIndexing}
                class="w-20 bg-black border border-outline rounded px-2 py-1 font-mono text-sm text-bright focus:outline-none focus:border-[#333] disabled:opacity-50" />
            </div>

            <button
              onclick={handleRefresh}
              disabled={isIndexing}
              class="bg-surface border border-outline hover:bg-outline text-bright font-sans py-2 px-4 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
              {#if isIndexing}
                <span class="animate-pulse">Refreshing...</span>
              {:else}
                <span class="flex items-center gap-2">
                  <i class="i-ri-refresh-line"></i>
                  <span>Refresh</span>
                </span>
              {/if}
            </button>
          </div>
        </div>
      </header>

      <div class="px-6 py-4 border-b border-secondary">
        <SearchBar bind:query={searchQuery} bind:source={searchSource} onSearch={performSearch} />
      </div>

      <main class="flex-1 p-6 overflow-hidden">
        {#if isSearching}
          <div class="flex items-center justify-center h-full">
            <span class="font-sans text-muted animate-pulse">Searching...</span>
          </div>
        {:else if totalPosts === 0}
          <EmptyState onRefresh={handleRefresh} />
        {:else}
          <DataTable posts={searchResults} {sortColumn} {sortDirection} onSort={handleSort} />
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

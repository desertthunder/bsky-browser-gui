<script lang="ts">
  import "@fontsource-variable/jetbrains-mono";
  import "@fontsource-variable/geist";
  import "@fontsource-variable/lora";
  import { onMount } from "svelte";
  import { fade, slide } from "svelte/transition";
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
  let selectedPost = $state<main.SearchResult | null>(null);
  let pageSize = $state(25);

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

  async function handleLogout() {
    try {
      await Logout();
      isLoggedIn = false;
      authInfo = null;
      searchResults = [];
      totalPosts = 0;
      searchQuery = "";
      searchSource = "";
      handle = "";
      selectedPost = null;
      status = "Please log in to continue";
      toaster.success("Logged out");
    } catch (err) {
      toaster.error(`Logout failed: ${err}`);
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
      const results = await Search(query.trim(), source, pageSize, sortColumn, sortDirection);
      searchResults = results;
      if (selectedPost && !results.some((post) => post.uri === selectedPost?.uri)) {
        selectedPost = null;
      }
    } catch (err) {
      console.error("Search failed:", err);
      searchResults = [];
      toaster.error("Search failed");
    } finally {
      isSearching = false;
    }
  }

  function handleSort(column: string) {
    if (sortColumn === column) {
      sortDirection = sortDirection === "asc" ? "desc" : "asc";
    } else {
      sortColumn = column;
      sortDirection = "desc";
    }
    performSearch(searchQuery, searchSource);
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
              <label for="handle" class="text-muted mb-2 block font-sans text-sm"> Bluesky Handle </label>
              <input
                id="handle"
                type="text"
                placeholder="username.bsky.social"
                bind:value={handle}
                onkeydown={handleKeydown}
                disabled={isLoading}
                class="border-outline text-bright w-full rounded border bg-black px-4 py-2 font-mono text-sm placeholder-[#333] focus:border-[#333] focus:outline-none disabled:opacity-50" />
            </div>

            <button
              onclick={handleLogin}
              disabled={isLoading || !handle.trim()}
              class="bg-surface border-outline hover:bg-outline text-bright w-full rounded border px-4 py-2 font-sans transition-colors disabled:cursor-not-allowed disabled:opacity-50">
              {#if isLoading}
                <span class="animate-pulse">Authenticating...</span>
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

          <div class="flex items-center gap-3">
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
              onclick={handleRefresh}
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
              onclick={handleLogout}
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
        <SearchBar bind:query={searchQuery} bind:source={searchSource} bind:pageSize onSearch={performSearch} />
      </div>

      <main class="flex-1 overflow-hidden p-6">
        {#if isSearching}
          <div class="flex h-full items-center justify-center">
            <span class="text-muted animate-pulse font-sans">Searching...</span>
          </div>
        {:else if totalPosts === 0}
          <EmptyState onRefresh={handleRefresh} />
        {:else}
          <div class="flex h-full min-h-0 flex-col gap-6 xl:flex-row">
            <div class="min-h-0 min-w-0 flex-1">
              <DataTable
                posts={searchResults}
                {sortColumn}
                {sortDirection}
                selectedPostURI={selectedPost?.uri ?? null}
                onSort={handleSort}
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

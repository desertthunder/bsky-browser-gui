<script lang="ts">
  import "@fontsource-variable/jetbrains-mono";
  import "@fontsource-variable/geist";
  import "@fontsource-variable/lora";
  import { onMount } from "svelte";
  import { Login, Whoami, IsAuthenticated } from "../wailsjs/go/main/AuthService";

  type AuthInfo = { handle: string; did: string };

  let handle = $state("");
  let isLoading = $state(false);
  let status = $state("");
  let isLoggedIn = $state(false);
  let authInfo = $state<AuthInfo | null>(null);

  onMount(async () => {
    await checkAuthStatus();
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
    }
  }

  async function handleLogin() {
    if (!handle.trim()) {
      status = "Please enter your Bluesky handle";
      return;
    }

    isLoading = true;
    status = "Opening browser for authentication...";

    try {
      await Login(handle.trim());
      status = "Login successful!";
      await checkAuthStatus();
    } catch (err) {
      status = `Login failed: ${err}`;
    } finally {
      isLoading = false;
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" && !isLoading) {
      handleLogin();
    }
  }
</script>

<main class="min-h-screen bg-black text-[#e5e5e5] flex items-center justify-center p-4">
  <div class="w-full max-w-md">
    {#if !isLoggedIn}
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
              class="w-full bg-black border border-outline rounded px-4 py-2 font-mono text-sm text-[#e5e5e5] placeholder-[#333] focus:outline-none focus:border-[#333] disabled:opacity-50" />
          </div>

          <button
            onclick={handleLogin}
            disabled={isLoading || !handle.trim()}
            class="w-full bg-surface border border-outline hover:bg-outline text-[#e5e5e5] font-sans py-2 px-4 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
            {#if isLoading}
              <span class="animate-pulse">Authenticating...</span>
            {:else}
              Login with Bluesky
            {/if}
          </button>
        </div>

        {#if status}
          <div class="mt-4 p-3 bg-black border border-outline rounded">
            <p class="font-mono text-xs text-muted">{status}</p>
          </div>
        {/if}
      </div>
    {:else}
      <div class="text-center">
        <h1 class="font-serif text-3xl mb-4">Welcome!</h1>
        <div class="bg-surface border border-outline rounded-lg p-6">
          <p class="font-sans text-[#e5e5e5] mb-2">
            Logged in as <span class="font-mono text-muted">@{authInfo?.handle}</span>
          </p>
          <p class="font-mono text-xs text-[#333] truncate" title={authInfo?.did}>
            {authInfo?.did}
          </p>
        </div>
        <p class="font-sans text-sm text-muted mt-8">Search functionality coming soon...</p>
      </div>
    {/if}
  </div>
</main>

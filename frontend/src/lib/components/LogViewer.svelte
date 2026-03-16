<script lang="ts">
  import { Clear, GetEntries } from "../../../wailsjs/go/main/LogService";
  import { EventsOn } from "../../../wailsjs/runtime/runtime";
  import { onMount } from "svelte";

  type LogLevel = "DEBUG" | "INFO" | "WARN" | "ERROR";

  type LogEntry = { level: LogLevel; message: string; timestamp: string };

  type Props = { visible: boolean };

  let { visible }: Props = $props();

  let logs = $state<LogEntry[]>([]);
  let scrollLock = $state(false);
  let logContainer: HTMLDivElement | undefined = $state(undefined);
  let filterLevel = $state<LogLevel | "ALL">("ALL");

  const levels: LogLevel[] = ["DEBUG", "INFO", "WARN", "ERROR"];

  function getLevelColor(level: LogLevel): string {
    switch (level) {
      case "DEBUG":
        return "text-gray-500";
      case "INFO":
        return "text-primary";
      case "WARN":
        return "text-yellow-400";
      case "ERROR":
        return "text-red-400";
    }
  }

  function getLevelBgColor(level: LogLevel | "ALL"): string {
    switch (level) {
      case "DEBUG":
        return "bg-gray-600";
      case "INFO":
        return "bg-blue-600";
      case "WARN":
        return "bg-yellow-600";
      case "ERROR":
        return "bg-red-600";
      default:
        return "bg-gray-600";
    }
  }

  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp);
    return date.toLocaleTimeString("en-US", { hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit" });
  }

  function scrollToBottom() {
    if (logContainer && !scrollLock) {
      logContainer.scrollTop = logContainer.scrollHeight;
    }
  }

  function toggleScrollLock() {
    scrollLock = !scrollLock;
  }

  function setFilterLevel(level: LogLevel | "ALL") {
    filterLevel = level;
  }

  function clearLogs() {
    logs = [];
    void Clear();
  }

  function filteredLogs() {
    if (filterLevel === "ALL") {
      return logs;
    }
    return logs.filter((log) => log.level === filterLevel);
  }

  onMount(() => {
    GetEntries()
      .then((entries) => {
        logs = entries.map((entry) => ({
          level: entry.level as LogLevel,
          message: entry.message,
          timestamp: entry.timestamp,
        }));
        setTimeout(scrollToBottom, 0);
      })
      .catch((err) => {
        console.error("Failed to load logs:", err);
      });

    EventsOn("log:line", (entry: LogEntry) => {
      logs = [...logs, entry];

      if (logs.length > 1000) {
        logs = logs.slice(logs.length - 1000);
      }

      setTimeout(scrollToBottom, 0);
    });

    EventsOn("log:cleared", () => {
      logs = [];
    });
  });

  $effect(() => {
    if (!scrollLock) {
      setTimeout(scrollToBottom, 0);
    }
  });
</script>

{#if visible}
  <div class="border-outline bg-surface/95 border-t shadow-[0_-18px_48px_rgba(0,0,0,0.35)] backdrop-blur">
    <div class="bg-surface border-outline relative flex items-center justify-between border-b px-4 py-2.5">
      <div class="flex items-center gap-3">
        <div
          class="bg-primary/12 text-primary flex h-8 w-8 items-center justify-center rounded-full border border-white/8">
          <i class="i-ri-terminal-box-line text-sm"></i>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-bright font-mono text-sm">Logs</span>
          <span class="text-muted font-mono text-xs">({logs.length})</span>
        </div>
      </div>

      <div class="text-muted hidden items-center gap-2 font-mono text-[11px] tracking-[0.14em] uppercase md:flex">
        <span class="rounded-full border border-white/8 px-2 py-1">Integrated Console</span>
        <span class="rounded-full border border-white/8 px-2 py-1">{scrollLock ? "Pinned" : "Follow"}</span>
      </div>

      <div class="absolute top-0 left-1/2 h-1.5 w-24 -translate-x-1/2 rounded-full bg-white/10"></div>
    </div>

    <div class="flex h-72 min-h-0 flex-1 flex-col bg-black/80">
      <div class="border-outline flex flex-wrap items-center justify-between gap-3 border-b px-4 py-2">
        <div class="flex flex-wrap items-center gap-1">
          {#each ["ALL", ...levels] as level}
            <button
              onclick={() => setFilterLevel(level as LogLevel | "ALL")}
              class="rounded-full px-2.5 py-1 font-mono text-xs transition-colors {filterLevel === level
                ? getLevelBgColor(level) + ' text-white'
                : 'bg-surface text-muted hover:text-bright border border-white/6'}">
              {level}
            </button>
          {/each}
        </div>

        <div class="flex items-center gap-2">
          <button
            onclick={toggleScrollLock}
            class="rounded-full px-3 py-1.5 font-mono text-xs transition-colors {scrollLock
              ? 'bg-yellow-600 text-white'
              : 'bg-surface text-muted hover:text-bright border border-white/6'}"
            title={scrollLock ? "Scroll locked" : "Auto-scroll enabled"}>
            {#if scrollLock}
              <span class="flex items-center gap-2">
                <i class="i-ri-lock-2-line"></i>
                <span>Locked</span>
              </span>
            {:else}
              <span class="flex items-center gap-2">
                <i class="i-ri-arrow-down-box-line"></i>
                <span>Follow</span>
              </span>
            {/if}
          </button>

          <button
            onclick={clearLogs}
            class="bg-surface text-muted rounded-full border border-white/6 px-3 py-1.5 font-mono text-xs transition-colors hover:text-red-400">
            Clear
          </button>
        </div>
      </div>

      <div bind:this={logContainer} class="min-h-0 flex-1 space-y-0.5 overflow-y-auto p-3 font-mono text-xs">
        {#each filteredLogs() as log}
          <div
            class="flex items-start gap-2 rounded-lg border border-transparent px-2 py-1.5 hover:border-white/6 hover:bg-white/4">
            <span class="text-muted shrink-0">{formatTimestamp(log.timestamp)}</span>
            <span class="{getLevelColor(log.level)} w-12 shrink-0">[{log.level}]</span>
            <span class="text-bright break-all">{log.message}</span>
          </div>
        {:else}
          <div class="text-muted py-8 text-center">No logs</div>
        {/each}
      </div>
    </div>
  </div>
{/if}

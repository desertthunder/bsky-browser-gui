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
  <div class="border-outline flex flex-col border-t bg-black">
    <!-- Header -->
    <div class="bg-surface border-outline flex items-center justify-between border-b px-4 py-2">
      <div class="flex items-center gap-2">
        <span class="text-bright font-mono text-sm">Logs</span>
        <span class="text-muted font-mono text-xs">({logs.length})</span>
      </div>

      <div class="flex items-center gap-2">
        <!-- Level Filter Buttons -->
        <div class="mr-4 flex items-center gap-1">
          {#each ["ALL", ...levels] as level}
            <button
              onclick={() => setFilterLevel(level as LogLevel | "ALL")}
              class="rounded px-2 py-1 font-mono text-xs transition-colors {filterLevel === level
                ? getLevelBgColor(level) + ' text-white'
                : 'text-muted hover:text-bright bg-black'}">
              {level}
            </button>
          {/each}
        </div>

        <!-- Scroll Lock Toggle -->
        <button
          onclick={toggleScrollLock}
          class="rounded px-2 py-1 font-mono text-xs transition-colors {scrollLock
            ? 'bg-yellow-600 text-white'
            : 'text-muted hover:text-bright bg-black'}"
          title={scrollLock ? "Scroll locked" : "Auto-scroll enabled"}>
          {#if scrollLock}
            <span class="flex items-center">
              <i class="i-ri-lock-2-line"></i>
            </span>
          {:else}
            <span class="flex items-center">
              <i class="i-ri-arrow-down-box-line"></i>
            </span>
          {/if}
        </button>

        <!-- Clear Button -->
        <button
          onclick={clearLogs}
          class="text-muted rounded bg-black px-2 py-1 font-mono text-xs transition-colors hover:text-red-400">
          Clear
        </button>
      </div>
    </div>

    <!-- Log Container -->
    <div
      bind:this={logContainer}
      class="flex-1 space-y-0.5 overflow-y-auto p-2 font-mono text-xs"
      style="max-height: 200px;">
      {#each filteredLogs() as log}
        <div class="flex items-start gap-2 rounded px-1 hover:bg-white/5">
          <span class="text-muted shrink-0">{formatTimestamp(log.timestamp)}</span>
          <span class="{getLevelColor(log.level)} w-12 shrink-0">[{log.level}]</span>
          <span class="text-bright break-all">{log.message}</span>
        </div>
      {:else}
        <div class="text-muted text-center py-4">No logs</div>
      {/each}
    </div>
  </div>
{/if}

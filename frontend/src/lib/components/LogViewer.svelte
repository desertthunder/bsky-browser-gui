<script lang="ts">
  import { EventsOn } from "../../../wailsjs/runtime/runtime";
  import { onMount } from "svelte";

  type LogLevel = "DEBUG" | "INFO" | "WARN" | "ERROR";

  interface LogEntry {
    level: LogLevel;
    message: string;
    timestamp: string;
  }

  interface Props {
    visible: boolean;
  }

  let { visible }: Props = $props();

  let logs = $state<LogEntry[]>([]);
  let scrollLock = $state(false);
  let logContainer: HTMLDivElement | undefined = $state(undefined);
  let filterLevel = $state<LogLevel | "ALL">("ALL");

  const levels: LogLevel[] = ["DEBUG", "INFO", "WARN", "ERROR"];

  const levelColors: Record<LogLevel, string> = {
    DEBUG: "text-gray-500",
    INFO: "text-blue-400",
    WARN: "text-yellow-400",
    ERROR: "text-red-400",
  };

  const levelBgColors: Record<LogLevel | "ALL", string> = {
    ALL: "bg-gray-700",
    DEBUG: "bg-gray-600",
    INFO: "bg-blue-600",
    WARN: "bg-yellow-600",
    ERROR: "bg-red-600",
  };

  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp);
    return date.toLocaleTimeString("en-US", {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
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
  }

  function filteredLogs() {
    if (filterLevel === "ALL") {
      return logs;
    }
    return logs.filter((log) => log.level === filterLevel);
  }

  onMount(() => {
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
  <div class="border-t border-outline bg-black flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-2 bg-surface border-b border-outline">
      <div class="flex items-center gap-2">
        <span class="font-mono text-sm text-bright">Logs</span>
        <span class="font-mono text-xs text-muted">({logs.length})</span>
      </div>

      <div class="flex items-center gap-2">
        <!-- Level Filter Buttons -->
        <div class="flex items-center gap-1 mr-4">
          {#each ["ALL", ...levels] as level}
            <button
              onclick={() => setFilterLevel(level as LogLevel | "ALL")}
              class="font-mono text-xs px-2 py-1 rounded transition-colors {filterLevel === level
                ? levelBgColors[level] + ' text-white'
                : 'bg-black text-muted hover:text-bright'}">
              {level}
            </button>
          {/each}
        </div>

        <!-- Scroll Lock Toggle -->
        <button
          onclick={toggleScrollLock}
          class="font-mono text-xs px-2 py-1 rounded transition-colors {scrollLock
            ? 'bg-yellow-600 text-white'
            : 'bg-black text-muted hover:text-bright'}"
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
          class="font-mono text-xs px-2 py-1 rounded bg-black text-muted hover:text-red-400 transition-colors">
          Clear
        </button>
      </div>
    </div>

    <!-- Log Container -->
    <div
      bind:this={logContainer}
      class="flex-1 overflow-y-auto p-2 font-mono text-xs space-y-0.5"
      style="max-height: 200px;">
      {#each filteredLogs() as log}
        <div class="flex items-start gap-2 hover:bg-white/5 px-1 rounded">
          <span class="text-muted shrink-0">{formatTimestamp(log.timestamp)}</span>
          <span class="{levelColors[log.level]} shrink-0 w-12">[{log.level}]</span>
          <span class="text-bright break-all">{log.message}</span>
        </div>
      {:else}
        <div class="text-muted text-center py-4">No logs</div>
      {/each}
    </div>
  </div>
{/if}

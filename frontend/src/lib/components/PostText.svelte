<script lang="ts">
  import { parseFacets, renderFacets, truncateRenderedFacets, type Facet } from "../facets";

  interface Props {
    text: string;
    facetsJson?: string;
    maxLength?: number;
    class?: string;
  }

  let { text, facetsJson = "", maxLength = 0, class: className = "" }: Props = $props();

  function getRenderedFacets(text: string, facetsJson: string, maxLength: number) {
    const facets = parseFacets(facetsJson);
    const rendered = renderFacets(text, facets);
    
    if (maxLength > 0 && maxLength < text.length) {
      const { facets: truncated } = truncateRenderedFacets(rendered, maxLength);
      return truncated;
    }
    
    return rendered;
  }

  let renderedFacets = $derived(getRenderedFacets(text, facetsJson, maxLength));
</script>

<span class="{className}">
  {#each renderedFacets as facet}
    {#if facet.type === 'link'}
      <a 
        href={facet.href} 
        target="_blank" 
        rel="noopener noreferrer"
        class="text-blue-400 hover:text-blue-300 hover:underline"
        onclick={(e) => e.stopPropagation()}>
        {facet.text}
      </a>
    {:else if facet.type === 'mention'}
      <a 
        href={facet.href} 
        target="_blank" 
        rel="noopener noreferrer"
        class="text-blue-400 hover:text-blue-300 hover:underline"
        onclick={(e) => e.stopPropagation()}>
        {facet.text}
      </a>
    {:else if facet.type === 'tag'}
      <a 
        href={facet.href} 
        target="_blank" 
        rel="noopener noreferrer"
        class="text-pink-400 hover:text-pink-300 hover:underline"
        onclick={(e) => e.stopPropagation()}>
        {facet.text}
      </a>
    {:else}
      <span>{facet.text}</span>
    {/if}
  {/each}
</span>

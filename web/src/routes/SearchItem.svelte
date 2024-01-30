<script lang="ts">
  import type { SearchItem } from "$lib/api";

  export let item: SearchItem;
  export let highlighted: boolean;
</script>

<div class="search-item">
  <div class="name">
    {item.name}
  </div>
  <div class="id">
    {item.id}
  </div>
</div>

<style>
  .search-item {
    display: flex;
    flex-direction: column;
  }
</style>

<script lang="ts">
  import {
    Configuration,
    DefaultApi,
    type SearchItem as SearchItemStruct,
  } from "$lib/api";
  import SearchItem from "./SearchItem.svelte";

  import { api } from "../store";

  let conf = new Configuration({
    basePath: "/",
  });

  let filteredAthletes: SearchItemStruct[];
  let inputValue: string;

  let searchRefresh = false;
  const filterAthletes = async () => {
    if (!searchRefresh) {
      filteredAthletes = await api.searchAthleteGet({ name: inputValue });
      searchRefresh = true;
      setTimeout(() => {
        searchRefresh = false;
      }, 1000);
    }
  };

  $: if (!inputValue) {
    filteredAthletes = [];
  } else {
    highlightedAthlete = filteredAthletes[highlightedIndex];
  }
  let highlightedAthlete: SearchItemStruct;
  let highlightedIndex: number;

  const submitValue = () => {};
</script>

<form autocomplete="off" on:submit|preventDefault={submitValue}>
  <div class="autocomplete">
    <input
      type="text"
      id="athlete-input"
      placeholder="Search Athlete"
      bind:value={inputValue}
      on:input={filterAthletes}
    />
  </div>

  <input type="submit" />

  {#if filteredAthletes.length > 0}
    <ul id="autocomplete-items-list">
      {#each filteredAthletes as athlete, i}
        <SearchItem item={athlete} highlighted={i === highlightedIndex} />
      {/each}
    </ul>
  {/if}
</form>

<style>
  div.autocomplete {
    position: relative;
    display: inline-block;
    width: 300px;
  }

  input {
    border: 1px solid transparent;
    background-color: #f1f1f1;
    padding: 10px;
    font-size: 16px;
    margin: 0;
  }
</style>

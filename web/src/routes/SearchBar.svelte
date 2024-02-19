<script>
    import SearchItem from "./SearchItem.svelte";

    export let inputValue;
    export let submitValue;
    export let highlightedIndex;

    var filteredAthletes = Array();
</script>

<form autocomplete="off" on:submit|preventDefault={submitValue}>
    <div class="autocomplete">
        <input
            type="text"
            id="athlete-input"
            placeholder="Search athlete, team, division, or conference"
            bind:value={inputValue}
            on:input={filteredAthletes}
        />
    </div>

    <input type="submit" />

    {#if filteredAthletes.length > 0}
        <ul id="autocomplete-items-list">
            {#each filteredAthletes as athlete, i}
                <SearchItem
                    item={athlete}
                    highlighted={i === highlightedIndex}
                />
            {/each}
        </ul>
    {/if}
</form>

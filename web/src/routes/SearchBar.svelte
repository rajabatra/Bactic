<script>
    import SearchItem from "./SearchItem.svelte";

    var filteredAthletes = Array();

    import { onMount } from "svelte";
    import SearchBar from "./SearchBar.svelte";

    var filteredAthletes;
    var inputValue;

    let searchRefresh = false;
    let delayedRefresh = false;
    const filterAthletes = async () => {
        if (inputValue) {
            if (!searchRefresh) {
                filteredAthletes = await fetch(
                    "/api/search/athlete?" +
                        new URLSearchParams({ name: inputValue }),
                    {
                        method: "GET",
                    },
                ).then((res) => {
                    return res.json();
                });

                searchRefresh = true;
                delayedRefresh = false;
                setTimeout(() => {
                    searchRefresh = false;
                }, 500);
            } else if (!delayedRefresh) {
                delayedRefresh = true;
                setTimeout(() => {
                    filterAthletes();
                }, 500);
            }
        }
    };

    let highlightedAthlete;
    let highlightedIndex;
    $: if (!inputValue) {
        filteredAthletes = [];
    } else {
        highlightedAthlete = filteredAthletes[highlightedIndex];
    }

    const submitValue = () => {};
</script>

<form autocomplete="off" on:submit|preventDefault={submitValue}>
    <div class="autocomplete">
        <input
            type="text"
            id="athlete-input"
            placeholder="Search athlete, team, division, or conference"
            bind:value={inputValue}
            on:input={filterAthletes}
        />
    </div>

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

<style>
    div.autocomplete {
        position: relative;
        display: inline-block;
        width: 300px;
    }

    form {
        padding: 0;
    }

    input {
        border: 1px solid transparent;
        border-radius: 0.5rem;
        height: 1.5em;
        width: 20em;
        background-color: #f1f1f1;
        padding: 10px;
        font-size: 16px;
        margin: 0;
    }

    input:focus {
        outline: none;
    }
</style>

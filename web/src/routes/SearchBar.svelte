<script>
    import SearchItem from "./SearchItem.svelte";

    var filteredAthletes = Array();

    // see model_search_item.go
    const item_to_class = {
        0: 'athlete',
        1: 'school',
        2: 'region',
    }

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
                    res = res.json();
                    return res;
                });

                filteredAthletes.forEach((ath) => {
                    ath.item_class = item_to_class[ath.item_type];
                    ath.url = `/${ath.item_class}/${ath.id}`;
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
        <ul id="results">
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

    #results {
        display: block;
        position: absolute;
        background: transparent;
        margin: 0;
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

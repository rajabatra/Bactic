<script>
    import { onMount } from "svelte";
    import SearchBar from "./SearchBar.svelte";

    var filteredAthletes;
    var inputValue;

    let searchRefresh = false;
    const filterAthletes = async () => {
        if (!searchRefresh) {
            filteredAthletes = await fetch("/api/search/athlete", {
                method: "GET",
                body: new URLSearchParams(`name=${inputValue}`),
            }).then((res) => {
                return res.body;
            });
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
    let highlightedAthlete;
    let highlightedIndex;

    const submitValue = () => {};

    // a list of stats article summaries that allow us to dynamically render the article list
    var statsSummeries = Array();
    onMount(async () => {
        const response = await fetch("/api/stats/summaries")
            .then((res) => {
                statsSummaries = res.json();
            })
            .catch((err) => {
                console.log(err);
            });
    });
</script>

<section class="search">
    <SearchBar />
</section>

<section class="article-summaries">
    {#each statsSummeries as article, i}{/each}
</section>

<style>
    div.autocomplete {
        position: relative;
        display: inline-block;
        width: 300px;
    }

    input {
        border: 1px solid transparent;
        border-radius: 20px;
        background-color: #f1f1f1;
        padding: 10px;
        font-size: 16px;
        margin: 0;
    }

    .article-summaries {
        display: grid;
    }
</style>

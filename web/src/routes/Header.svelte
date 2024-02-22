<script>
    import logo from "$lib/assets/logo.png";
    import SearchBar from "./SearchBar.svelte";
</script>

<nav>
    <a href="/" id="home-button">
        <h1 id="bactic-title">BACTIC</h1>
        <img src={logo} alt="Logo" id="bactic-logo" />
    </a>
    <SearchBar />
</nav>

<style>
    @import url("https://fonts.googleapis.com/css2?family=Open+Sans:ital,wght@0,300..800;1,300..800&display=swap");

    #home-button {
        display: flex;
        flex-direction: row;
        align-items: center;
        flex-basis: 40%;
    }
    #bactic-title {
        font-family: "Open Sans", sans-serif;
        font-weight: 800;
        font-style: italic;
        font-size: 2em;
        margin: 0 0.5rem;
    }

    #bactic-logo {
        height: 2em;
    }
    nav {
        display: flex;
        flex-direction: row;
        background-color: #4c80b4;
        overflow-x: hidden;
        align-items: center;
    }

    nav a {
        padding: 0.3rem 1rem;
        text-decoration: none;
        font-size: 25px;
        color: #89d4f3;
        transition: 0.3s;
    }

    nav a:hover {
        color: #f1f1f1;
    }
</style>

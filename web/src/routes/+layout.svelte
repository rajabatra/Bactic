<script>
    import Footer from "./Footer.svelte";
    import Header from "./Header.svelte";
</script>

<div class="main-content">
    <Header />
    <div class="slot-content">
        <slot />
    </div>

    <Footer />
</div>

<style>
    .main-content {
        background-color: rgb(0, 0, 0);
        margin: 0;
    }
    .slot-content {
        min-height: calc(100vh - 95px);
        padding: 0 20px;
    }
</style>

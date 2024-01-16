<script>
  import Footer from "./Footer.svelte";
  import Header from "./Header.svelte";
</script>

<Header />

<div class="main-content">
  <div class="slot-content">
    <slot />
  </div>
  <Footer />
</div>

<style>
  .main-content {
    position: absolute;
    left: 160px;
    width: calc(100% - 160px);
    padding: 0;
  }
  .slot-content {
    min-height: calc(100vh - 95px);
    padding: 0 20px;
  }
</style>

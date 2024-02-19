<script>
    export let item;
    export let highlighted;
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

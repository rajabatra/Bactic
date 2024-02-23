<script>
    export let item;
    export let highlighted;
</script>

<li class="search-item">
    <a href={item.url}>
        <div class="name">
            {item.name}
        </div>
        <div class="id">
            {item.id}
        </div>
    </a>
</li>

<style>
    .search-item {
        display: flex;
        flex-direction: row;
        background-color: #f1f1f1;
        padding: .1em .5em;
        width: 300px;
    }

    li.search-item:last-child{
        border-radius: 0 0 1em 1em;
    }
</style>

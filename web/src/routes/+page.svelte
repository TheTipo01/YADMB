<script>
    import Queue from "./queue.svelte";
    import Song from "./song.svelte"
    import Favorites from "./favorites.svelte"
    import {Heading} from "flowbite-svelte"
</script>
<Heading tag="h1" align="center">Gerry UI</Heading>
<div class="grid grid-cols-3 gap-4 m-5 items-center justify-item-center">
    <div>
        <Heading tag="h4">Queue</Heading>
        <Queue />
    </div>
    <div>
        <Song />
    </div>
    <div>
        <Heading tag="h4">Favorites</Heading>
        <Favorites />
    </div>
</div>


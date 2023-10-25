<script>
    import Queue from "./queue.svelte";
    import Favorites from "./favorites.svelte"
    import { Tabs, TabItem} from "flowbite-svelte"
    import {Avatar} from "flowbite-svelte"
    import StarSolid from "flowbite-svelte-icons/StarSolid.svelte"
    import ListMusicSolid from "flowbite-svelte-icons/ListMusicSolid.svelte"
    let GuildId = "very cool id";
    let token = "very cool token";

    let activetab = "queue";

</script>
<div class="flex justify-center">
    <Avatar src="src/assets/logo_yadmb.png" size="xl" class="mt-1"  />
</div>

<Tabs style="underline">
    <TabItem open active={activetab === "queue"} on:click={() => (activetab = "queue")}>
        <div slot="title" class="flex items-center gap-2">
            <ListMusicSolid />
            Queue
        </div>
        {#if activetab === "queue"}
            <Queue GuildId={GuildId} token={token} />
        {/if}
    </TabItem>
    <TabItem active={activetab === "favorites"} on:click={() => (activetab = "favorites")}>
        <div slot="title" class="flex items-center gap-2">
            <StarSolid />
            Favorites
        </div>
        {#if activetab === "favorites"}
            <Favorites token={token} />
        {/if}
    </TabItem>
</Tabs>
<div >
</div>


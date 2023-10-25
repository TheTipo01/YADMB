<script>
    import Queue from "./queue.svelte";
    import Favorites from "./favorites.svelte"
    import { Tabs, TabItem} from "flowbite-svelte"
    import StarSolid from "flowbite-svelte-icons/StarSolid.svelte"
    import ListMusicSolid from "flowbite-svelte-icons/ListMusicSolid.svelte"
    let GuildId = "very cool guild id";
    let token = "very cool token"

</script>
<Tabs style="underline">
    <TabItem open>
        <div slot="title" class="flex items-center gap-2">
            <ListMusicSolid />
            Queue
        </div>
        <Queue GuildId = {GuildId} token = {token} />
    </TabItem>
    <TabItem>
        <div slot="title" class="flex items-center gap-2">
            <StarSolid />
            Favorites
        </div>
        <Favorites token = {token} />
    </TabItem>
</Tabs>
<div >
</div>


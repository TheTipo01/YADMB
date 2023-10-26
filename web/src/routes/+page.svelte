<script>
    import Queue from "./queue.svelte";
    import Favorites from "./favorites.svelte"
    import { Tabs, TabItem} from "flowbite-svelte"
    import {Avatar} from "flowbite-svelte"
    import StarSolid from "flowbite-svelte-icons/StarSolid.svelte"
    import ListMusicSolid from "flowbite-svelte-icons/ListMusicSolid.svelte"
    import {GetGuildID, GetToken} from "../lib/utilities"
	import { onMount } from "svelte";
    import logo from "../assets/logo_yadmb.png"

    // variables
    let GuildId = '';
    let token = '';
    let activetab = "queue";

    onMount(() => {
        GuildId = GetGuildID();
        token = GetToken();
    });
    

</script>
<div class="flex justify-center">
    <Avatar src="{logo}" size="xl" class="mt-1"  />
</div>


<Tabs style="underline">
    <TabItem open active={activetab === "queue"} on:click={() => (activetab = "queue")}>
        <div slot="title" class="flex items-center gap-2">
            <ListMusicSolid />
            Queue
        </div>
        {#if activetab === "queue"}
            {#if GuildId !== '' && token !== ''}
                <Queue GuildId={GuildId} token={token} />
            {/if}
        {/if}
    </TabItem>
    <TabItem active={activetab === "favorites"} on:click={() => (activetab = "favorites")}>
        <div slot="title" class="flex items-center gap-2">
            <StarSolid />
            Favorites
        </div>
        {#if activetab === "favorites"}
            {#if token !== ''}
                <Favorites token={token} />
            {/if}
        {/if}
    </TabItem>
</Tabs>
<div>
</div>


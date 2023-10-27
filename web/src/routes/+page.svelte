<script>
    import Queue from "./queue.svelte";
    import Favorites from "./favorites.svelte"
    import { Tabs, TabItem} from "flowbite-svelte"
    import {Avatar} from "flowbite-svelte"
    import StarSolid from "flowbite-svelte-icons/StarSolid.svelte"
    import ListMusicSolid from "flowbite-svelte-icons/ListMusicSolid.svelte"
    import {GetGuildID, GetToken, GetHost} from "../lib/utilities"
	import { onMount } from "svelte";
    import logo from "../assets/logo_yadmb.png"

    // variables
    let GuildId = '';
    let token = '';
    let host = '';
    let activetab = "queue";

    onMount(() => {
        GuildId = GetGuildID();
        token = GetToken();
        host = GetHost();
        document.title = "YADMB Web UI"
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
            {#if GuildId !== '' && token !== '' && host !== ''}
                <Queue GuildId={GuildId} token={token} host={host} />
            {/if}
        {/if}
    </TabItem>
    <TabItem active={activetab === "favorites"} on:click={() => (activetab = "favorites")}>
        <div slot="title" class="flex items-center gap-2">
            <StarSolid />
            Favorites
        </div>
        {#if activetab === "favorites"}
            {#if token !== '' && host !== ''}
                <Favorites token={token} host={host} />
            {/if}
        {/if}
    </TabItem>
</Tabs>
<div>
</div>


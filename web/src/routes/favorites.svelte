<script>
    import { Button } from "flowbite-svelte";
    import { GetFavorites, AddFavorite, RemoveFavorite} from "../lib/favorites";
    import {KeyPressed} from "../lib/utilities"
    import {Modal} from "flowbite-svelte";
    import {Input, Label} from "flowbite-svelte"
    import {Heading, P} from "flowbite-svelte"
    import TrashBinSolid from "flowbite-svelte-icons/TrashBinSolid.svelte"
    import {PlusSolid} from "flowbite-svelte-icons";
    import {AddToQueue} from "../lib/queue"
    import {AddObjectToArray, RemoveObjectFromArray} from "../lib/utilities"
    import {Response} from "$lib/error.js";

    // props
    export let GuildId;
    export let token;
    export let host;

    // variables
    export let favorites = GetFavorites(token, host);
    let showModal = false;
</script>

<!-- Modal Button -->
<Button class="w-25 absolute right-9 bottom-5" on:click={() => (showModal = true)}>
    Add to Favorites
</Button>

<!-- Modal Component -->
<Modal title="Add Favorite" bind:open={showModal} autoclose>
    <form id="form">
        <!-- Form inputs -->
        <div class="grid grid-rows-3">
            <div>
                <Label for="name" class="mb-2">Song Name</Label>
                <Input type="text" id="name" required/>
            </div>
            <div>
                <Label for="link" class="mb-2">Song Link</Label>
                <Input type="text" id="link" required/>
            </div>
            <div>
                <Label for="folder" class="mb-2">Folder</Label>
                <Input type="text" id="folder" />
            </div>
        </div>
    </form>
    
    <!-- Submit button -->
    <svelte:fragment slot="footer">
        <Button on:click={ async () => {
            let result = await AddFavorite(token, host, favorites);
            switch(result) {
                // TODO: Add error handling
                case Response.FAVORITE_TOKEN_ERROR:
                case Response.DUPLICATE_ERROR:
                    break;
                default:
                    favorites = AddObjectToArray(favorites, result);
            }
        }} on:keydown={(e) => (KeyPressed(e, "favorite", "", token, host))}>Add</Button>
    </svelte:fragment>
</Modal>

{#await favorites}
    <P>Loading Favorites</P>
{:then favorite} 
    {#if favorite.length !== 0}
        <div class="grid grid-cols-3">
            <Heading tag="h5">Song Name</Heading>
            <Heading tag="h5">Link</Heading>
            <Heading tag="h5">Folder</Heading>
        </div>
        {#each favorite as song}
            <div class="grid grid-cols-3 mt-5">
                <P>{song.name}</P>
                <P>{song.link}</P>
                <div class="grid grid-cols-3">
                    <P>{song.folder}</P>
                    <Button on:click={ async () => {
                        let result = await RemoveFavorite(token, song.name, host);
                        switch(result) {
                            // TODO: Add error handling
                            case Response.FAVORITE_TOKEN_ERROR:
                            case Response.FAVORITE_NOT_FOUND:
                                break;
                            default:
                                favorites = RemoveObjectFromArray(favorites, song);
                        }
                    }} class="w-1/3"><TrashBinSolid /></Button>
                    <Button on:click={() => AddToQueue(GuildId, token, host, song.link, false, false, false, false)} class="w-1/3"><PlusSolid /></Button>
                </div>
                
            </div>
        {/each}
    {:else}
        <P>You have no favorites</P>
    {/if}
{/await}

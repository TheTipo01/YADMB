<script>
    // Various Components and functions
    import {A, Avatar, Button, Checkbox, Heading, Input, Label, Modal, P} from "flowbite-svelte";
    import {AddToQueueHTML, RemoveFromQueue} from "../lib/queue";
    import {ToggleSong} from "../lib/song";
    import PlaySolid from "flowbite-svelte-icons/PlaySolid.svelte";
    import PauseSolid from "flowbite-svelte-icons/PauseSolid.svelte";
    import Error from "./errors.svelte"

    // props
    export let GuildId;
    export let token;
    export let host;
    export let queue;
    export let timestamp;


    // variables
    let code = queue;
    let showModal = false;
    let isPlaylist = false;

</script>

<!-- Modal Button -->
<Button class="w-25 absolute right-9 bottom-5" on:click={() => (showModal = true)}>
    Add to Queue
</Button>

<!-- Modal component -->
<Modal title="Add To Queue" bind:open={showModal} autoclose>
    <form id="form">
        <div class="grid grid-rows-2">
            <!-- Song input -->
            <div>
                <Label for="song" class="mb-2">Song Link/Name</Label>
                <Input type="text" id="song" autofocus on:keydown={(e) => {
                    if (e.key === "Enter") {
                        code = AddToQueueHTML(GuildId, token, host);
                        showModal = false;
                    }
                }} 
                required/>
            </div>
            <!-- Checkboxes -->
            <div class="flex gap-3">
                <Checkbox on:click={() => (isPlaylist=!isPlaylist)} id="playlist">Playlist</Checkbox>
                {#if isPlaylist}
                    <Checkbox id="shuffle">Shuffle</Checkbox>
                {:else}
                    <Checkbox disabled id="shuffle">Shuffle</Checkbox>
                {/if}
                <Checkbox id="loop">Loop song</Checkbox>
                <Checkbox id="priority">Priority</Checkbox>
            </div>
        </div>
    </form>
    <!-- Submit Button -->
    <svelte:fragment slot="footer">
        <Button on:click={() => {code = AddToQueueHTML(GuildId, token, host)}}>Add</Button>
    </svelte:fragment>
</Modal>

<!-- Error Toast -->
<Error response={code} />

<!-- Queue -->
{#await queue}
    <P>Fetching Queue</P>
{:then json}
    {#if json.length !== 0 && typeof (json) != "number"}
        <div class="grid grid-row-2 gap-4">
            <!-- First Row -->
            <div class="grid grid-cols-2">
                <!-- Left side of the grid shows the current song -->
                <div class="justify-self-center">
                    <!-- Thumbnail and pause/resume buttons -->
                    <div >
                        <A href={json[0].link}><img alt="thumbnail" src={json[0].thumbnail} href={json.link} class="max-w-3xl"/></A>
                        {#if timestamp != undefined}
                            <P align="center"> {timestamp} / {json[0].duration} </P>
                        {:else}
                            <P align="center"> Fetching... </P>
                        {/if}
                    </div>


                    <div class="mt-5 grid grid-cols-2 gap-2">
                        <!-- Pause -->
                        <div class="justify-self-end">
                            <Button on:click={() => code = ToggleSong(GuildId, token, "pause", host)} disabled={json[0].isPaused}>
                                <PauseSolid/>
                            </Button>
                        </div>
                        
                        <!-- Resume -->
                        <div class="justify-self-start">
                            <Button on:click={() => code = ToggleSong(GuildId, token, "resume", host) }
                                    disabled={!json[0].isPaused}>
                                <PlaySolid/>
                            </Button>
                        </div>
                    </div>
                </div>

                <!-- Right side of the grid renders the actual queue -->
                <div>
                    <Heading tag="h4" class="mb-5" align="center">Queue</Heading>
                    {#each json as song, index}
                        {#if index !== 0}
                            <div class="grid grid-cols-3 justify-items-center mt-3">
                                <Avatar src={song.thumbnail} rounded/>
                                <A href={song.link}>{song.title}</A>
                                <P>{song.duration}</P>
                            </div>
                        {/if}
                    {/each}
                </div>
            </div>

            <!-- Second Row -->
            <div>
                <div class="grid grid-cols-2">
                    <div>
                        <!-- Title and various info -->
                        <div>
                            <A class="mt-0" href={json[0].link}><P weight="bold" class="mt-0 justify-self-start">{json[0].title}</P></A>
                            <P>Requested by {json[0].user}</P>

                            <!-- Skip button -->
                            <Button on:click={() => code = RemoveFromQueue(GuildId, token, false, host)}>Skip song</Button>
                        </div>
                    </div>

                    <!-- Clear button -->
                    <div class="justify-self-start">
                        <Button align="right" on:click={() => code = RemoveFromQueue(GuildId, token, true, host)}>Clear Queue
                        </Button>
                    </div>
                </div>
            </div>
        </div>

    <!-- If there are any errors the 'Error' component is rendered instead -->
    {:else}
        <P>Queue is empty</P>
    {/if}
{/await}



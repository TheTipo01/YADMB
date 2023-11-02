<script>
    // Various Components and functions
    import {A, Avatar, Button, Checkbox, Heading, Input, Label, Modal, P} from "flowbite-svelte";
    import {AddToQueueHTML, GetQueue, RemoveFromQueue} from "../lib/queue";
    import {ToggleSong} from "../lib/song";
    import {
        AddObjectToArray,
        RemoveFirstObjectFromArray,
        ClearArray,
        KeyPressed,
        TogglePause,
        SetPause
    } from "../lib/utilities"
    import PlaySolid from "flowbite-svelte-icons/PlaySolid.svelte";
    import PauseSolid from "flowbite-svelte-icons/PauseSolid.svelte";
    import Error from "./errors.svelte"
    import {onMount} from "svelte";

    // props
    export let GuildId;
    export let token;
    export let host;


    // variables
    let queue = GetQueue(GuildId, token, host);
    let showModal = false;
    let isPlaylist = false;

    // WebSocket
    onMount(async () => {
        let websocket_url = `${host}/ws/${GuildId}?` + new URLSearchParams({"token": token}).toString();
        // If the host is in https, use wss instead of ws
        if (window.location.protocol === "https:") {
            websocket_url = websocket_url.replace("https://", "wss://");
        } else {
            websocket_url = websocket_url.replace("http://", "ws://");
        }

        // Connects to the websocket
        const socket = new WebSocket(websocket_url);
        socket.onmessage = function(e) {
            // Enum for the events
            const Notification = Object.freeze({
                NewSong: 0,
                Skip: 1,
                Pause: 2,
                Resume: 3,
                Clear: 4,
                Finished: 5,
            });
            let signal = JSON.parse(e.data)
            switch(signal.notification) {
                case Notification.NewSong: // New song
                    queue = AddObjectToArray(queue, signal.song);
                    break;
                case Notification.Skip:
                case Notification.Finished: // Song skipped or finished
                    queue = RemoveFirstObjectFromArray(queue);
                    queue = SetPause(queue, false);
                    break;
                case Notification.Clear: // Queue cleared
                    queue = ClearArray(queue);
                    queue = SetPause(queue, false);
                    break;
                case Notification.Resume:
                case Notification.Pause: // Song paused or resumed
                    queue = TogglePause(queue);
                    break;
            }
        }
    });

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
                <Input type="text" id="song" on:keydown={(e) => (KeyPressed(e, "queue", GuildId, token, host))} required/>
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
        <Button on:click={() => AddToQueueHTML(GuildId, token, host)}>Add</Button>
    </svelte:fragment>
</Modal>

<!-- Queue -->
{#await queue} 
<P>Fetching Queue</P>
{:then json}
{#if json.length !== 0 && typeof(json) != "number"}
<div class="grid grid-cols-2 gap-4">
    <!-- Left side of the grid shows the current song -->
    <div >
        <!-- Thumbnail and pause/resume buttons -->
        <img alt="thumbnail" src={json[0].thumbnail} class="justify-self-center"/>
        <div class="mt-5 grid grid-cols-2 gap-2">
            <div class="justify-self-end"><Button on:click={() => ToggleSong(GuildId, token, "pause", host)} disabled={json[0].isPaused}><PauseSolid /></Button></div>
            <div class="justify-self-start"><Button on:click={() => ToggleSong(GuildId, token, "resume", host) } disabled={!json[0].isPaused}><PlaySolid  /></Button></div>
        </div>
        <!-- Title and various info -->
        <a class="mt-5" href={json[0].link}>{json[0].title}</a>
        <P>Requested by {json[0].user}</P>
        <!-- Skip button -->
        <Button on:click={() => RemoveFromQueue(GuildId, token, false, host)}>Skip song</Button>
    </div>

    <!-- Right side of the grid just renders the actual queue -->
    <div>
        <Heading tag="h4" class="mb-5" align="center">Queue</Heading>
        {#each json as song, index}
            {#if index !== 0}
            <div class="grid grid-cols-3 justify-items-center mt-3" >
                <Avatar src={song.thumbnail} rounded/>
                <A href={song.link}>{song.title}</A>
                <P>{song.duration}</P>
            </div>
            {/if}
        {/each}
        <div class="grid grid-rows-1 justify-items-end mt-5">
            <Button align="right" on:click={() => RemoveFromQueue(GuildId, token, true, host)}>Clear Queue</Button>
        </div>
        
    </div>
</div>

<!-- If there are any errors the 'Error' component is rendered instead -->
{:else if typeof(json) === "number"}
    <Error code={json} />
{:else}
    <P>Queue is empty</P>
{/if}
{/await}


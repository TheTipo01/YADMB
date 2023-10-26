<script>
    // Various Components and functions
    import {A, Avatar, Button, Checkbox, Heading, Input, Label, Modal, P} from "flowbite-svelte";
    import {AddToQueue, GetQueue, RemoveFromQueue} from "../lib/queue";
    import {ToggleSong} from "../lib/song";
    import {addObjectToArray, RemoveObjectFromArray, ClearArray} from "../lib/utilities"
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
    let isPaused = false;
    let showModal = false;
    let isPlaylist = false;

    // WebSocket
    onMount(async () => {
        let websocket_url = `${host}/ws/${GuildId}?` + new URLSearchParams({"token": token}).toString();
        // If the host is in https, use wss instead of ws
        if (host.startsWith("https")) {
            websocket_url = 'wss://' + websocket_url;
        } else {
            websocket_url = 'ws://' + websocket_url;
        }

        const socket = new WebSocket(websocket_url);
        socket.onmessage = function(e) {
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
                    queue = addObjectToArray(queue, signal.song);
                    break;
                case Notification.Skip:
                case Notification.Finished:
                    queue = RemoveObjectFromArray(queue);
                    break;
                case Notification.Clear:
                    queue = ClearArray(queue);
                    break;
                case Notification.Pause:
                    isPaused = true;
                    break;
                case Notification.Resume:
                    isPaused = false;
                    break;
            }
        }

        return () => socket.close();
    });


</script>

<Button class="w-25 absolute right-9 bottom-5" on:click={() => (showModal = true)}>
    Add to Queue
</Button>

<Modal title="Add To Queue" bind:open={showModal} autoclose>
    <form id="form">
        <div class="grid grid-rows-2">
            <div>
                <Label for="song" class="mb-2">Song Link/Name</Label>
                <Input type="text" id="song" required/>
            </div>
            <div class="grid grid-cols-4">
                <Checkbox on:click={() => (isPlaylist=!isPlaylist)} id="playlist">Playlist</Checkbox>
                <Checkbox disabled={!isPlaylist} id="shuffle">Shuffle</Checkbox>
                <Checkbox id="loop">Loop song</Checkbox>
                <Checkbox id="priority">Priority</Checkbox>
            </div>
        </div>
    </form>
    <svelte:fragment slot="footer">
        <Button on:click={() => AddToQueue(GuildId, token, host)}>Add</Button>
    </svelte:fragment>
</Modal>


{#await queue} 
<P>Fetching Queue</P>
{:then json}
{#if json.length !== 0 && typeof(json) != "number"}
<div class="grid grid-cols-2 gap-4">
    <div >
        <img alt="thumbnail" src={json[0].thumbnail} class="justify-self-center"/>
        <div class="mt-5 grid grid-cols-2 gap-2">
            {#if json[0].isPaused}
                <div class="justify-self-end"><Button on:click={() => ToggleSong(GuildId, token, "pause", host)} disabled={!isPaused}><PauseSolid /></Button></div>
                <div class="justify-self-start"><Button on:click={() => ToggleSong(GuildId, token, "resume", host) } disabled={isPaused}><PlaySolid  /></Button></div>
            {:else}
                <div class="justify-self-end"><Button on:click={() => ToggleSong(GuildId, token, "pause", host)} disabled={isPaused}><PauseSolid /></Button></div>
                <div class="justify-self-start"><Button on:click={() => ToggleSong(GuildId, token, "resume", host) } disabled={!isPaused}><PlaySolid  /></Button></div>
            {/if}
        </div>
        <a class="mt-5" href={json[0].link}>{json[0].title}</a>
        <P>Requested by {json[0].user}</P>
        <Button on:click={() => RemoveFromQueue(GuildId, token, false, host)}>Skip song</Button>
    </div>
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
{:else if typeof(json) === "number"}
    <Error code={json} />
{:else}
    <P>Queue is empty</P>
{/if}
{/await}


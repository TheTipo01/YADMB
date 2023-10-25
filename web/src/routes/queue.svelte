<script>
    // Various Components and functions
    import {Button} from "flowbite-svelte";
    import {Avatar} from "flowbite-svelte";
    import {P, Heading, A} from "flowbite-svelte";
    import {Modal} from "flowbite-svelte";
    import {Input, Label, Checkbox} from "flowbite-svelte";
	import { GetQueue, AddToQueue, RemoveFromQueue } from "../lib/queue";
    import { ToggleSong } from "../lib/song";
    import PlaySolid from "flowbite-svelte-icons/PlaySolid.svelte";
    import PauseSolid from "flowbite-svelte-icons/PauseSolid.svelte";
    import Error from "./errors.svelte"
	import { onMount } from "svelte";

    // props
    export let GuildId;
    export let token;

    let queue = GetQueue(GuildId, token);
    let isPaused = false;

    // Function to add an object to the array 
    function addObjectToArray(promise, objectToAdd) {
    return promise.then((array) => {
            const newArray = [...array, ...objectToAdd];
            return newArray;
        });
    }

    // Function to remove the first object from the array 
    function RemoveObjectFromArray(promise) {
    return promise.then((array) => {
            array.shift();
            return array;
        });
    }

    function ClearArray(promise) {
        return promise.then((array) => {
            return []
        })
    }

    onMount(async () => {
        let websocket_url = `wss://gerry.thetipo.rocks/ws/${GuildId}?` + new URLSearchParams({"token": token}).toString();
        const socket = new WebSocket(websocket_url);
        socket.onmessage = function(e) {
            const Notification = Object.freeze({
                NewSong: 0,
                Skip: 1,
                Pause: 2,
                Resume: 3,
                Clear: 4,
            });
            let signal = JSON.parse(e.data)
            switch(signal.notification) {
                case Notification.NewSong: // New song
                    queue = addObjectToArray(queue, signal.song);
                    break;
                case Notification.Skip:
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
    });
    
    let showModal = false;
    let isPlaylist = false;


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
        <Button on:click={() => AddToQueue(GuildId, token)}>Add</Button>
    </svelte:fragment>
</Modal>


{#await queue} 
<P>Fetching Queue</P>
{:then json}
{#if json.length != 0 && typeof(json) != "number"}
<div class="grid grid-cols-2 gap-4">
    <div >
        <img alt="thumbnail" src={json[0].thumbnail} class="justify-self-center"/>
        <div class="mt-5 grid grid-cols-2 gap-2">
            <div class="justify-self-end"><Button on:click={() => ToggleSong(GuildId, token, "pause")} disabled={isPaused}><PauseSolid /></Button></div>
            <div class="justify-self-start"><Button on:click={() => ToggleSong(GuildId, token, "resume") } disabled={!isPaused}><PlaySolid  /></Button></div>
        </div>
        <P class="mt-5">{json[0].title}</P>
        <P>Requested by {json[0].user}</P>
        <Button on:click={() => RemoveFromQueue(GuildId, token)}>Skip song</Button>
    </div>
    <div>
        <Heading tag="h4" class="mb-5" align="center">Queue</Heading>
        {#each json as song, index}
            {#if index != 0}
            <div class="grid grid-cols-3 justify-items-center mt-3" >
                <Avatar src={song.thumbnail} rounded/>
                <A href={song.link}>{song.title}</A>
                <P>{song.duration}</P>
            </div>
            {/if}
        {/each}
        <div class="grid grid-rows-1 justify-items-end mt-5">
            <Button align="right" on:click={() => RemoveFromQueue(GuildId, token, true)}>Clear Queue</Button>
        </div>
        
    </div>
</div>
{:else if typeof(json) === "number"}
    <Error code={json} />
{:else}
    <P>Queue is empty</P>
{/if}
{/await}


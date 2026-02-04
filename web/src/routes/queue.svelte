<script>
    import {A, Avatar, Button, Checkbox, Heading, Input, Label, Modal, P} from "flowbite-svelte";
    import {AddToQueueHTML, RemoveFromQueue} from "../lib/queue";
    import {ToggleSong} from "../lib/song";
    import {PauseSolid, PlaySolid} from "flowbite-svelte-icons";
    import Error from "./errors.svelte"

    /** @type {{GuildId: any, token: any, host: any, queue: any, timestamp: any}} */
    let {
        GuildId,
        token,
        host,
        queue,
        timestamp
    } = $props();

    let code = $state(null);
    let showModal = $state(false);
    let isPlaylist = $state(false);

    // Create reactive state for the queue data
    let queueData = $state(null);
    let isLoading = $state(true);

    // Unwrap the promise prop into local state
    $effect(() => {
        isLoading = true;
        Promise.resolve(queue).then(data => {
            queueData = data;
            isLoading = false;
        });
    });

    // Helper function to handle Toggle and update UI immediately
    async function handleToggle(action) {
        // Optimistic Update: Update the UI immediately so the button gets disabled
        if (queueData && queueData.length > 0) {
            queueData[0].isPaused = (action === "pause");
        }

        // Perform the API call
        code = await ToggleSong(GuildId, token, action, host);
    }
</script>

<!-- Modal Button -->
<Button class="w-25 absolute right-9 bottom-5" onclick={() => (showModal = true)}>
    Add to Queue
</Button>

<!-- Modal component -->
<Modal title="Add To Queue" bind:open={showModal} autoclose>
    <form id="form">
        <div class="grid grid-rows-2">
            <!-- Song input -->
            <div>
                <Label for="song" class="mb-2">Song Link/Name</Label>
                <!-- Updated on:keydown to onkeydown for Svelte 5 -->
                <Input type="text" id="song" autofocus onkeydown={(e) => {
                    if (e.key === "Enter") {
                        code = AddToQueueHTML(GuildId, token, host);
                        showModal = false;
                    }
                }}
                       required/>
            </div>
            <!-- Checkboxes -->
            <div class="flex gap-3">
                <Checkbox onclick={() => (isPlaylist=!isPlaylist)} id="playlist">Playlist</Checkbox>
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
	{#snippet footer()}
        <Button onclick={() => {code = AddToQueueHTML(GuildId, token, host)}}>Add</Button>
	{/snippet}
</Modal>

<!-- Error Toast -->
<Error response={code} />

<!-- Queue Display -->
{#if isLoading}
    <P>Fetching Queue</P>
{:else}
    {#if queueData && queueData.length !== 0 && typeof (queueData) != "number"}
        <div class="grid grid-row-2 gap-4">
            <!-- First Row -->
            <div class="grid grid-cols-2">
                <!-- Left side of the grid shows the current song -->
                <div class="justify-self-center">
                    <!-- Thumbnail and pause/resume buttons -->
                    <div>
                        <A href={queueData[0].link}><img alt="thumbnail" src={queueData[0].thumbnail} class="max-w-3xl"/></A>
                        {#if timestamp !== undefined}
                            <P align="center"> {timestamp} / {queueData[0].duration} </P>
                        {:else}
                            <P align="center"> Fetching... </P>
                        {/if}
                    </div>

                    <div class="mt-5 grid grid-cols-2 gap-2">
                        <!-- Pause -->
                        <div class="justify-self-end">
                            <Button onclick={() => handleToggle("pause")} disabled={queueData[0].isPaused}>
                                <PauseSolid/>
                            </Button>
                        </div>

                        <!-- Resume -->
                        <div class="justify-self-start">
                            <Button onclick={() => handleToggle("resume")} disabled={!queueData[0].isPaused}>
                                <PlaySolid/>
                            </Button>
                        </div>
                    </div>
                </div>

                <!-- Right side of the grid renders the actual queue -->
                <div>
                    <Heading tag="h4" class="mb-5" align="center">Queue</Heading>
                    {#each queueData as song, index}
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
                            <A class="mt-0" href={queueData[0].link}><P weight="bold" class="mt-0 justify-self-start">{queueData[0].title}</P></A>
                            <P>Requested by {queueData[0].user}</P>
                            <Button onclick={() => code = RemoveFromQueue(GuildId, token, false, host)}>Skip song</Button>
                        </div>
                    </div>

                    <!-- Clear button -->
                    <div class="justify-self-start">
                        <Button align="right" onclick={() => code = RemoveFromQueue(GuildId, token, true, host)}>Clear Queue</Button>
                    </div>
                </div>
            </div>
        </div>

    <!-- If there are any errors the 'Error' component is rendered instead -->
    {:else}
        <P>Queue is empty</P>
    {/if}
{/if}


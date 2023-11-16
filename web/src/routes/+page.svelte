<script>
    import Queue from "./queue.svelte";
    import Favorites from "./favorites.svelte"
    import {Avatar, TabItem, Tabs} from "flowbite-svelte"
    import StarSolid from "flowbite-svelte-icons/StarSolid.svelte"
    import ListMusicSolid from "flowbite-svelte-icons/ListMusicSolid.svelte"
    import {
        AddObjectToArray,
        ClearArray,
        RemoveFirstObjectFromArray,
        SetPause,
        TogglePause,
        GetGuildID,
        GetToken,
        GetHost,
        GetPauseStatus,
        GetFrames,
        GetTime, AddObjectToArrayAsSecond,
    } from "../lib/utilities"
    import {onMount} from "svelte";
    import {GetQueue} from "$lib/queue"
    import logo from "../assets/logo_yadmb.png"

    // variables
    const FrameSeconds = 50.00067787;
    let GuildId = '';
    let token = '';
    let host = '';
    let activetab = "queue";
    let playing;
    let queue = null;
    let seconds = 0;
    let timestamp;

    onMount(async () => {
        GuildId = GetGuildID();
        token = GetToken();
        host = GetHost();
        queue = GetQueue(GuildId, token, host);
        playing = await GetPauseStatus(queue);

        // Timestamp
        if(seconds === 0) seconds = await GetFrames(queue) / FrameSeconds;
        timestamp = GetTime(Math.floor(seconds));

        let websocket_url = `${host}/ws/${GuildId}?` + new URLSearchParams({"token": token}).toString();
        // If the host is in https, use wss instead of ws
        if (window.location.protocol === "https:") {
            websocket_url = websocket_url.replace("https://", "wss://");
        } else {
            websocket_url = websocket_url.replace("http://", "ws://");
        }

        // Connects to the websocket
        function start() {
            const socket = new WebSocket(websocket_url);
            socket.onmessage = function (e) {
                // Enum for the events
                const Notification = Object.freeze({
                    NewSong: 0,
                    Skip: 1,
                    Pause: 2,
                    Resume: 3,
                    Clear: 4,
                    Finished: 5,
                    Playing: 6,
                    NewSongPriority: 7,
                });
                let signal = JSON.parse(e.data)
                switch (signal.notification) {
                    case Notification.NewSong: // New song
                        queue = AddObjectToArray(queue, signal.song);
                        break;
                    case Notification.Skip:
                    case Notification.Finished: // Song skipped or finished
                        queue = RemoveFirstObjectFromArray(queue);
                        queue = SetPause(queue, false);
                        playing = false;
                        seconds = 0;
                        break;
                    case Notification.Clear: // Queue cleared
                        queue = ClearArray(queue);
                        queue = SetPause(queue, false);
                        seconds = 0;
                        break;
                    case Notification.Resume:
                    case Notification.Pause: // Song paused or resumed
                        queue = TogglePause(queue);
                        playing = !playing;
                        break;
                    case Notification.Playing:
                        playing = true;
                        break;
                    case Notification.NewSongPriority:
                        queue = AddObjectToArrayAsSecond(queue, signal.song);
                        break;
                }
            }

            socket.onclose = function(e) {
                if(!socket || socket.readyState === WebSocket.CLOSED) start();
            }
            socket.onerror = socket.onclose;

        }
        
        start();
    });
    setInterval(function() {if(playing){seconds += 1; timestamp = GetTime(seconds)}}, 1000)

</script>
<div class="flex justify-center">
    <Avatar src="{logo}" size="xl" class="mt-1"/>
</div>


<Tabs style="underline">
    <TabItem open active={activetab === "queue"} on:click={() => (activetab = "queue")}>
        <div slot="title" class="flex items-center gap-2">
            <ListMusicSolid/>
            Queue
        </div>
        {#if activetab === "queue"}
            {#if GuildId !== '' && token !== '' && host !== '' && queue != null && timestamp != undefined}
                <Queue GuildId={GuildId} token={token} host={host} queue={queue} timestamp = {timestamp}/>
            {/if}
        {/if}
    </TabItem>
    <TabItem active={activetab === "favorites"} on:click={() => (activetab = "favorites")}>
        <div slot="title" class="flex items-center gap-2">
            <StarSolid/>
            Favorites
        </div>
        {#if activetab === "favorites"}
            {#if token !== '' && host !== ''}
                <Favorites GuildId={GuildId} token={token} host={host}/>
            {/if}
        {/if}
    </TabItem>
</Tabs>
<div>
</div>


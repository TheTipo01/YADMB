<script>
	import Queue from './queue.svelte';
	import Favorites from './favorites.svelte';
	import { Avatar, TabItem, Tabs } from 'flowbite-svelte';
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
		GetTime,
		AddObjectToArrayAsSecond
	} from '../lib/utilities.js';
	import { onMount } from 'svelte';
	import { GetQueue } from '../lib/queue.js';
	import logo from '../assets/logo_yadmb.png';
	import { ListMusicSolid, StarSolid } from 'flowbite-svelte-icons';

	// variables
	const FrameSeconds = 50.00067787;
	let GuildId = $state('');
	let token = $state('');
	let host = $state('');
	let activetab = $state('queue');
	let playing;
	let queue = $state(null);
	let seconds = 0;
	let timestamp = $state();

	onMount(async () => {
		GuildId = GetGuildID();
		token = GetToken();
		host = GetHost();
		queue = GetQueue(GuildId, token, host);
		playing = await GetPauseStatus(queue);

		// Timestamp
		if (seconds === 0) seconds = (await GetFrames(queue)) / FrameSeconds;
		timestamp = GetTime(Math.floor(seconds));

		let websocket_url = `${host}/ws/${GuildId}?` + new URLSearchParams({ token: token }).toString();
		// If the host is in https, use wss instead of ws
		if (window.location.protocol === 'https:') {
			websocket_url = websocket_url.replace('https://', 'wss://');
		} else {
			websocket_url = websocket_url.replace('http://', 'ws://');
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
					LoopFinished: 8
				});
				let signal = JSON.parse(e.data);
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
					case Notification.LoopFinished:
						seconds = 0;
						break;
				}
			};

			socket.onclose = function (e) {
				if (!socket || socket.readyState === WebSocket.CLOSED) start();
			};
			socket.onerror = socket.onclose;
		}

		start();
	});
	setInterval(function () {
		if (playing) {
			seconds += 1;
			timestamp = GetTime(seconds);
		}
	}, 1000);
</script>

<div class="flex justify-center">
	<Avatar src={logo} size="xl" class="mt-1 " />
</div>

<Tabs style="underline">
	<TabItem title="Queue" open active={activetab === 'queue'} onclick={() => (activetab = 'queue')}>
		{#snippet titleSlot()}
			<div class="flex items-center gap-2">
				<ListMusicSolid />
				Queue
			</div>
		{/snippet}
		{#if activetab === 'queue'}
			{#if GuildId !== '' && token !== '' && host !== '' && queue != null && timestamp != undefined}
				<Queue {GuildId} {token} {host} {queue} {timestamp} />
			{/if}
		{/if}
	</TabItem>
	<TabItem
		title="Favorites"
		active={activetab === 'favorites'}
		onclick={() => (activetab = 'favorites')}
	>
		{#snippet titleSlot()}
			<div class="flex items-center gap-2">
				<StarSolid />
				Favorites
			</div>
		{/snippet}
		{#if activetab === 'favorites'}
			{#if token !== '' && host !== ''}
				<Favorites {GuildId} {token} {host} />
			{/if}
		{/if}
	</TabItem>
</Tabs>
<div></div>

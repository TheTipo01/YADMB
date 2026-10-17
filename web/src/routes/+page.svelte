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
	import { onDestroy, onMount } from 'svelte';
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
	let started = $state(false);
	let queue = $state(null);
	let seconds = 0;
	let timestamp = $state();
	let socket;
	let reconnectTimer;
	let clockTimer;
	let stopped = false;

	onMount(async () => {
		clockTimer = setInterval(function () {
			if (playing) {
				seconds += 1;
				timestamp = GetTime(seconds);
			}
		}, 1000);

		GuildId = GetGuildID();
		token = GetToken();
		host = GetHost();
		queue = GetQueue(GuildId, token, host);

		const frames = await GetFrames(queue);
		const paused = await GetPauseStatus(queue);
		seconds = frames / FrameSeconds;
		started = frames > 0;
		playing = paused && started;

		timestamp = GetTime(Math.floor(seconds));

		let websocket_url = `${host}/ws/${GuildId}?` + new URLSearchParams({ token: token }).toString();
		// If the host is in https, use wss instead of ws
		if (window.location.protocol === 'https:') {
			websocket_url = websocket_url.replace('https://', 'wss://');
		} else {
			websocket_url = websocket_url.replace('http://', 'ws://');
		}

		// Connects to the websocket, reconnecting with exponential backoff.
		let reconnectDelay = 1000;

		function connect() {
			if (stopped) {
				return;
			}

			socket = new WebSocket(websocket_url);

			socket.onopen = function () {
				reconnectDelay = 1000;
			};

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
					LoopFinished: 8,
					Sync: 9
				});

				let signal;
				try {
					signal = JSON.parse(e.data);
				} catch (error) {
					console.error('Invalid websocket message', error);
					return;
				}

				const songs = Array.isArray(signal.song) ? signal.song : [];

				switch (signal.notification) {
					case Notification.NewSong: // New song
						queue = AddObjectToArray(queue, songs);
						break;
					case Notification.Skip:
					case Notification.Finished: // Song skipped or finished
						queue = RemoveFirstObjectFromArray(queue);
						queue = SetPause(queue, false);
						started = false;
						playing = false;
						seconds = 0;
						timestamp = GetTime(0);
						break;
					case Notification.Clear: // Queue cleared
						queue = ClearArray(queue);
						queue = SetPause(queue, false);
						started = false;
						playing = false;
						seconds = 0;
						timestamp = GetTime(0);
						break;
					case Notification.Resume:
					case Notification.Pause: // Song paused or resumed
						queue = TogglePause(queue);
						playing = signal.notification === Notification.Resume ? started : false;
						break;
					case Notification.Playing:
						// The bot is preparing the song; the clock starts on Started.
						started = false;
						playing = false;
						seconds = 0;
						timestamp = GetTime(0);
						break;
					case Notification.Sync: {
						const frames = typeof signal.frames === 'number' ? signal.frames : 0;
						seconds = frames / FrameSeconds;
						timestamp = GetTime(Math.floor(seconds));

						if (typeof signal.playing === 'boolean') {
							started = signal.playing;
							playing = signal.playing;
						}
						break;
					}
					case Notification.NewSongPriority:
						queue = AddObjectToArrayAsSecond(queue, songs);
						break;
					case Notification.LoopFinished:
						seconds = 0;
						timestamp = GetTime(0);
						break;
				}
			};

			socket.onclose = function () {
				if (stopped) {
					return;
				}

				clearTimeout(reconnectTimer);
				reconnectTimer = setTimeout(connect, reconnectDelay);
				reconnectDelay = Math.min(reconnectDelay * 2, 30000);
			};

			socket.onerror = function () {
				socket?.close();
			};
		}

		connect();
	});

	onDestroy(() => {
		stopped = true;
		clearTimeout(reconnectTimer);
		clearInterval(clockTimer);
		socket?.close();
	});
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

<script>
	import { A, Avatar, Button, Checkbox, Heading, Input, Label, Modal, P } from 'flowbite-svelte';
	import { AddToQueueHTML, RemoveFromQueue } from '../lib/queue';
	import { ToggleSong } from '../lib/song';
	import { PauseSolid, PlaySolid } from 'flowbite-svelte-icons';
	import Error from './errors.svelte';
	import { Response } from '../lib/error';
	import { SafeHref } from '../lib/utilities';

	/** @type {{GuildId: string, token: string, host: string, queue: any, timestamp: string | undefined}} */
	let { GuildId, token, host, queue, timestamp } = $props();

	/** @typedef {{ title: string, link: string, thumbnail: string, duration: string, user: string, frames: number, isPaused: boolean }} QueueSong */

	let code = $state(/** @type {number | Promise<number> | null} */ (null));
	let showModal = $state(false);
	let isPlaylist = $state(false);

	// Create reactive state for the queue data
	let queueData = $state(/** @type {QueueSong[] | number | null} */ (null));
	let isLoading = $state(true);

	// Unwrap the promise prop into local state
	$effect(() => {
		isLoading = true;
		Promise.resolve(queue).then((data) => {
			queueData = data;
			isLoading = false;
		});
	});

	// Helper function to handle Toggle and update UI immediately
	/** @param {'pause' | 'resume'} action */
	async function handleToggle(action) {
		if (!Array.isArray(queueData) || queueData.length === 0) {
			return;
		}

		const previous = queueData[0].isPaused;

		// Optimistic Update: Update the UI immediately so the button gets disabled
		queueData[0].isPaused = action === 'pause';

		// Perform the API call
		code = await ToggleSong(GuildId, token, action, host);

		if (code !== Response.SUCCESS) {
			queueData[0].isPaused = previous;
		}
	}
</script>

<!-- Modal Button -->
<Button class="w-25 absolute right-9 bottom-5" onclick={() => (showModal = true)}>
	<P align="center">Add to Queue</P>
</Button>

<!-- Modal component -->
<Modal title="Add To Queue" bind:open={showModal} autoclose>
	<form id="form">
		<div class="grid grid-rows-2">
			<!-- Song input -->
			<div>
				<Label for="song" class="mb-2">Song Link/Name</Label>
				<!-- Updated on:keydown to onkeydown for Svelte 5 -->
				<Input
					type="text"
					id="song"
					autofocus
					onkeydown={(e) => {
						if (e.key === 'Enter') {
							code = AddToQueueHTML(GuildId, token, host);
							showModal = false;
						}
					}}
					required
				/>
			</div>
			<!-- Checkboxes -->
			<div class="flex gap-3">
				<Checkbox onclick={() => (isPlaylist = !isPlaylist)} id="playlist">Playlist</Checkbox>
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
		<Button
			onclick={() => {
				code = AddToQueueHTML(GuildId, token, host);
			}}>Add</Button
		>
	{/snippet}
</Modal>

<!-- Error Toast -->
<Error response={code} />

<!-- Queue Display -->
{#if isLoading}
	<P>Fetching Queue</P>
{:else if typeof queueData === 'number'}
	<Error response={queueData} />
{:else if queueData && queueData.length !== 0 && typeof queueData != 'number'}
	<div class="grid grid-row-2 gap-4">
		<!-- First Row -->
		<div class="grid grid-cols-2">
			<!-- Left side of the grid shows the current song -->
			<div class="justify-self-center">
				<!-- Thumbnail and pause/resume buttons -->
				<div>
					<A href={SafeHref(queueData[0].link)}
						><img alt="thumbnail" src={queueData[0].thumbnail} class="max-w-3xl" /></A
					>
					{#if timestamp !== undefined}
						<P align="center">{timestamp} / {queueData[0].duration}</P>
					{:else}
						<P align="center">Fetching...</P>
					{/if}
				</div>

				<div class="mt-5 grid grid-cols-2 gap-2">
					<!-- Pause -->
					<div class="justify-self-end">
						<Button onclick={() => handleToggle('pause')} disabled={queueData[0].isPaused}>
							<PauseSolid />
						</Button>
					</div>

					<!-- Resume -->
					<div class="justify-self-start">
						<Button onclick={() => handleToggle('resume')} disabled={!queueData[0].isPaused}>
							<PlaySolid />
						</Button>
					</div>
				</div>
			</div>

			<!-- Right side of the grid renders the actual queue -->
			<div>
				<Heading tag="h4" class="mb-5 text-center">Queue</Heading>
				{#each queueData as song, index (index)}
					{#if index !== 0}
						<div class="grid grid-cols-3 justify-items-center mt-3">
							<Avatar src={song.thumbnail} cornerStyle="rounded" />
							<A href={SafeHref(song.link)}>{song.title}</A>
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
						<A class="mt-0" href={SafeHref(queueData[0].link)}
							><P weight="bold" class="mt-0 justify-self-start">{queueData[0].title}</P></A
						>
						<P>Requested by {queueData[0].user}</P>
						<Button onclick={() => (code = RemoveFromQueue(GuildId, token, false, host))}
							><P align="center">Skip song</P></Button
						>
					</div>
				</div>

				<!-- Clear button -->
				<div class="justify-self-start">
					<Button onclick={() => (code = RemoveFromQueue(GuildId, token, true, host))}
						><P align="center">Clear Queue</P></Button
					>
				</div>
			</div>
		</div>
	</div>

	<!-- If there are any errors the 'Error' component is rendered instead -->
{:else}
	<P>Queue is empty</P>
{/if}

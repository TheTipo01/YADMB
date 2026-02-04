<script>
	import { Toast } from 'flowbite-svelte';
	import { get_message, Response } from '../lib/error.js';
	import { CheckCircleSolid, CloseCircleSolid } from 'flowbite-svelte-icons';

	// autohide variable
	let open = $state(true);
	setTimeout(() => {
		open = false;
	}, 7000);

	// props
	/** @type {{response: any}} */
	let { response } = $props();
</script>

{#await response}
	<Toast position="bottom-left">Processing...</Toast>
{:then code}
	{#if typeof code === 'number'}
		{#if code === Response.SUCCESS}
			<Toast color="green" position="bottom-left" bind:open>
				{#snippet icon()}
					<CheckCircleSolid class="w-5 h-5" />
					<span class="sr-only">Check icon</span>
				{/snippet}
				{get_message(code)}
			</Toast>
		{:else}
			<Toast color="red" position="bottom-left" bind:open>
				{#snippet icon()}
					<CloseCircleSolid class="w-5 h-5" />
					<span class="sr-only">Error icon</span>
				{/snippet}
				{get_message(code)}
			</Toast>
		{/if}
	{/if}
{/await}

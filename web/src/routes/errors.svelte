<script>
	import { Toast } from 'flowbite-svelte';
	import { get_message, Response } from '../lib/error.js';
	import { CheckCircleSolid, CloseCircleSolid } from 'flowbite-svelte-icons';

	// props
	/** @type {{response: any}} */
	let { response } = $props();

	let open = $state(false);
	let loading = $state(true);
	let code = $state(/** @type {number | null} */ (null));

	$effect(() => {
		loading = true;
		open = false;
		/** @type {ReturnType<typeof setTimeout> | undefined} */
		let timer;

		Promise.resolve(response).then((result) => {
			code = result;
			loading = false;

			if (typeof result === 'number') {
				open = true;

				timer = setTimeout(() => {
					open = false;
				}, 7000);
			}
		});

		return () => clearTimeout(timer);
	});
</script>

<!-- Render Logic -->
{#if loading}
	<!-- Show processing while waiting for promise -->
	<Toast position="bottom-left">Processing...</Toast>
{:else if open && typeof code === 'number'}
	<!-- Show Result once loaded -->
	{#if code === Response.SUCCESS}
		<Toast color="green" position="bottom-left" bind:toastStatus={open}>
			{#snippet icon()}
				<CheckCircleSolid class="w-5 h-5" />
				<span class="sr-only">Check icon</span>
			{/snippet}
			{get_message(code)}
		</Toast>
	{:else}
		<Toast color="red" position="bottom-left" bind:toastStatus={open}>
			{#snippet icon()}
				<CloseCircleSolid class="w-5 h-5" />
				<span class="sr-only">Error icon</span>
			{/snippet}
			{get_message(code)}
		</Toast>
	{/if}
{/if}

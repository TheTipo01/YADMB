<script>
	import { Accordion, AccordionItem, Button, Input, Label, Modal, P, A } from 'flowbite-svelte';
	import { AddFavorite, GetFavorites, RemoveFavorite } from '../lib/favorites';
	import { AddObjectToArray, RemoveObjectFromArray, GetFolders } from '../lib/utilities';
	import {
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import { TrashBinSolid, PlusOutline } from 'flowbite-svelte-icons';
	import { AddToQueue } from '../lib/queue.js';
	import { Response } from '../lib/error.js';
	import Error from './errors.svelte';

	// props
	/** @type {{GuildId: any, token: any, host: any}} */
	let { GuildId, token, host } = $props();

	// variables
	let favorites = $state(GetFavorites(token, host));
	let code = $state(favorites);
	let showModal = $state(false);
</script>

<!-- Modal Button -->
<Button class="w-25 absolute right-9 bottom-5" onclick={() => (showModal = true)}>
	<P>Add to Favorites</P>
</Button>

<!-- Modal Component -->
<Modal title="Add Favorite" bind:open={showModal} autoclose>
	<form id="form">
		<!-- Form inputs -->
		<div class="grid grid-rows-3">
			<div>
				<Label for="name" class="mb-2">Song Name</Label>
				<Input
					type="text"
					id="name"
					onkeydown={(e) => {
						if (e.key === 'Enter') {
							code = AddFavorite(token, host);
							showModal = false;
						}
					}}
					required
				/>
			</div>
			<div>
				<Label for="link" class="mt-2 mb-2">Song Link</Label>
				<Input
					type="text"
					id="link"
					onkeydown={(e) => {
						if (e.key === 'Enter') {
							code = AddFavorite(token, host);
							showModal = false;
						}
					}}
					required
				/>
			</div>
			<div>
				<Label for="folder" class="mt-2 mb-2">Folder</Label>
				<Input
					type="text"
					id="folder"
					onkeydown={(e) => {
						if (e.key === 'Enter') {
							code = AddFavorite(token, host);
							showModal = false;
						}
					}}
				/>
			</div>
		</div>
	</form>

	<!-- Submit button -->
	{#snippet footer()}
		<Button
			onclick={async () => {
				let result = await AddFavorite(token, host, favorites);
				switch (result) {
					case Response.FAVORITE_TOKEN_ERROR:
						code = Response.FAVORITE_TOKEN_ERROR;
						break;
					case Response.DUPLICATE_ERROR:
						code = Response.DUPLICATE_ERROR;
						break;
					default:
						code = Response.SUCCESS;
						favorites = AddObjectToArray(favorites, result);
				}
			}}
			>Add
		</Button>
	{/snippet}
</Modal>

<Error response={code} />

<!-- Favorites list -->
{#await favorites}
	<P>Loading Favorites</P>
{:then favorite}
	{#if favorite.length !== 0}
		<Accordion>
			{#each GetFolders(favorite) as folder}
				<!-- Folder Header -->
				<AccordionItem>
					{#snippet header()}
						{folder !== '' ? folder : 'No Folder'}
					{/snippet}
					<!-- Folder Content -->
					<Table hoverable shadow>
						<TableHead>
							<TableHeadCell align="center" class="w-1/2">Name</TableHeadCell>
							<TableHeadCell align="center" class="w-2/5">Link</TableHeadCell>
							<TableHeadCell class="w-1/4">
								<span class="sr-only">Add To Queue</span>
							</TableHeadCell>
							<TableHeadCell>
								<span class="sr-only">Remove From Favorites</span>
							</TableHeadCell>
						</TableHead>
						<TableBody>
							{#each favorite as song}
								<TableBodyRow>
									{#if song.folder === folder}
										<TableBodyCell class="w-1/2"><P align="center">{song.name}</P></TableBodyCell>
										<TableBodyCell align="center" class="w-2/5"
											><A href={song.link}>{song.link}</A></TableBodyCell
										>
										<TableBodyCell
											onclick={() =>
												(code = AddToQueue(
													GuildId,
													token,
													host,
													song.link,
													false,
													false,
													false,
													false
												))}
										>
											<div class="flex items-center justify-center cursor-pointer">
												<PlusOutline size="lg" align="center" />
											</div>
											<P>Add to Queue</P>
										</TableBodyCell>
										<TableBodyCell
											onclick={async () => {
												let result = await RemoveFavorite(token, song.name, host);
												switch (result) {
													case Response.FAVORITE_TOKEN_ERROR:
														code = Response.FAVORITE_TOKEN_ERROR;
														break;
													case Response.FAVORITE_NOT_FOUND:
														code = Response.FAVORITE_NOT_FOUND;
														break;
													default:
														code = Response.SUCCESS;
														favorites = RemoveObjectFromArray(favorites, song);
												}
											}}
										>
											<div class="flex items-center justify-center cursor-pointer">
												<TrashBinSolid size="lg" />
											</div>
											<P>Remove</P>
										</TableBodyCell>
									{/if}
								</TableBodyRow>
							{/each}
						</TableBody>
					</Table>
				</AccordionItem>
			{/each}
		</Accordion>
	{:else}
		<P>You have no favorites</P>
	{/if}
{/await}

<script>
    import {Button} from "flowbite-svelte";
    import {Avatar} from "flowbite-svelte"
    import {P, Heading, A} from "flowbite-svelte";
	import { GetQueue, AddToQueue, RemoveFromQueue } from "../lib/queue";
    import PlaySolid from "flowbite-svelte-icons/PlaySolid.svelte"
    import PauseSolid from "flowbite-svelte-icons/PauseSolid.svelte"
    let queue = GetQueue("145618799947677696", "iAADRNu7eEfcHmcT9j9pbUsI4Rff7ihG");
</script>

<Button class="w-25 absolute right-9 bottom-5">
    Add to Queue
</Button>



{#await queue} 
<P>Fetching Queue</P>
{:then json}
<div class="grid grid-cols-2 gap-4">
    <div >
        <img alt="thumbnail" src={json[0].thumbnail} class=""/>
        <div class="mt-5 grid grid-cols-2">
            <div class="justify-self-end"><PauseSolid /></div>
            <div class="justify-self-start"><PlaySolid  /></div>
        </div>
        <P class="mt-5">{json[0].title}</P>
        <P>Requested by {json[0].user}</P>
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
    </div>
</div>
    {/await}

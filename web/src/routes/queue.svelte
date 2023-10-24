<script>
    import {AddToQueue} from '../lib/queue'
    import {Label, Input} from 'flowbite-svelte'
    import {Heading} from 'flowbite-svelte'
    import {Button} from 'flowbite-svelte'
    let result;
    async function check_result() {
        result = await AddToQueue();
    }
</script>

<Heading tag="h3">Add a song to the queue</Heading>
<form>
    <Label for="id" class="block mb-2">Guild ID</Label> 
    <Input type="text" id="id" /> 
    <Label for="link" class="mt-2">Song name or link</Label> 
    <Input type="text" id="song" />
    <Label for="token" class="mt-2">User token</Label> 
    <Input type="text" id="token" /> 
    <Label for="playlist" class="mt-2"> Playlist </Label>
    <Input type="checkbox" id="playlist" />
    <Label for="shuffle" class="mt-2"> Shuffle </Label>
    <Input type="checkbox" id="shuffle" />
    <Label for="loop" class="mt-2"> Loop </Label>
    <Input type="checkbox" id="loop" />
    <Label for="priority" class="mt-2"> Priority </Label>
    <Input type="checkbox" id="priority" />

    <Button class="mt-2" on:click={check_result} >Submit</Button>
</form>
<script>
    import {AddToQueue, RemoveFromQueue, GetQueue} from '../lib/queue'
    import {Label, Input} from 'flowbite-svelte'
    import {Button, Dropdown, DropdownItem } from "flowbite-svelte"
</script>

<div class="grid grid-row-3 gap-4">
    <Button>Add a song to the queue</Button>
    <Dropdown>
        <DropdownItem>
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
            
                <Button class="mt-2" on:click={AddToQueue} >Submit</Button>
            </form>
        </DropdownItem>
    </Dropdown>
    <Button>Skip song</Button>
    <Dropdown>
        <DropdownItem>
            <form>
                <Label for="id" class="block mb-2">Guild ID</Label> 
                <Input type="text" id="id" /> 
                <Label for="clear" class="mt-2">Clear queue?</Label> 
                <Input type="checkbox" id="clear" />

                <Button class="mt-2" on:click={RemoveFromQueue} >Submit</Button>
            </form>
        </DropdownItem>
    </Dropdown>
    <Button>Show queue</Button>
    <Dropdown>
        <DropdownItem>
            <Label for="id" class="block mb-2"> Guild ID</Label>
            <Input type="text" id="id" /> 
            <Label for="token" class="mt-2">User token</Label> 
            <Input type="text" id="token" />
            
            <Button class="mt-2" on:click={GetQueue} >Submit</Button>
        </DropdownItem>
    </Dropdown>

</div>
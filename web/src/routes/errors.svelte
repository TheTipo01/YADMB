<script>
    import {Toast} from "flowbite-svelte";
    import {get_message, Response} from "../lib/error"
    import CheckCircleSolid from "flowbite-svelte-icons/CheckCircleSolid.svelte"
    import CloseCircleSolid from "flowbite-svelte-icons/CloseCircleSolid.svelte"

    // autohide variable
    let open = true;
    setTimeout(() => {open = false}, 7000);

    // props
    export let response;

</script>

{#await response}
    <Toast position="bottom-left">
        Processing...
    </Toast>
{:then code}
    {#if typeof(code) === "number"}
        {#if code === Response.SUCCESS}
        <Toast color="green" position="bottom-left" bind:open>
            <svelte:fragment slot="icon">
                <CheckCircleSolid class="w-5 h-5"/>
                <span class="sr-only">Check icon</span>
            </svelte:fragment>
            {get_message(code)}
        </Toast>
        {:else}
        <Toast color="red" position="bottom-left" bind:open>
            <svelte:fragment slot="icon">
                <CloseCircleSolid class="w-5 h-5"/>
                <span class="sr-only">Error icon</span>
            </svelte:fragment>
            {get_message(code)}
        </Toast>
        {/if}
    {/if}
{/await}

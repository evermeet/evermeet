<script>
    import Avatar from 'svelte-boring-avatars';
    export let calendar;
    export let key = null;
    export let size = 25;
</script>

<div class="{calendar.personal ? 'rounded-full' : 'rounded'} overflow-hidden" style="width: {size}px; height: {size}px;">
    {#if calendar.img}
        <img alt={calendar.handle} src={calendar.img} class="" style="width: {size}px; height: {size}px;" />
    {:else}
        <div class="">
            <Avatar
                {size}
                name={key || calendar.handle || calendar.did}
                square={true}
                variant="bauhaus"
                colors={['#ff79c6', '#ff79c6', '#bd93f9', '#ffb86c', '#414558']}
            />
        </div>
    {/if}
</div>
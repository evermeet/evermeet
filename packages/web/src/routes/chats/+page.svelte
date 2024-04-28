<script>
    import { user } from '$lib/stores'
    
    import Chat from "$lib/../components/Chat.svelte";
    $: item = $user
    let room = 'general'
    
</script>

<div class="page-extra-wide">
    <h1 class="heading1">Chats</h1>

    <Chat {item} {room} />

</div>
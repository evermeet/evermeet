<script>
    import Chat from "$lib/../components/Chat.svelte";

    const { data } = $props()

    const room = $derived(data.room || data.rooms[0])
    const messages = $derived(data.messages)

</script>

<div class="page-extra-wide">
    <h1 class="heading1">Chats</h1>

    <div class="flex gap-4">
        <ul class="menu bg-base-100 min-w-36 shrink-0">
            {#each data.rooms as r}
                <h2 class="menu-title text-xs">{r.repo}</h2>
                <li>
                    <a href="/chats?room={r.slug}@{r.repo}" class:active={room.id === r.id}>
                        <div>#{r.slug}</div>
                    </a>
                </li>
            {/each}
        </ul>
        <div class="grow">
            <div class="font-mono breadcrumbs mb-2">
                <ul>
                  <li><a>{room.repo}</a></li> 
                  <li>#{room.slug}</li>
                </ul>
              </div>
            <Chat item={data.user} {room} chatData={messages}/>
        </div>
    </div>
</div>
<script>
    export let item;
    export let preview;

    const c = item;
</script>

<a href="/{c.slug}">
    <div class="itembox itembox-hover h-full">
        <div class="flex">
            <div class="w-12 h-12 mb-2 grow">
                <img
                    class="{c.personal ? "rounded-full" : "rounded-lg"} w-12 h-12"
                    src={c.img} />
            </div>
            {#if preview}
                <button class="btn btn-sm">Subscribe</button>
            {/if}
        </div>
        <div class="text-lg font-semibold">{c.name}</div>
        {#if preview}
            <div class="mt-1 text-sm text-neutral-content">{c.description}</div>
        {:else}
            <div class="text-sm mt-1 text-neutral-content">{c.subs} subscribers</div>
            <div class="mt-4 text-sm text-neutral-content">
                {#if c.personal}
                    Personal
                {:else}
                    1 admin
                {/if}
            </div>
        {/if}
    </div>
</a>
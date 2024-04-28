<script>
    export let item;

    function formatRef([key, url]) {
        const map = {
            web: {
                re: /^https?:\/\/([^/]+\/)?([^/]+)/,
                handle: (match) => match[2],
            },
            twitter: {
                re: /twitter.com\/(.*?)\/?$/,
                handle: (match) => `@${match[1]}`,
            },
            matrix: {
                re: /matrix\.to\/#\/(.*)/,
                handle: (match) => match[1],
            },
            github: {
                re: /github.com\/([^/]+)/,
                handle: (match) => `@${match[1]}`,
            },
        };

        const def = map[key];
        if (def) {
            const match = url.match(def.re);
            if (match) {
                return [key, url, def.handle(match)];
            }
        }

        return [key, url];
    }
</script>

<div class="flex gap-4 flex-wrap items-center">
    {#each Object.entries(item.refs).map(formatRef) as [key, url, handle]}
        <div class="flex gap-1 items-center">
            <div class="badge badge-sm badge-ghost text-neutral-content/50">{key}</div>
            <div class="text-sm"><a href={url} class="hover:underline">{handle || url}</a></div>
        </div>
    {/each}
</div>

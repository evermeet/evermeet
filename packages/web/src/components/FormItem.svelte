<script>
    import countries from 'i18n-iso-countries';
    import enLocale from 'i18n-iso-countries/langs/en.json';
    import UserAvatar from './UserAvatar.svelte';
    import { blobUrl } from '$lib/api';
    import { Photo, Trash } from 'svelte-heros-v2';

    countries.registerLocale(enLocale);
    const countriesAll = countries.getNames("en", {select: "official"});

    export let row;
    export let config = {};
    export let formData;

    let image = {};
    function image_onSelected (e) {
        const img = e.target.files[0];
        const rData = new FileReader();
        rData.readAsArrayBuffer(img);
        rData.onload = re => {
            image.data = {
                body: re.target.result,
                size: img.size,
                mimeType: img.type,
                name: img.name,
            }
            $formData[row.column] = image
        }
        const rDataUrl = new FileReader();
        rDataUrl.readAsDataURL(img);
        rDataUrl.onload = re => {
            image.dataUrl = re.target.result
        }
    }
    function image_onRemove() {
        image = {}
        $formData[row.column] = undefined
    }

</script>

<div>
    {#if row.title}
        <label for={row.title}>{row.title}</label>
    {/if}
    {#if row.view === 'slug'}
        <div class="join">
            <input value="{config.domain}/" class="input input-disabled w-32" />
            <input id={row.title} type="text" placeholder={row.placeholder} class="input input-bordered w-64" bind:value={$formData[row.column]} />
        </div>
    {/if}
    {#if row.view === 'country'}
        <select id={row.title} class="select {config.bordered && 'select-bordered'} w-full max-w-xs" value={$formData[row.column]}>
            <option value="">(not specified)</option>
            {#each Object.keys(countriesAll) as country}
                <option value={country.toLowerCase()}>{countriesAll[country]}</option>
            {/each}
        </select>
    {/if}
    {#if row.view === 'textarea'}
        <textarea id={row.title} class="textarea {config.bordered && "textarea-bordered"} w-full font-mono text-sm" placeholder={row.placeholder} bind:value={$formData[row.column]}></textarea>
    {/if}
    {#if row.view === 'string-disabled'}
        <input id={row.title} type="text" class="input {config.bordered && "input-bordered"} input-disabled w-96" value={row.value} />
    {/if}
    {#if row.view === 'image'}
        <div class="flex gap-4 mt-4">
            <div on:click={() => document.getElementById(row.title).click()}>
                <UserAvatar user={{ did: config.user.did, avatarBlob: $formData[row.column] && config.user.avatarBlob }} data={image?.dataUrl} size="100" />
            </div>
            <div>
                <div>
                    <div>Type: {$formData[row.column] ? 'Custom' : 'Automatically generated'}</div>
                    {#if $formData[row.column] && typeof($formData[row.column]) === 'object'}
                        <div class="mt-2 text-sm text-base-content/50">
                            <div>Mime Type: {$formData[row.column]?.data?.mimeType}</div>
                            <div>Size: {$formData[row.column]?.data?.size}</div>
                        </div>
                    {:else if $formData[row.column] && typeof($formData[row.column]) === 'string'}
                        <div class="mt-2 text-base-content/50">
                            CID: <a class="hover:underline" target="_blank" href={blobUrl(config.user.did, $formData[row.column])}>{$formData[row.column]}</a>
                        </div>
                    {/if}
                </div>
                <div class="mt-2">
                    {#if $formData[row.column]}
                        <button type="button" class="btn btn-sm" on:click|preventDefault={image_onRemove}><Trash class="outline-none" size="15" /> Remove image</button>
                    {:else}
                        <button type="button" class="btn btn-sm btn-primary" on:click|preventDefault={() => document.getElementById(row.title).click()}>Add image</button>
                    {/if}
                </div>
            </div>
        </div>
        <input id={row.title} type="file" class="hidden" accept=".jpg, .jpeg, .png, .webp, .avif" on:change={image_onSelected} bind:value={$formData[row.column]} />
    {/if}
    {#if row.view === 'password'}
        <input id={row.title} type="password" placeholder={row.placeholder} class="input {config.bordered && 'input-bordered'} {row.class} w-full" bind:value={$formData[row.column]} />
    {/if}
    {#if row.type === 'string' && !row.view}
        <input id={row.title} type="text" placeholder={row.placeholder} class="input {config.bordered && 'input-bordered'} {row.class} w-full" bind:value={$formData[row.column]} />
    {/if}
</div>
<script>
    import countries from 'i18n-iso-countries';
    import enLocale from 'i18n-iso-countries/langs/en.json';

    countries.registerLocale(enLocale);
    const countriesAll = countries.getNames("en", {select: "official"});

    export let row;
    export let config;
    export let formData;
</script>

<div>
    <label for={row.title}>{row.title}</label>
    {#if row.view === 'slug'}
        <div class="join">
            <input value="{config.domain}/" class="input input-disabled w-32" />
            <input id={row.title} type="text" placeholder={row.placeholder} class="input input-bordered w-64" bind:value={$formData[row.column]} />
        </div>
    {/if}
    {#if row.view === 'country'}
        <select id={row.title} class="select select-bordered w-full max-w-xs" value={$formData[row.column]}>
            <option value="">(not specified)</option>
            {#each Object.keys(countriesAll) as country}
                <option value={country.toLowerCase()}>{countriesAll[country]}</option>
            {/each}
        </select>
    {/if}
    {#if row.view === 'textarea'}
        <textarea id={row.title} class="textarea textarea-bordered w-full font-mono h-[25em] text-sm" placeholder={row.placeholder} bind:value={$formData[row.column]}></textarea>
    {/if}
    {#if row.view === 'string-disabled'}
        <input id={row.title} type="text" class="input input-bordered input-disabled w-96" value={row.value} />
    {/if}
    {#if row.type === 'string' && !row.view}
        <input id={row.title} type="text" placeholder={row.placeholder} class="input input-bordered w-full" bind:value={$formData[row.column]} />
    {/if}
</div>
<script>
    import { writable } from 'svelte/store';
    import FormItem from './FormItem.svelte';

    export let schema;
    export let layout;
    export let item;
    export let config;
    export let onSubmit = async (d) => { alert(JSON.stringify(d)); return false; };

    if (config.schema) {
        schema = config.schema
    }
    if (config.layout) {
        layout = config.layout
    }

    async function submitForm() {
        const resp = await onSubmit($formData)
        if (resp) {
            isChanged = false
        }
    }

    const orig = JSON.stringify(item)
    const formData = writable(item)

    function getRowConfig (item) {
        if (typeof item === 'string') {
            const opt = schema.properties[item]
            opt.column = item
            return opt
        }
        if (item.column) {
            Object.assign(item, schema.properties[item.column])
        }
        return item
    }

    let isChanged = false

    formData.subscribe((current) => {

        //console.log(current, orig)
        if (JSON.stringify(current) !== orig) {
            isChanged = true
        }
    })
    //$: isChanged = orig === JSON.parse(JSON.stringify($formData))

</script>

<form class="form" on:submit|preventDefault={submitForm}>
    {#each layout.map(getRowConfig) as row}
        {#if row.group}
            <div class="grid grid-cols-{row.group.length} gap-2">
                {#each row.group.map(getRowConfig) as subrow}
                    <FormItem row={subrow} {config} {formData} />
                {/each}
            </div>
        {:else}
            <FormItem {row} {config} {formData} />
        {/if}
    {/each}

    <div class="mt-4 border-t border-neutral/20 pt-4">
        <button type="submit" class="btn btn-primary" class:btn-disabled={!isChanged}>{config.submitButton || 'Save'}</button>
    </div>
</form>
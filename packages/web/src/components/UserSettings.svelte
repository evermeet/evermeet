<script>
    import ManagePage from './ManagePage.svelte'
    import { user } from '$lib/stores'

    export let selectedTab;

    $: u = $user
</script>

<ManagePage {selectedTab}
    baseUrl="/me/settings"
    title="Settings"
    tabs={[
        { id: null, name: 'Profile' },
        { id: 'settings', name: 'Settings' },
    ]}
    >

    {#if !selectedTab}
        
        <div>Handle: {u.handle}</div>
        <div>DID: {u.did}</div>

    {/if}

</ManagePage>
<script>
    import ManagePage from './ManagePage.svelte';

    export let selectedTab;
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
        test
    {/if}

</ManagePage>
<script>
    import ManagePage from './ManagePage.svelte';
    import AdminDashboard from './AdminDashboard.svelte';

    export let selectedTab;
</script>

<ManagePage {selectedTab}
    baseUrl="/admin"
    title="Administration"
    tabs={[
        { id: null, name: 'Dashboard' },
        { id: 'users', name: 'Users' },
        { id: 'events', name: 'Events' },
        { id: 'calendars', name: 'Calendars' },
        { id: 'settings', name: 'Settings' },
    ]}
    >

    {#if !selectedTab}
        <AdminDashboard />
    {/if}

</ManagePage>
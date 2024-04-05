<script>
    import EventManageAudit from './EventManageAudit.svelte';
    import EventManageSettings from './EventManageSettings.svelte';

    import ManagePage from './ManagePage.svelte';

    export let item;
    export let selectedTab;

</script>

<ManagePage {selectedTab}
        baseUrl="/manage/event/{item.id}"
        title={item.name}
        button={{name: 'Event page ↗', link: `/${item.slug}`}}
        breadcrumb={item.calendar ? { name: item.calendar.name, link: `/${item.calendar.slug}` } : null}
        tabs={[
            { id: null, name: 'Dashboard' },
            { id: 'guests', name: 'Guests' },
            { id: 'settings', name: 'Settings' },
            { id: 'audit', name: 'Audit' },
        ]}
    >

    {#if selectedTab === 'audit'}
        <EventManageAudit {item} />
    {/if}

    {#if selectedTab === 'insights'}
        insights
    {/if}

    {#if selectedTab === 'settings'}
        <EventManageSettings {item} />
    {/if}

    {#if !selectedTab}
        overview
    {/if}

</ManagePage>
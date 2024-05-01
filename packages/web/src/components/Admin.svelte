<script>
  import ManagePage from "./ManagePage.svelte";
  import AdminDashboard from "./AdminDashboard.svelte";
  import AdminTable from "./AdminTable.svelte";

  export let selectedTab;
</script>

<ManagePage
  {selectedTab}
  baseUrl="/admin"
  title="Administration"
  tabs={[
    { id: null, name: "Dashboard" },
    { id: "users", name: "Users" },
    { id: "events", name: "Events" },
    { id: "calendars", name: "Calendars" },
    { id: "settings", name: "Settings" },
  ]}
  width="page-extra-wide"
>
  {#if selectedTab === "users"}
    <AdminTable collection="users" />
  {/if}

  {#if selectedTab === "events"}
    <AdminTable collection="events" />
  {/if}

  {#if selectedTab === "calendars"}
    <AdminTable collection="calendars" />
  {/if}

  {#if !selectedTab}
    <AdminDashboard />
  {/if}
</ManagePage>

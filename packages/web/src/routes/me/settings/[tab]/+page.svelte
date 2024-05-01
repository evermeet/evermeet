<script>
  import UserSettings from "../../../../components/UserSettings.svelte";
  import { config } from "$lib/stores";

  export let data;
</script>

<svelte:head>
  <title>Settings | {$config.sitename || $config.domain}</title>
</svelte:head>

<UserSettings selectedTab={data.tab} />

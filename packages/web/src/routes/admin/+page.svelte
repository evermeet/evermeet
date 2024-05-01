<script>
  import Admin from "../../components/Admin.svelte";
  import { config } from "$lib/stores";
</script>

<svelte:head>
  <title>Admin | {$config.sitename || $config.domain}</title>
</svelte:head>

<Admin />

<script>
    import Admin from "../../components/Admin.svelte";
    import { config } from '$lib/config';
</script>

<svelte:head>
    <title>Admin | {config.sitename || config.domain}</title> 
</svelte:head>

<Admin />
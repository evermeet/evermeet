<script>
    import UserSettings from "../../../components/UserSettings.svelte";
    import { config } from '$lib/config';
</script>

<svelte:head>
    <title>Settings | {config.sitename || config.domain}</title> 
</svelte:head>

<UserSettings />
<script>
  import EventManage from "../../../../../components/EventManage.svelte";
  export let data;
</script>

<EventManage item={data.item} selectedTab={data.tab} />

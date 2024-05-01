<script>
  import Avatar from "svelte-boring-avatars";
  import { imgBlobUrl } from "$lib/api";

  export let user;
  export let size = 25;
  export let data = null;

  $: src =
    data ||
    (user.avatarBlob && imgBlobUrl(user.did, user.avatarBlob, size)) ||
    user.img;
</script>

{#if src}
  <img
    alt={user.handle}
    {src}
    class="rounded-full aspect-square"
    style="width: {size}px; height: {size}px;"
  />
{:else}
  <div class="rounded-full">
    <Avatar
      {size}
      name={user.did}
      variant="bauhaus"
      colors={["#ff79c6", "#ff79c6", "#bd93f9", "#ffb86c", "#414558"]}
    />
    <!--Avatar
            size={45}
            name={$user.did}
            variant="marble"
            colors={['#92A1C6', '#146A7C', '#F0AB3D', '#C271B4', '#C20D90']}
        -->
  </div>
{/if}

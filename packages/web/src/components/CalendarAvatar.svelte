<script>
  import Avatar from "svelte-boring-avatars";
  import { imgBlobUrl } from "$lib/api";

  export let calendar;
  export let key = null;
  export let size = 25;
  export let data = null;
  export let className = "";

  function imgUrl(cid) {
    let rsize = size;
    if (rsize <= 100) {
      rsize = 100;
    }
    if (rsize <= 200) {
      rsize = 200;
    }
    return imgBlobUrl(calendar.did, cid, rsize);
  }
  $: src =
    data ||
    (calendar.avatarBlob && imgUrl(calendar.avatarBlob)) ||
    calendar.img;
</script>

<div
  class="{calendar.personal
    ? 'rounded-full'
    : 'rounded'} overflow-hidden {className}"
  style="width: {size}px; height: {size}px;"
>
  {#if src}
    <img
      alt={calendar.handle}
      {src}
      class="w-full h-full aspect-square bg-base-200"
    />
  {:else}
    <Avatar
      {size}
      name={key || calendar.handle || calendar.did}
      square={true}
      variant="bauhaus"
      colors={["#ff79c6", "#ff79c6", "#bd93f9", "#ffb86c", "#414558"]}
    />
  {/if}
</div>

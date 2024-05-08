<script>
  import { calendarSubscribe, calendarUnsubscribe } from "$lib/actions";
  import { getContext } from "svelte";
  import CalendarAvatar from "./CalendarAvatar.svelte";
  import CalendarSubscribeButton from "./CalendarSubscribeButton.svelte";
  import { t } from "$lib/i18n";
  import { goto } from "$app/navigation";

  const { item, preview = null } = $props();

  const user = getContext("user");
  const c = $derived(item);

  function handleClick(url) {
    return (e) => {
      if (e.target.nodeName === "BUTTON" || e.target.nodeName === "A") {
        return false;
      }
      goto(url);
    };
  }
</script>

<div
  aria-hidden="true"
  onclick={handleClick(item.baseUrl)}
  class="cursor-pointer"
>
  <div class="itembox itembox-hover h-full group">
    <div class="flex">
      <div class="w-12 h-12 mb-2 grow">
        <CalendarAvatar calendar={item} size="48" />
      </div>
      {#if preview}
        <CalendarSubscribeButton
          {item}
          {user}
          btnClass="btn-sm"
          isCompact="true"
        />
      {/if}
    </div>
    <div class="text-lg font-semibold">
      <a href={c.baseUrl} title={c.name}>{c.name}</a>
    </div>
    {#if c._remote}
      <div class="badge badge-neutral font-mono text-xs my-2">{c._remote}</div>
    {/if}
    {#if preview}
      <div class="mt-1 text-sm text-base-content/75">{c.description}</div>
    {:else}
      <div class="text-sm mt-1 text-base-content/75">{c.subs} subscribers</div>
      <div class="mt-4 text-sm text-base-content/75">
        {#if c.personal}
          Personal
        {:else}
          {c.managers?.length || 0} admin
        {/if}
      </div>
    {/if}
  </div>
</div>

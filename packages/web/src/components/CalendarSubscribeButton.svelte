<script>
  import { page } from "$app/stores";
  import { t } from "$lib/i18n";
  import { xrpcCall } from "$lib/api";

  const { user, item, btnClass = "btn-md", isCompact = false } = $props();

  let subscribeToggleLoading = $state(false);
  async function subscribeToggle(e, remove = false) {
    e.preventDefault();
    const method = `app.evermeet.calendar.${item.$userContext?.isSubscribed ? "unsubscribe" : "subscribe"}Calendar`;
    let resp;
    subscribeToggleLoading = true;
    try {
      resp = await xrpcCall({ fetch, user }, method, null, {
        did: item.did,
      });
    } catch (e) {
      console.error(e);
      subscribeToggleLoading = false;
      return false;
    }
    item.subsCount = resp.calendar.subsCount;
    item.$userContext = resp.calendar.$userContext;
    subscribeToggleLoading = false;
  }
</script>

{#snippet button(text, badgeText, onclick, btnColor = "btn-neutral")}
  {#if typeof onclick === "string"}
    <a class="btn {btnClass} {btnColor}" href={onclick}>
      {text}
      {#if badgeText !== null}
        <span class="badge">{badgeText}</span>
      {/if}
      {#if subscribeToggleLoading}<span
          class="loading loading-infinity opacity-50"
        ></span>{/if}
    </a>
  {:else}
    <button class="btn {btnClass} {btnColor}" {onclick}>
      {text}
      {#if badgeText !== null}
        <span class="badge">{badgeText}</span>
      {/if}
      {#if subscribeToggleLoading}<span
          class="loading loading-infinity opacity-50"
        ></span>{/if}
    </button>
  {/if}
{/snippet}

{#if user}
  <div class="flex gap-1.5">
    {#if item.$userContext?.isSubscribed}
      {@render button(
        $t`Subscribed`,
        !isCompact ? item.subsCount : null,
        subscribeToggle,
        "btn-base-300",
      )}
    {:else}
      {@render button(
        $t`Subscribe`,
        !isCompact ? item.subsCount : null,
        subscribeToggle,
      )}
    {/if}
    {#if item.$userContext?.isManager && !isCompact}
      {@render button(
        $t`Manage`,
        null,
        `/manage/calendar/${item.id}`,
        "btn-accent",
      )}
    {/if}
    <!--{#if item.$userContext?.isManager}
            <a href="/manage/calendar/{item.id}" class="btn {btnClass} btn-accent">Manage</a>
        {:else if item.$userContext?.isSubscribed}
            <button class="btn {btnClass} btn-neutral" onclick={(e) => subscribeToggle(e)}
            >{$t`Subscribed`}</button
            >
        {:else}
            <button class="btn {btnClass} btn-secondary" onclick={subscribeToggle}
            >{$t`Subscribe`}</button
            >
        {/if}-->
  </div>
{:else}
  <div>
    {@render button(
      $t`Subscribe`,
      !isCompact ? item.subsCount : null,
      "/login",
    )}
  </div>
{/if}

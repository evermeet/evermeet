<script>
  import { Eye, Check } from "lucide-svelte";
  import { PencilSquare } from "svelte-heros-v2";
  import confetti from "canvas-confetti";
  import { t } from "$lib/i18n";
  import { xrpcCall } from "$lib/api";

  const { item, user, btnClass = "btn-sm", isShort = false } = $props();

  let toggleWatchLoading = $state(false);

  async function toggleWatch(e) {
    e.preventDefault();
    const method = `app.evermeet.event.${item.$userContext?.watching ? "unwatch" : "watch"}Event`;
    let resp;
    toggleWatchLoading = true;
    try {
      resp = await xrpcCall({ fetch, user }, method, null, {
        eventHandle: item.handleUrl,
      });
    } catch (e) {
      console.error(e);
      toggleWatchLoading = false;
      return false;
    }
    item.watchCount = resp.event.watchCount;
    item.$userContext = resp.event.$userContext;
    toggleWatchLoading = false;
  }

  function runConfetti(e) {
    e.preventDefault();
    confetti({
      particleCount: 200,
      spread: 140,
      decay: 0.95,
      angle: 100,
      scalar: 1.5,
      origin: {
        x: 0,
        y: 0.8,
      },
    });
    confetti({
      particleCount: 200,
      angle: 100,
      spread: 140,
      decay: 0.95,
      scalar: 1.5,
      origin: {
        x: 1,
        y: 0.8,
      },
    });
  }
</script>

<button
  class="btn {btnClass} {item.$userContext?.isAttending === true
    ? 'btn-base-300'
    : 'btn-neutral'}"
  onclick={runConfetti}
  >{#if item.$userContext?.isAttending}<Check
      size="18"
      class="text-success"
    />{:else}<PencilSquare size="18" />{/if}
  {item.$userContext?.isAttending ? $t`Attending` : $t`Attend`}</button
>
<button
  class="btn {btnClass} {item.$userContext?.watching === true
    ? 'btn-base-300'
    : 'btn-neutral'}"
  onclick={toggleWatch}
  disabled={toggleWatchLoading}
>
  <Eye size="18" class={item.$userContext?.watching ? "text-success/80" : ""} />
  {item.$userContext?.watching === true ? $t`Watching` : $t`Watch`}
  <span class="badge badge-sm">{item.watchCount || 0}</span>
  {#if toggleWatchLoading}<span class="loading loading-infinity opacity-50"
    ></span>{/if}
</button>

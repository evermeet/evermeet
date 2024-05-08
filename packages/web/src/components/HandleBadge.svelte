<script>
  import { CheckCircle, LockClosed } from "svelte-heros-v2";
  import { fade } from "svelte/transition";
  import { createTooltip, melt } from "@melt-ui/svelte";
  import { config } from "$lib/stores.js";
  import { t } from "$lib/i18n";

  export let item;
  export let size = "normal";
  export const type = null;
  export let margin = "mb-4 mt-2";

  const {
    elements: { trigger, content, arrow },
    states: { open },
  } = createTooltip({
    positioning: {
      placement: "bottom",
    },
    openDelay: 0,
    closeDelay: 0,
    closeOnPointerDown: false,
    forceVisible: true,
  });
</script>

<div class="flex gap-3 items-center {margin}">
  <a
    href="https://{item.handleUrl}"
    class="font-mono opacity-50 hover:opacity-75 {size === 'small'
      ? 'text-sm'
      : ''}">@{item.handle || item.handleUrl}</a
  >
  <div
    class="text-success opacity-50 flex gap-1 text-sm items-center hover:opacity-100 trigger cursor-help"
    use:melt={$trigger}
  >
    <LockClosed class="w-[1em]" />{$t`verified`}
  </div>
</div>

{#if $open}
  <div
    use:melt={$content}
    transition:fade={{ duration: 100 }}
    class="itembox py-3 w-[25em] shadow-xl"
  >
    <div use:melt={$arrow}></div>
    <div class="">
      <div class="mb-2 flex gap-1.5 items-center text-success opacity-75">
        {$t`This handle is verified`}
        <CheckCircle class="" />
      </div>

      <div class="mb-4 text-base-content/75 text-sm">
        {$t`You can trust that the (sub)domain belongs to the same owner as this profile.`}
      </div>
      <div class="text-sm text-base-content/75 leading-7">
        {$t`Domain`}:
        <span class="font-mono badge badge-neutral-300"
          >{item.handle || item.calendar?.handle}</span
        ><br />
        {$t`DID`}:
        <a
          href="{$config.plcServer}/{item.did ||
            'did:plc:h4c53spyxe6wab5j7jonafju'}"
          class="font-mono badge badge-neutral-300 hover:badge-accent"
          >{item.did || "did:plc:h4c53spyxe6wab5j7jonafju"}</a
        ><br />
        {$t`Verification Type`}:
        <a
          href="https://dns.google/query?name=_evermeet.{item.handle ||
            item.calendar?.handle}&rr_type=TXT&ecs="
          class="badge badge-neutral-300 hover:badge-accent"
          >{$t`DNS Record (proof)`}</a
        ><br />
      </div>
    </div>
  </div>
{/if}

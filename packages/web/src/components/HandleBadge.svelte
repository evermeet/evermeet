<script>
    import { CheckCircle, LockClosed } from 'svelte-heros-v2';
    import { fade } from 'svelte/transition';
    import { createTooltip, melt } from '@melt-ui/svelte';
    import { config } from '$lib/stores.js';

    export let item;
    export let size = 'normal';
    export let type;

    const {
        elements: { trigger, content, arrow },
        states: { open },
    } = createTooltip({
        positioning: {
        placement: 'bottom',
        },
        openDelay: 0,
        closeDelay: 0,
        closeOnPointerDown: false,
        forceVisible: true,
    });
</script>

<div class="flex gap-3 items-center mb-4 mt-2">
    <a href="https://{item.handleUrl}" class="font-mono opacity-50 hover:opacity-75 {size === 'small' ? 'text-sm' : ''}">@{item.handle || item.handleUrl}</a>
    <div class="text-success opacity-50 flex gap-1 text-sm items-center hover:opacity-100 trigger cursor-help" use:melt={$trigger}><LockClosed class="w-[1em]" />verified</div>
</div>

{#if $open}
  <div
    use:melt={$content}
    transition:fade={{ duration: 100 }}
    class="itembox py-3 w-[25em] shadow-xl"
  >
    <div use:melt={$arrow} />
    <div class="">
        <div class="mb-2 flex gap-1.5 items-center text-success opacity-75">This handle is verified <CheckCircle class="" /></div>
            
        <div class="mb-4 text-neutral-content text-sm">You can trust that the (sub)domain belongs to the same owner as this profile.</div>
        <div class="text-sm text-neutral-content leading-7">
            Domain: <span class="font-mono badge badge-neutral-300">{item.handle || item.calendar?.handle}</span><br />
            DID: <a href="{$config.plcServer}/{item.did || 'did:plc:h4c53spyxe6wab5j7jonafju'}" class="font-mono badge badge-neutral-300 hover:badge-accent">{item.did || 'did:plc:h4c53spyxe6wab5j7jonafju'}</a><br />
            Verification Type: <a href="https://dns.google/query?name=_evermeet.{item.handle || item.calendar?.handle}&rr_type=TXT&ecs=" class="badge badge-neutral-300 hover:badge-accent">DNS Record (proof)</a><br />
        </div>
    </div>
  </div>
{/if}
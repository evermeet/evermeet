<script>
    import { CheckCircle, LockClosed } from 'svelte-heros-v2';
    import { fade } from 'svelte/transition';
    import { createTooltip, melt } from '@melt-ui/svelte';

    export let item;

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
    <a href="https://{item.handleUrl}" class="font-mono opacity-50 hover:opacity-75">@{item.handle || item.handleUrl}</a>
    <div class="text-success opacity-50 flex gap-1 text-sm items-center hover:opacity-100 trigger cursor-help" use:melt={$trigger}><LockClosed class="w-[1em]" />verified</div>
</div>

{#if $open}
  <div
    use:melt={$content}
    transition:fade={{ duration: 100 }}
    class="itembox py-2"
  >
    <div use:melt={$arrow} />
    <div class="text-sm">Verification status: Verified</div>
  </div>
{/if}
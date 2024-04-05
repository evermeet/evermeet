<script>
    import { format } from 'date-fns';
    import { onMount, onDestroy } from 'svelte';
    import { getTimeZones, rawTimeZones, timeZonesNames, abbreviations } from "@vvo/tzdb";
    import { formatInTimeZone } from 'date-fns-tz';
    import { enGB } from 'date-fns/locale/en-GB'

    const currentTz = Intl.DateTimeFormat().resolvedOptions().timeZone;
    const tzData = getTimeZones().find(tz => tz.name === currentTz);
    const tzInfo = formatInTimeZone(new Date(), currentTz, 'zzzz (XXX)', { locale: enGB })

    function formatted (date = new Date) {
        return formatInTimeZone(date, currentTz, 'HH:mm zzz', { locale: enGB })
    }

    let interval;
    let time = formatted()
    onMount(() => {
        interval = setInterval(() => {
            time = formatted()
        }, 1000)
    })

    onDestroy(() => {
        clearInterval(interval)
    })
</script>

<div class="tooltip tooltip-bottom" data-tip={tzInfo}>
    <a href="https://time.is/" target="_blank" class="text-neutral-content opacity-75">{time}</a>
</div>

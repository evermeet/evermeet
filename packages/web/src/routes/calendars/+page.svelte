<script>
    import EventDetail from '../../components/EventDetail.svelte';
    import CalendarBox from '../../components/CalendarBox.svelte';
    import { config } from '$lib/config';

    const calendars = [
        {slug: "tree", name: "tree", subs: 5, img: 'https://cdn.lu.ma/cdn-cgi/image/format=auto,fit=cover,dpr=2,background=white,quality=75,width=48,height=48/avatars/39/1bf939a4-c139-4008-96d8-ad5f729d94c3', personal: true},
        {slug: "web3privacy", name: "Web3Privacy Now Events", subs: 256, img: 'https://images.lumacdn.com/cdn-cgi/image/format=auto,fit=cover,dpr=2,background=white,quality=75,width=48,height=48/calendars/97/99ef3056-7787-4e83-9247-fca4d4fa948d' },
    ]
</script>

<svelte:head>
    <title>Calendars | {config.sitename || config.domain}</title> 
</svelte:head>

<div class="page-wide">
    <h1 class="text-3xl font-semibold">Calendars</h1>

    <div class="mt-6">
        <h2 class="text-xl font-semibold">My Calendars</h2>
    </div>

    <div class="mt-6 grid grid-cols-3 gap-3">
        {#each calendars as c}
            <CalendarBox item={c} />
        {/each}
    </div>

    <div class="mt-8 w-full h-1 border border-neutral border-l-0 border-r-0 border-b-0">  
    </div>

    <div class="mt-8">
        <h2 class="text-xl font-semibold">Subscribed Calendars</h2>
    </div>

    <div class="mt-6">
        {#each calendars as c}
            <div class="mb-3 itembox flex gap-8">
                <div class="w-[12rem]">
                    <div class="w-12 h-12 mb-2">
                        <img
                            class="{c.personal ? "rounded-full" : "rounded-lg"}"
                            src={c.img} />
                    </div>
                    <div class="text-lg font-medium">{c.name}</div>
                    <button class="btn btn-sm mt-4 btn-neutral">View calendar →</button>
                </div>
                <div class="">
                    <div class="text-neutral-content text-sm">Upcoming Events</div>
                    <div class="mt-4">
                        <div class="font-medium">W3PN Meetup @ ETHDam 2024</div>
                        <div class="text-sm mt-1 text-neutral-content">Thu, Apr 11, 6:00 PM</div>
                    </div>
                    <div class="mt-4">
                        <div class="font-medium">W3PN Meetup @ ETHDam 2024</div>
                        <div class="text-sm mt-1 text-neutral-content">Thu, Apr 11, 6:00 PM</div>
                    </div>
                </div>
            </div>
        {/each}
    </div>
</div>
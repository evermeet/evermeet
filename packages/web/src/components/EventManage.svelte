<script>
    import { browser } from '$app/environment'; 
    import { parse } from 'marked';
    import { config } from '../lib/config.js';
    import { writable } from 'svelte/store';

    import countries from 'i18n-iso-countries';
    import enLocale from 'i18n-iso-countries/langs/en.json';

    countries.registerLocale(enLocale);
    const countriesAll = countries.getNames("en", {select: "official"});

    export let item;
    export let selectedTab = null;

    const tabs = [
        { id: null, name: 'Dashboard' },
        { id: 'guests', name: 'Guests' },
        { id: 'settings', name: 'Settings' },
        { id: 'audit', name: 'Audit' },
    ];

    function changeTab (tab) {
        return function() {
            selectedTab = tab;
        }
    }
    function submitForm () {
        console.log($formData)
        return false;
    }
    
    const formData = writable(item)

</script>

<div class="page-wide">
    {#if item.calendar}
        <div class="text-neutral-content text-sm"><a href="/{item.calendar.slug}">{item.calendar.name} →</a></div>
    {/if}
    <div class="flex items-center">
        <div class="text-3xl mt-1.5 font-semibold grow">{item.name}</div>
        <div><a href="/{item.slug}"><button class="btn btn-accent btn-sm">Event page ↗</button></a></div>
    </div>

    <div>
        <div class="xtabs text-left mt-6">
            {#each tabs as tab}
                <a href="/manage/event/{item.id}/{tab.id ? tab.id : ''}" class={selectedTab === tab.id ? 'active' : ''} on:click={changeTab(tab.id)}>{tab.name}</a>
            {/each}
        </div>
    </div>
</div>

<div class="h-0 border border-neutral border-t-0 border-l-0 border-r-0 mb-8">
</div>

<div class="page-wide">
    {#if selectedTab === 'insights'}
        insights
    {/if}

    {#if selectedTab === 'settings'}

        <div class="flex gap-8">
            <div>
                <ul class="menu bg-base-200 w-56 rounded-box">
                    <li>
                        <a href="#basic">Basic Info</a>
                        <ul>
                            <li><a href="#id">Event ID</a></li>
                            <li><a href="#name">Name</a></li>
                            <li><a href="#name">Public URL</a></li>
                            <li><a href="#name">Venue</a></li>
                            <li><a href="#name">Date and time</a></li>
                            <li><a href="#timezone">Timezone</a></li>
                            <li><a href="#timezone">Description</a></li>
                        </ul>
                    </li>
                    <li><a href="#appearance">Appearance</a></li>
                    <li><a href="#hosts">Hosts</a></li>
                    <li>
                        <a href="#registration">Registration</a>
                        <ul>
                            <li><a href="#capacity">Maximum capacity</a></li>
                            <li><a href="#capacity">Ticket types</a></li>
                        </ul>
                    </li>
                    <li><a href="#transfer">Calendar transfer</a></li>
                    <li><a href="#cancel">Cancel event</a></li>
                </ul>
            </div>
            <div class="settings">
                <h3 id="basic" class="text-2xl mb-8 font-semibold">Basic Info</h3>

                <form class="form" on:submit|preventDefault={submitForm}>
                    <div>
                        <label for="eid">Event ID</label>
                        <input id="eid" type="text" placeholder="Event name" class="input input-disabled input-bordered w-96" value="{item.id}@{config.domain}" />
                    </div>
                    <div>
                        <label for="event-name">Name</label>
                        <input id="event-name" type="text" placeholder="Event name" class="input input-bordered w-full" bind:value={$formData.name} />
                    </div>
                    <div>
                        <label for="slug">Public URL (slug)</label>
                        <div class="join">
                            <input value="{config.domain}/" class="input input-disabled w-32" />
                            <input id="slug" type="text" placeholder="myevent-slug" class="input input-bordered w-64" bind:value={$formData.slug} />
                        </div>
                    </div>
                    <div class="grid grid-cols-3 gap-2">
                        <div>
                            <label for="venue">Venue Name</label>
                            <input id="venue" type="text" placeholder="Venue name" class="input input-bordered w-full" bind:value={$formData.placeName} />
                        </div>
                        <div>
                            <label for="city">City</label>
                            <input id="city" type="text" placeholder="City name" class="input input-bordered w-full" bind:value={$formData.placeCity} />
                        </div>
                        <div>
                            <label for="country">Country</label>
                            <select id="country" class="select select-bordered w-full max-w-xs" value={$formData.placeCountry}>
                                <option value="">(not specified)</option>
                                {#each Object.keys(countriesAll) as country}
                                    <option value={country.toLowerCase()}>{countriesAll[country]}</option>
                                {/each}
                            </select>
                        </div>
                    </div>
                    <div class="grid grid-cols-2 gap-2">
                        <div>
                            <label for="start">Date Start</label>
                            <input id="start" type="text" placeholder="Date start" class="input input-bordered w-full" bind:value={$formData.dateStart} />                   
                        </div>
                        <div>
                            <label for="end">Date End</label>
                            <input id="end" type="text" placeholder="Date end" class="input input-bordered w-full" bind:value={$formData.dateEnd} />                   
                        </div>
                    </div>
                    <div>
                        <label for="tz">Timezone</label>
                        <input id="tz" type="text" placeholder="Timezone" class="input input-bordered w-full" bind:value={$formData.timezone} />                   
                    </div>
                    <div>
                        <label for="description">Description</label>
                        <textarea id="description" class="textarea textarea-bordered w-full font-mono h-[25em] text-sm" placeholder="Event description" bind:value={$formData.description}></textarea>
                    </div>

                    <div class="mt-4">
                        <button type="submit" class="btn btn-primary">Save</button>
                    </div>
                </form>

                <hr />

                <h3 id="hosts" class="text-2xl my-8 font-semibold">Hosts</h3>

                <hr />

                <h3 id="registration" class="text-2xl my-8 font-semibold">Registration</h3>

                <form class="form" on:submit|preventDefault={submitForm}>
                    <div>
                        <label for="capacity">Maximum capacity</label>
                        <input id="capacity" type="text" placeholder="Unlimited" class="input input-bordered w-32 mr-1.5" bind:value={$formData.maxCapacity} /> guests
                    </div>
                </form>

                <hr />

                <h3 id="cancel" class="text-2xl my-8 font-semibold">Cancel event</h3>

                <button class="btn">Delete this event</button>

            </div>
        </div>
    {/if}

    {#if !selectedTab}
        overview
    {/if}
</div>
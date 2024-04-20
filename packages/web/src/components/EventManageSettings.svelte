<script>
    import { writable } from 'svelte/store';
    import { parse } from 'marked';
    import { config } from '$lib/stores';

    //import Lexical from './Lexical.svelte';
    //import Lexical from 'svelte-lexical';
    //import {RichTextComposer} from 'svelte-lexical';
    //  import PlaygroundEditorTheme from '../themes/PlaygroundEditorTheme';

    import Form from './Form.svelte';

    export let item;

    function submitForm () {
        console.log($formData)
        return false;
    }
    
    const formData = writable(item)

    const syncMethods = [
        {
            id: "luma-r2l-v1:sync",
            name: "Lu.ma → evermeet"
        },
        {
            id: 'meetup-r2l-v1:sync',
            name: 'Meetup.com → evermeet'
        },
        {
            id: 'pretix-r2l-v1:sync',
            name: 'Pretix → evermeet'
        }
    ]

</script>

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
                    <li><a href="#description">Description</a></li>
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
            <li>
                <a href="#sync">Synchronization</a>
            </li>
            <li>
                <a href="#proposals">Proposals (CfP)</a>
            </li>
            <li><a href="#transfer">Calendar transfer</a></li>
            <li><a href="#cancel">Cancel event</a></li>
        </ul>
    </div>
    <div class="settings">
        <h3 id="basic" class="text-2xl mb-8 font-semibold">Basic Info</h3>

        <Form {item} config={$config} schema={{
            type: 'object',
            properties: {
                name: {
                    title: 'Name',
                    type: 'string',
                    placeholder: 'Your event name',
                },
                slug: {
                    title: 'Public URL (slug)',
                    type: 'string',
                    placeholder: 'myevent-slug',
                    view: 'slug',
                },
                placeName: {
                    title: 'Venue Name',
                    type: 'string',
                    placeholder: 'Venue XY'
                },
                placeCity: {
                    title: 'City',
                    type: 'string',
                    placeholder: 'My city'
                },
                placeCountry: {
                    title: 'Country',
                    type: 'string',
                    view: 'country'
                },
                dateStart: {
                    title: 'Date Start',
                    type: 'string'
                },
                dateEnd: {
                    title: 'Date End',
                    type: 'string'
                },
                timezone: {
                    title: 'Timezone',
                    type: 'string',
                    placeholder: 'Default timezone',
                },
                description: {
                    title: 'Description',
                    type: 'string',
                    placeholder: 'Event description',
                    view: 'textarea',
                },
                web: {
                    title: 'Website',
                    type: 'string',
                    placeholder: 'https://my-event.com'
                }
            }
        }} layout={[
            { title: 'Event ID', view: 'string-disabled', disabled: true, value: `${item.id}@${config.domain}` },
            'name',
            'slug',
            {
                group: [
                    'placeName',
                    'placeCity',
                    'placeCountry'
                ]
            },
            {
                group: [
                    'dateStart',
                    'dateEnd'
                ]
            },
            'timezone',
            'description',
            'web',
        ]} />

        <hr />

        <!--RichTextComposer theme={PlaygroundEditorTheme} /-->

        <h3 id="hosts" class="text-2xl my-8 font-semibold">Hosts</h3>

        <hr />

        <h3 id="registration" class="text-2xl my-8 font-semibold">Registration</h3>

        <form class="form" on:submit|preventDefault={submitForm}>
            <div>
                <label for="capacity">Maximum capacity</label>
                <input id="capacity" type="text" placeholder="Unlimited" class="input input-bordered w-32 mr-1.5" bind:value={$formData.maxCapacity} /> guests
            </div>

            <h4 id="tickets" class="text-xl my-6 font-semibold">Ticket types</h4>
        </form>

        <hr />

        <h3 id="sync" class="text-2xl my-8 font-semibold">Synchronization</h3>

        <form class="form" on:submit|preventDefault={submitForm}>
            {#if $formData.sync}
                <div class="mb-4">
                    {#each $formData.sync as si}
                        <div class="itembox">
                            <div class="font-semibold text-lg">{syncMethods.find(x => x.id === si.method).name}</div>
                            <div class="mt-4">
                                <label for="start">Lu.ma event URL</label>
                                <input id="start" type="text" placeholder="luma url" class="input input-bordered w-full" bind:value={si.url} />                   
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}

            <div class="dropdown">
                <div tabindex="0" role="button" class="btn m-1">+ Add synchronization method</div>
                <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-300 rounded-box">
                    {#each syncMethods as sm}
                        <li><a>{sm.name}</a></li>
                    {/each}
                </ul>
            </div>
        </form>

        <hr />

        <h3 id="cancel" class="text-2xl my-8 font-semibold">Cancel event</h3>

        <button class="btn">Delete this event</button>

    </div>
</div>
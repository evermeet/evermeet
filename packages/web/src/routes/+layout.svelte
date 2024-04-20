<script>
    import "../app.css";
    import "/node_modules/flag-icons/css/flag-icons.min.css";

    import { Ticket, Calendar, Sparkles, Fire, WrenchScrewdriver, Bell, MagnifyingGlass } from 'svelte-heros-v2';
    import { browser } from '$app/environment';
    import { pkg } from '../lib/config.js';
    import { page } from '$app/stores';
    import { user, config } from '$lib/stores';

    import CurrentTime from "../components/CurrentTime.svelte";

    export let data;
    user.set(data.user);
    config.set(data.config);

    const menu = [
        {
            title: 'Events',
            url: '/events',
            ico: Ticket
        },
        {
            title: 'Calendars',
            url: '/calendars',
            ico: Calendar
        },
        {
            title: 'Explore',
            url: '/explore',
            ico: Sparkles
        },
        /*{
            title: 'Admin',
            url: '/admin',
            ico: WrenchScrewdriver
        }*/
    ]
    
</script>

<svelte:head>
    <title>{$config.sitename || $config.domain}</title> 
</svelte:head>

<div class="navbar px-6">
    <div class="navbar-start">
        <div><a href="/" class="font-mono flex gap-1.5 text-sm items-center"><Fire /> {$config.sitename || $config.domain}</a></div>
    </div>
    <div class="navbar-center max-w-[80rem] w-auto">
        {#if $user}
            <ul class="menu menu-horizontal w-full gap-1">
                {#each menu as mi}
                    <li><a href={mi.url} class="{$page.url.pathname.match(new RegExp("^"+mi.url)) ? 'active' : ''}"><svelte:component this={mi.ico} /> {mi.title}</a></li>
                {/each}
            </ul>
        {/if}
    </div>
    <div class="navbar-end flex">
        <div class="text-sm">
            <ul class="menu menu-horizontal menu-sm">
                <li>
                    <CurrentTime />
                </li>
                {#if $user}
                    <li><a href="/create">Create Event</a></li>
                {:else if $page.url.pathname != '/'}
                    <li><a href="/">Explore events ↗</a></li>
                {/if}
            </ul>
        </div>
        {#if $user}
            <div class="mr-2 flex text-neutral-content gap-1">
                <div class="tooltip tooltip-bottom" data-tip="Search ⎯ ⌘K">
                    <div class="w-8 h-8 rounded-full aspect-square border border-[0.4em] border-transparent hover:border-neutral hover:bg-neutral cursor-pointer flex items-center justify-center"><MagnifyingGlass size="20" /></div>
                </div>
                <div class="w-8 h-8 rounded-full aspect-square border border-[0.4em] border-transparent hover:border-neutral hover:bg-neutral cursor-pointer flex items-center justify-center"><Bell size="20" /></div>
            </div>
            <div class="dropdown dropdown-end dropdown-hover">
                <div tabindex="0" class="">
                    <img src={$user.img} class="w-8 h-8 rounded-full aspect-square border border-[0.3em] border-transparent hover:border-neutral cursor-pointer" />
                </div>
                <ul tabindex="0" class="p-2 shadow menu dropdown-content z-[1] bg-base-300 rounded-box w-44">
                    <li><a href="/me">My profile</a></li>
                    <li><a href="/admin">Administration</a></li>
                    <li><a href="/me/settings">Settings</a></li>
                    <li><a href="/logout">Sign Out</a></li>
                </ul>
            </div>
        {:else}
            <a class="btn btn-sm btn-accent" href="/login?next={encodeURIComponent($page.url.pathname)}">Login</a>
        {/if}
    </div>
</div>

<div>
    <slot />

    <div class="page-wide">
        <footer class="footer items-center p-4 text-neutral-content border-neutral mt-16 pt-6 border border-l-0 border-r-0 border-b-0 opacity-75">
            <aside class="items-center grid-flow-col">
              <p><a href="https://github.com/evermeet/evermeet" class="hover:underline" target="_blank">{pkg.name}</a> v{pkg.version} (<a href="https://docs.evermeet.app" class="hover:underline">docs</a>)</p>
            </aside> 
            <nav class="grid-flow-col gap-4 md:place-self-center md:justify-self-end">
              <a><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" class="fill-current"><path d="M24 4.557c-.883.392-1.832.656-2.828.775 1.017-.609 1.798-1.574 2.165-2.724-.951.564-2.005.974-3.127 1.195-.897-.957-2.178-1.555-3.594-1.555-3.179 0-5.515 2.966-4.797 6.045-4.091-.205-7.719-2.165-10.148-5.144-1.29 2.213-.669 5.108 1.523 6.574-.806-.026-1.566-.247-2.229-.616-.054 2.281 1.581 4.415 3.949 4.89-.693.188-1.452.232-2.224.084.626 1.956 2.444 3.379 4.6 3.419-2.07 1.623-4.678 2.348-7.29 2.04 2.179 1.397 4.768 2.212 7.548 2.212 9.142 0 14.307-7.721 13.995-14.646.962-.695 1.797-1.562 2.457-2.549z"></path></svg>
              </a>
              <a><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" class="fill-current"><path d="M19.615 3.184c-3.604-.246-11.631-.245-15.23 0-3.897.266-4.356 2.62-4.385 8.816.029 6.185.484 8.549 4.385 8.816 3.6.245 11.626.246 15.23 0 3.897-.266 4.356-2.62 4.385-8.816-.029-6.185-.484-8.549-4.385-8.816zm-10.615 12.816v-8l8 3.993-8 4.007z"></path></svg></a>
              <a><svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" class="fill-current"><path d="M9 8h-3v4h3v12h5v-12h3.642l.358-4h-4v-1.667c0-.955.192-1.333 1.115-1.333h2.885v-5h-3.808c-3.596 0-5.192 1.583-5.192 4.615v3.385z"></path></svg></a>
            </nav>
        </footer>
    </div>
</div>

<script>
    import { user, session } from '$lib/stores'
    import Cookies from 'js-cookie'
    import { goto } from '$app/navigation';
    import { createDropdownMenu, melt } from '@melt-ui/svelte';
    import { writable } from 'svelte/store';
    import { fly } from 'svelte/transition';
    
    import UserAvatar from './UserAvatar.svelte';

    const {
        elements: { trigger, menu, item, separator, arrow },
        builders: { createSubmenu },
        states: { open },
    } = createDropdownMenu({
        forceVisible: true,
        loop: true,
    });

    const {
        elements: { subMenu, subTrigger },
        states: { subOpen },
    } = createSubmenu();
    
    function doLogout () {
        session.set(null)
        Cookies.remove('evermeet-session-id')
        goto('/')
    }
</script>


<button class="block trigger rounded-full border-[0.3em] border-transparent hover:border-neutral" use:melt={$trigger} aria-label="User menu">
    <UserAvatar user={$user} />
</button>

{#if $open}
    <ul class="popup-menu text-sm max-w-96" use:melt={$menu} transition:fly={{ duration: 150, y: -10 }}>
        <li>
            <div class="flex gap-4 items-center bg-base-200 py-2 px-4 rounded mb-2">
                <UserAvatar user={$user} size="45" />
                <div class="">
                    <div>{$user.name}</div>
                    <div class="text-base-content/75 break-all">@{$user.handle}</div>
                </div>
            </div>

        </li>
        <li use:melt={$item}><a href="/me">My profile</a></li>
        <li use:melt={$item}><a href="/admin">Administration</a></li>
        <li use:melt={$item}><a href="/me/settings">Settings</a></li>
        <li use:melt={$item}><a alt="Sign Out" on:click|preventDefault={doLogout}>Sign Out</a></li>
    </ul>
{/if}


    <!--ul tabindex="0" class="p-2 shadow menu dropdown-content z-[1] bg-base-300 rounded-box w-44">

    </ul-->

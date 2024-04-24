
import { Ticket, Calendar, Sparkles, PlusCircle, Document, 
    ArrowLeftStartOnRectangle, ArrowRightEndOnRectangle } from 'svelte-heros-v2';

export function searchItemsBase () {
    return [
        {
            id: 'events',
            type: 'general',
            name: 'Events',
            handle: 'events',
            baseUrl: '/events',
            icon: Ticket,
        },
        {
            id: 'calendars',
            type: 'general',
            name: 'Calendars',
            baseUrl: '/calendars',
            icon: Calendar,
        },
        {
            id: 'explore',
            type: 'general',
            name: 'Explore',
            baseUrl: '/',
            icon: Sparkles,
        },
        {
            id: 'create-calendar',
            type: 'general',
            name: 'Create Calendar',
            baseUrl: '/create-calendar',
            icon: PlusCircle,
            description: 'Create a new Calendar',
            keywords: 'cc c c',
        },
        {
            id: 'create-event',
            type: 'general',
            name: 'Create Event',
            baseUrl: '/create',
            icon: PlusCircle,
            description: 'Create a new Event',
            keywords: 'ce c e create create',
        },
        {
            id: 'documentation',
            type: 'general',
            name: 'Documentation',
            baseUrl: 'https://docs.evermeet.app',
            icon: Document,
            description: 'Read documentation of Evermeet',
            keywords: 'docs',
        },
        {
            id: 'logout',
            type: 'general',
            name: 'Sign out',
            baseUrl: '/logout',
            icon: ArrowLeftStartOnRectangle,
            description: 'Logout from current session',
            keywords: 'lo so logout signout exit',
        },
        {
            id: 'login',
            type: 'general',
            name: 'Sign in',
            baseUrl: '/login',
            icon: ArrowRightEndOnRectangle,
            description: 'Login to current instance',
            keywords: 'li si sign in log in',
        },
    ];
}
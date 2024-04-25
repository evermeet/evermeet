<script>
    import { writable } from "svelte/store";
    import TextareaAdvanced from "../../components/TextareaAdvanced.svelte";
    import CalendarSelector from "../../components/CalendarSelector.svelte";
    import VisibilitySelector from "../../components/VisibilitySelector.svelte";
    import DateSelector from "../../components/DateSelector.svelte";
    import { stringify } from "yaml";
    import { PlayCircle, StopCircle, Clock } from 'svelte-heros-v2';


    export let data;

    const name = writable('');
    const calendar = writable({ value: data.calendars[0].id, label: data.calendars[0].id });
    const visibility = writable('public');
    const dateStart = writable(new Date());
    const dateEnd = writable(new Date());

    $: output = {
        name: $name,
        calendarId: $calendar.value,
        visibility: $visibility,
        dateStart: $dateStart,
        dateEnd: $dateEnd,
    }

</script>

<div class="page-wide">
    <h1 class="heading1">Create Event</h1>

    <div class="flex gap-8">
        <div class="w-[330px]">
            <div class="w-[330px]">
                <img
                    class="w-[330px] h-[330px] rounded-xl aspect-square bg-base-300" 
                    src="https://cdn.lu.ma/cdn-cgi/image/format=auto,fit=cover,dpr=2,quality=75,width=400,height=400/event-defaults/1-1/standard3.png" />
            </div>
        </div>
        <div class="w-full">
            <div class="flex gap-2 justify-end items-end">
                <CalendarSelector items={data.calendars} bind={calendar} />  
                <VisibilitySelector bind={visibility} />
            </div>

            <div class="mt-8">
                <TextareaAdvanced placeholder="Event Name" str={name} />
            </div>

            <div class="mt-8 bg-base-200 rounded-lg">
                <div class="px-6 py-1 flex gap-4">
                    <ul class="timeline timeline-vertical timeline-snap-icon timeline-compact opacity-75">
                        <li class="">
                            <div class="timeline-middle pt-1 text-base-content/75">
                                <div class="w-2.5 h-2.5">
                                    <svg height="100%" width="100%" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg"><circle r="45" cx="50" cy="50" fill="currentColor" /></svg>
                                </div>
                            </div>
                            <div class="timeline-end ml-3 mb-2">
                                <div class="flex">
                                    <div class="w-16 text-base-content/75">Start</div>
                                    <DateSelector bind={dateStart} />
                                </div>
                            </div>
                            <hr class="bg-neutral-content/25 mt-1"/>
                        </li>
                        <li>
                            <hr class="bg-neutral-content/25 mt-1"/>
                            <div class="timeline-middle pt-1 text-base-content/75">
                                <div class="w-2.5 h-2.5 ">
                                    <svg height="100%" width="100%" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg"><circle r="45" cx="50" cy="50" stroke="currentColor" fill="transparent" stroke-width="10" /></svg>
                                </div>
                            </div>
                            <div class="timeline-end ml-3 flex">
                                <div class="w-16 text-base-content/75">End</div>
                                <DateSelector bind={dateEnd} />
                            </div>
                        </li>
                
                    </ul>
                </div>
                <div class="py-1.5 px-5 w-full">
                    <div class="text-sm text-base-content/25 w-full flex items-center gap-2">
                        <Clock size="16" /> Duration: 1 hour 15 minutes
                    </div>
                </div>
            </div>


            <div class="mt-10">
                <button class="btn btn-primary w-full">Create Event</button>
            </div>
        </div>
    </div>

    <h2 class="heading2 mt-10">YAML Preview</h2>
    <pre class="pre font-mono p-4 bg-base-200 text-sm w-full text-base-content">{stringify(output)}</pre>
</div>
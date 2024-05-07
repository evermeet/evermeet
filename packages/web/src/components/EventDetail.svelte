<script>
  import { MapPin, CheckCircle, VideoCamera } from "svelte-heros-v2";
  import {
    Clock,
    KeyRound,
    UserCheck,
    UserX,
    EyeOff,
    Webcam,
    Eye,
    Check,
  } from "lucide-svelte";
  import {
    format,
    interval,
    formatDateInterval,
    formatTimeInterval,
    formatDurationInterval,
    timezonesOffset,
  } from "$lib/date";

  import { parse, parseInline } from "marked";
  import { config } from "$lib/stores";
  import { register, unregister } from "$lib/actions";
  import FlagIcon from "./FlagIcon.svelte";
  import HandleBadge from "./HandleBadge.svelte";
  import CalendarAvatar from "./CalendarAvatar.svelte";
  import { t, T, getCountryName } from "$lib/i18n";
  import { getContext } from "svelte";
  import confetti from "canvas-confetti";

  const { item } = $props();

  const { dateLocale: locale, timezone, lang } = getContext("locale");
  const user = getContext("user");

  const itemInterval = $derived(interval(item.dateStart, item.dateEnd));
  const itemTimezonesOffset = $derived(
    timezonesOffset(itemInterval.start, timezone, item.timezone),
  );

  let countryName = $derived(
    item.placeCountry ? getCountryName(item.placeCountry, lang) : "",
  );

  let userRegistered = $derived(
    user && user.events?.find((e) => e.ref === item.id) ? true : false,
  );

  function runConfetti() {
    confetti({
      particleCount: 200,
      spread: 140,
      decay: 0.95,
      angle: 100,
      scalar: 1.5,
      origin: {
        x: 0,
        y: 0.8,
      },
    });
    confetti({
      particleCount: 200,
      angle: 100,
      spread: 140,
      decay: 0.95,
      scalar: 1.5,
      origin: {
        x: 1,
        y: 0.8,
      },
    });
  }
</script>

<svelte:head>
  <title>{item.name} | {$config.sitename || $config.domain}</title>
</svelte:head>

<div data-id="#{item.handle}" class="flex gap-8">
  <div class="w-[330px]">
    <div class="w-[330px]">
      <img
        class="w-full aspect-square rounded-xl"
        src={item.img}
        alt={item.name}
      />
    </div>

    {#if user}
      <div class="mt-6">
        <div class="itembox text-base-content/75 flex gap-3">
          <div>{$t`You have manage access for this event.`}</div>
          <a class="btn btn-secondary" href="/manage/event/{item.id}"
            >{$t`Manage`}&nbsp;↗</a
          >
        </div>
      </div>
    {/if}

    {#snippet calendarMiniBox(c, series = false)}
      <div class="mt-6">
        <div class="flex gap-4 items-center">
          <div class="w-10 h-10 aspect-square">
            <CalendarAvatar calendar={c} size="40" />
          </div>
          <div>
            <div class="text-sm">
              {#if series}{$t`Series`}{:else}{$t`Presented by`}{/if}
            </div>
            <div class="font-medium">
              <a href={c.baseUrl}>{c.name}</a>
            </div>
          </div>
        </div>
        {#if c.description}
          <div class="mt-4 text-sm text-base-content/75">
            {c.description}
          </div>
        {/if}
      </div>
    {/snippet}

    {#if item.calendar}
      {@render calendarMiniBox(item.calendar, !!item.calendar.parent)}
      {#if item.calendar.parent}
        {@render calendarMiniBox(item.calendar.parent)}
      {/if}
    {/if}

    {#if item.hosts}
      <div
        class="mt-6 border-t-0 border-l-0 border-r-0 border border-neutral pb-2"
      >
        <div class="font-mono text-sm">Hosted by</div>
      </div>
      <div class="mt-4 mb-8">
        {#each item.hosts as host}
          <div class="flex gap-2 mb-2">
            <div class="w-6 h-6">
              <img
                class="rounded-full w-6 h-6 aspect-square object-cover"
                src={host.img}
                alt={host.name}
              />
            </div>
            <div>{host.name}</div>
          </div>
        {/each}
      </div>
    {/if}

    {#if item.guestCountTotal}
      <div
        class="mt-6 border-t-0 border-l-0 border-r-0 border border-neutral pb-2"
      >
        <div class="font-mono text-sm">{item.guestCountTotal} going</div>
      </div>
      <div class="mt-4 mb-8 text-sm text-base-content/75">
        {#if item.guests}
          {item.guests
            .map((g) => g.name)
            .slice(0, 2)
            .join(", ")} and {item.guestCountTotal} others
        {:else}{/if}
      </div>
    {/if}

    <div class="mt-8 text-sm text-accent">
      <a href="#contact">{$t`Contact the host`}</a>
    </div>
  </div>
  <div class="w-full">
    <div class="">
      <h1 class="text-5xl font-semibold font-mono">{item.name}</h1>
      <HandleBadge {item} size="small" type="event" margin="mt-2 mb-2" />
    </div>
    <div class="mb-6 flex gap-1.5">
      <button class="btn btn-sm btn-base-300" onclick={runConfetti}
        ><Check size="18" class="text-success" />
        {$t`Attending`} <span class="badge badge-sm">1</span></button
      >
      <button class="btn btn-sm btn-neutral" onclick={runConfetti}
        ><Eye size="18" />
        {$t`Watch`} <span class="badge badge-sm">10</span></button
      >
    </div>
    <div class="flex gap-4 items-center">
      <div class="w-10 h-10 border rounded-lg border-neutral">
        <div class="text-center">
          <div class="text-xs uppercase bg-neutral rounded-t-lg">
            {format(item.dateStart, "MMM", { locale })}
          </div>
          <div class="">{format(item.dateStart, "d", { locale })}</div>
        </div>
      </div>
      <div>
        <div class="text-lg font-mono">
          {formatDateInterval(itemInterval, { locale })}
        </div>
        <div class="text-sm text-base-content/75">
          {formatTimeInterval(itemInterval, { locale, user, tz: timezone })}
          <span class="text-base-content/40"
            >• {formatDurationInterval(itemInterval, { locale })}</span
          >
        </div>
        {#if itemTimezonesOffset !== 0}
          <div class="text-sm text-accent/75">
            <a
              href="https://time.is/compare/{format(
                item.dateStart,
                'HHmm_d_MMM_y',
                { timezone: item.timezone },
              )}_in_{item.placeCity.replace(/ /, '_')}"
            >
              {formatTimeInterval(itemInterval, {
                locale,
                user,
                tz: item.timezone,
              })}
            </a>
            ({item.timezone})
          </div>
        {/if}
      </div>
    </div>
    <div class="flex gap-4 items-center mt-4">
      <div
        class="w-10 h-10 border rounded-lg border-neutral flex justify-center items-center"
      >
        {#if item.mode === "offline"}
          <MapPin />
        {:else if item.mode === "online"}
          <VideoCamera />
        {:else}{/if}
      </div>
      <div>
        <div class="font-mono text-lg">
          {#if item.mode === "offline"}
            <!-- OFFLINE EVENT -->
            {#if item.placeRestrictedToGuests}
              <EyeOff size={20} strokeWidth={1.25} class="inline-block" />
              {$t`Register To See Address`}
            {:else if item.placeName}{item.placeName}{:else}<span
                class="opacity-75">{$t`To Be Determined (TBD)`}</span
              >{/if}
          {:else if item.mode === "online"}
            <!-- ONLINE EVENT -->
            {#if true}
              {$t`Online Event`}
            {:else if item.joinUrl}
              <a href={item.joinUrl} class="text-sm hover:underline"
                >{item.joinUrl.replace(/^https:\/\//, "")}</a
              >
            {:else}
              <span class="opacity-75">{$t`Link not yet published`}</span>
            {/if}
          {:else}
            {item.place}
          {/if}
        </div>
        {#if item.placeCity && countryName && item.mode !== "online"}
          <div class="text-sm text-base-content/75 flex gap-2 items-center">
            <div>{item.placeCity}, {countryName}</div>
            <FlagIcon country={item.placeCountry} {countryName} />
          </div>
        {/if}
      </div>
    </div>

    {#if item.registration?.enabled}
      <div class="mt-6 itembox no-padding w-full">
        {#if !userRegistered}
          <div class="bg-neutral rounded-t-lg py-2 px-4 text-sm">
            {$t`Registration`}
          </div>
        {/if}
        <div class="py-4 px-4">
          {#if userRegistered}
            <div>
              <img
                src={user.img}
                alt={user.name}
                class="w-10 rounded-full mb-2"
              />
            </div>
            <div class="text-xl font-semibold font-mono">{$t`You’re In`}</div>
            <div class="mt-1.5 text-base-content/75">
              {$t`A confirmation email has been sent to your email.`}
            </div>
            <div class="mt-4 text-sm text-base-content/75">
              <T msg="No longer able to attend? Notify the host by #.">
                <button
                  class="underline text-accent opacity-75 hover:opacity-100"
                  onclick={unregister(item.id)}
                  >{$t`canceling your registration`}</button
                >
              </T>
            </div>
          {:else}
            <div class="mb-4">
              {$t`Welcome! To join the event, please register below.`}
            </div>
            {#if user}
              <div class="flex gap-2 mt-4 items-center">
                <div>
                  <img
                    src={user.img}
                    alt={user.name}
                    class="rounded-full w-5"
                  />
                </div>
                <div class="font-semibold">
                  {user.name}
                  <!--span class="font-mono text-xs badge badge-neutral inline-block">{$user.did}</span-->
                </div>
              </div>
            {/if}
            <div class="mt-3">
              <button
                class="btn w-full text-lg btn-neutral"
                onclick={register(item.id)}>{$t`Register`}</button
              >
            </div>
          {/if}
        </div>
      </div>
    {/if}

    {#if item.description}
      <div
        class="mt-6 border-t-0 border-l-0 border-r-0 border border-neutral pb-2"
      >
        <h2 class="font-mono text-sm">{$t`About Event`}</h2>
      </div>

      <div class="mt-4 prose">
        {@html parse(item.description)}
      </div>
    {/if}
    {#if item.peopleLists && item.peopleLists.length > 0}
      <div
        class="mt-6 border-t-0 border-l-0 border-r-0 border border-neutral pb-2"
      >
        <h2 class="font-mono text-sm">{$t`People`}</h2>
      </div>
      <div class="mt-4">
        {#each item.peopleLists as pl}
          <div class="mb-6 relative">
            <h3 class="text-lg">{pl.name} ({pl.people.length})</h3>
            <div class="mt-4 relative grid grid-cols-5 gap-1.5 overflow-x-auto">
              {#each pl.people.map( (p) => item.people.find((i) => i.id === p), ) as p}
                <div class="shrink-0 group hover:bg-base-300 rounded-xl p-1.5">
                  <div>
                    <img
                      src={p.img}
                      alt={p.name}
                      class="rounded-xl aspect-square object-cover w-full bg-base-300 opacity-85 group-hover:grayscale-0 group-hover:opacity-100"
                    />
                  </div>
                  <div class="text-center mt-1.5 text-sm text-base-content/75">
                    {p.name}
                  </div>
                  <div class="text-xs text-base-content/50 text-center mt-1.5">
                    {p.caption}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

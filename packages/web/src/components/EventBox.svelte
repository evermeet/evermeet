<script>
  import { getContext } from "svelte";
  import { MapPin, Users, VideoCamera } from "svelte-heros-v2";
  import EventPrimaryButtons from "./EventPrimaryButtons.svelte";
  import FlagIcon from "./FlagIcon.svelte";
  import { config } from "$lib/stores";
  import { t } from "$lib/i18n";
  import { goto } from "$app/navigation";
  import {
    interval,
    formatTimeInterval,
    formatDurationInterval,
    timezonesOffset,
  } from "$lib/date";

  let titleElement = $state();

  const { item, virtual = false } = $props();
  const { dateLocale: locale, timezone, lang } = getContext("locale");
  const user = getContext("user");
  const itemInterval = $derived(interval(item.dateStart, item.dateEnd));
  const itemTimezonesOffset = $derived(
    timezonesOffset(itemInterval.start, timezone, item.timezone),
  );
  //const duration = $derived(intervalToDuration(itemInterval))

  function handleClick(url) {
    return (e) => {
      if (
        e.target.nodeName === "BUTTON" ||
        e.target.nodeName === "A" ||
        e.target.parentElement.nodeName === "BUTTON"
      ) {
        return false;
      }
      goto(url);
    };
  }
</script>

<div
  aria-hidden="true"
  onclick={!virtual ? handleClick(item.baseUrl) : () => {}}
  class={!virtual ? "cursor-pointer" : ""}
>
  <div class="mb-3 itembox {!virtual && 'itembox-hover'} flex gap-8">
    <div class="grow">
      <div class="text-base-content/75">
        {formatTimeInterval(itemInterval, { locale, user, tz: timezone })}
        {#if itemTimezonesOffset !== 0}
          •
          <span class="text-accent"
            >{formatTimeInterval(itemInterval, {
              locale,
              user,
              tz: item.timezone,
            })}</span
          >
        {/if}
        <span class="text-base-content/40">
          •
          {formatDurationInterval(itemInterval, { locale })}
        </span>
      </div>
      <div class="mt-1.5">
        <a
          href={item.baseUrl}
          class="text-xl font-medium"
          bind:this={titleElement}>{item.name}</a
        >
      </div>
      <!--div class="text-base-content/75 mt-1.5">by CryptoCanal</div-->
      <div class="flex gap-1.5 mt-1.5 text-base-content/75 items-center">
        {#if item.mode === "offline"}
          <MapPin tabindex="-1" />
          {#if item.placeRestrictedToGuests}
            <div>{$t`Only for Registered`}</div>
          {:else}
            <div>{item.placeName || "TBD"}</div>
          {/if}
        {:else if item.mode === "online"}
          <VideoCamera />
          <div>{$t`Online`}</div>
        {:else}
          <div>{item.placeName}</div>
        {/if}

        <div class="ml-2 opacity-75">
          <FlagIcon country={item.placeCountry} />
        </div>
        <div>{item.placeCity}</div>
      </div>
      {#if item.guestCount}
        <div class="flex gap-1.5 mt-1.5 text-base-content/75">
          <Users />
          <div>{item.guestCount} guests</div>
        </div>
      {/if}
      {#if item._remote}
        <div class="badge badge-neutral font-mono text-xs mt-2.5 opacity-50">
          {item._remote}
        </div>
      {/if}
      {#if !virtual}
        <div class="flex gap-1.5 mt-6">
          <EventPrimaryButtons {item} {user} btnClass="btn-sm" isShort="true" />
        </div>
      {/if}
    </div>
    <div class="w-[120px] h-[120px]">
      <img
        class="w-[120px] h-[120px] aspect-square rounded-lg"
        src={item.img}
        alt={item.name}
      />
    </div>
  </div>
</div>

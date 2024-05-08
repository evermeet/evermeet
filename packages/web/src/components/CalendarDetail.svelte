<script>
  import EventList from "./EventList.svelte";
  import { parse } from "marked";
  import { config } from "$lib/stores";
  import {
    CheckCircle,
    LockClosed,
    CalendarDays,
    QueueList,
    ChatBubbleLeft,
    Clipboard,
    Inbox,
    Users,
    PresentationChartBar,
    ArrowsPointingOut,
    ArrowsPointingIn,
  } from "svelte-heros-v2";
  import HandleBadge from "./HandleBadge.svelte";
  import CalendarAvatar from "./CalendarAvatar.svelte";
  import EventBox from "./EventBox.svelte";
  import Refs from "./Refs.svelte";
  import Chat from "./Chat.svelte";
  import ChatRoomSelect from "./ChatRoomSelect.svelte";
  import CalendarSubscribeButton from "./CalendarSubscribeButton.svelte";
  import { imgBlobUrl } from "$lib/api";
  import { t, plural } from "$lib/i18n";
  import { getContext } from "svelte";

  const { item, selectedTab, params } = $props();
  const user = getContext("user");

  let tabs = [
    {
      id: null,
      name: $t`Events`,
      ico: CalendarDays,
      bubble: item.events.length,
    },
    {
      id: "concepts",
      name: $t`Drafts`,
      ico: Inbox,
      bubble: item.concepts.length,
      bubbleAccent: true,
    },
    { id: "talks", name: $t`Talks`, ico: PresentationChartBar },
    { id: "contributors", name: $t`People`, ico: Users },
    //{ id: 'about', name: 'About' },
    { id: "feed", name: $t`Feed`, ico: QueueList },
  ];

  if (item.rooms && item.rooms.length > 0) {
    tabs.push({ id: "chat", name: $t`Chat`, ico: ChatBubbleLeft });
  }

  const isFullPage = $derived(selectedTab === null && !params.expand);
  const backdropImg = $derived(
    isFullPage &&
      ((item.headerBlob && imgBlobUrl(item.did, item.headerBlob, 1000)) ||
        item.backdropImg),
  );

  const currentRoom = $derived(
    typeof params.room === "string" ? params.room : "general",
  );
  const currentRoomObj = $derived(
    item.rooms.find((r) => r.slug === currentRoom),
  );

  function changeTab() {}
</script>

<svelte:head>
  <title>{item.name}</title>
</svelte:head>

{#if backdropImg}
  <div class="page-extra-wide relative">
    <div class="">
      <img
        alt={item.name}
        class="lg:rounded-2xl w-full object-cover max-h-[308px] aspect-[3.5/1] bg-base-300"
        src={backdropImg}
      />
    </div>
  </div>
{/if}

<div class="page-wide {backdropImg ? '-mt-12' : ''} z-0">
  <div class="flex items-center mb-2 gap-4">
    <div
      class="{backdropImg ? 'mb-6' : !isFullPage ? '' : 'mt-10'} {isFullPage
        ? 'grow'
        : ''} z-10"
    >
      <CalendarAvatar
        calendar={item}
        size={isFullPage ? 88 : 60}
        className={backdropImg ? "border-neutral/50 border-4" : ""}
      />
    </div>
    {#if !isFullPage}
      <div class="grow">
        <h1 class="text-3xl font-medium">{item.name}</h1>
        <HandleBadge {item} margin="mt-1.5" size="small" />
      </div>
      <div class="">
        <CalendarSubscribeButton {item} {user} btnClass="btn-sm" />
      </div>
    {/if}
  </div>
  {#if isFullPage}
    <div class="flex {item.backdropImg ? '' : 'mt-6'}">
      <h1 class="grow text-4xl font-medium">
        {item.name}
      </h1>
      <div>
        <CalendarSubscribeButton {item} {user} />
      </div>
    </div>
    <HandleBadge {item} />

    {#if item.description}
      <div class="mt-3 text-base-content/75">
        {@html parse(item.description)}
      </div>
    {/if}
    {#if item.refs}
      <div class="mt-4">
        <Refs {item} />
      </div>
    {/if}
  {/if}
</div>

<div class="page-wide">
  <div>
    <div class="xtabs text-left mt-6 text-sm">
      {#each tabs as tab}
        <a
          href="{item.baseUrl}/{tab.id ? tab.id : ''}"
          class="{selectedTab === tab.id
            ? 'active'
            : ''} flex items-center gap-1.5"
          onclick={changeTab(tab.id)}
        >
          {#if tab.ico}
            <svelte:component
              this={tab.ico}
              size="18"
              class="outline-none"
              tabindex="-1"
            />
          {/if}
          {tab.name}
          {#if tab.bubble}
            <div
              class="badge badge-sm {tab.bubbleAccent
                ? 'badge-accent'
                : 'badge-neutral'}"
            >
              {tab.bubble}
            </div>
          {/if}
        </a>
      {/each}
      <div class="grow"></div>
      {#if selectedTab === null}
        <div class="">
          {#if params.expand}
            <a
              href="{item.baseUrl}/{selectedTab || ''}"
              class="flex items-center gap-1.5 opacity-50 text-xs p-2 hover:opacity-100"
              ><ArrowsPointingIn size="16" class="" /> Show info</a
            >
          {:else}
            <a
              href="{item.baseUrl}/{selectedTab || ''}?expand"
              class="flex items-center gap-1.5 opacity-50 text-xs p-2 hover:opacity-100"
              ><ArrowsPointingOut size="16" class="" /> Expand</a
            >
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>

<div class="h-1 border border-neutral border-b-0 border-l-0 border-r-0"></div>

{#if selectedTab === "feed"}
  <div class="page-wide">
    <h2 class="text-2xl font-medium mt-6">{$t`Feed`}</h2>
  </div>
{:else if selectedTab === "concepts"}
  <div class="page-wide">
    <h2 class="text-2xl font-medium mt-6">{$t`Event Concepts`}</h2>

    <div class="mt-6">
      {#each item.concepts as event}
        <EventBox item={event} />
      {/each}
    </div>
  </div>
{:else if selectedTab === "contributors"}
  <div class="page-wide">
    <h2 class="text-2xl font-medium mt-6">{$t`Speakers`}</h2>

    <div class="mt-6">Contributors</div>
  </div>
{:else if selectedTab === "chat"}
  <div class="page-wide"></div>
  <div class="page-wide">
    <ChatRoomSelect rooms={item.rooms} baseUrl={item.baseUrl} {currentRoom} />
    <Chat {item} room={currentRoomObj} chatData={item._chat} />
  </div>
{:else if selectedTab === null}
  <div class="page-wide">
    {#if item.childrens?.length > 0}
      <h2 class="text-2xl font-medium mt-6">{$t`Series`}</h2>
      <div class="flex gap-3 mt-3">
        {#each item.childrens as c}
          <a
            href={c.baseUrl}
            class="py-3 px-4 bg-base-300/50 hover:bg-base-300 rounded-xl flex gap-4 items-center"
          >
            <div>
              <img
                src={c.img}
                alt={c.name}
                class="rounded-lg aspect-square object-cover w-16"
              />
            </div>
            <div class="text-left">
              <div class="text-lg">{c.name}</div>
              <div class="text-base-content/75 text-sm">
                {$plural(c.subsCount, {
                  one: "# subscriber",
                  other: "# subscribers",
                })}
              </div>
            </div>
          </a>
        {/each}
      </div>
    {/if}
    <h2 class="text-2xl font-medium mt-6">{$t`Planned Events`}</h2>
    <div class="mt-6">
      <EventList events={item.events} />
    </div>
    {#if item.pastEvents && item.pastEvents.length > 0}
      <h2 class="text-2xl font-medium mt-6">{$t`Past Events`}</h2>
      <div class="mt-6 transition-opacity">
        <EventList events={item.pastEvents} rowClass="" />
      </div>
    {/if}
  </div>
{/if}

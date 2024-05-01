<script>
  export let selectedTab = null;
  export let baseUrl;
  export let title;
  export let button;
  export let breadcrumb;
  export let tabs;
  export let width = "page-wide";

  function changeTab(tab) {
    return function () {
      selectedTab = tab;
    };
  }
</script>

<div class={width}>
  {#if breadcrumb}
    <div class="text-base-content/75 text-sm mb-1.5">
      <a href={breadcrumb.link}>{breadcrumb.name} →</a>
    </div>
  {/if}
  <div class="flex items-center">
    <div class="text-3xl font-semibold grow">{title}</div>
    {#if button}
      <div>
        <a class="btn btn-accent btn-sm" href={button.link}>{button.name}</a>
      </div>
    {/if}
  </div>

  <div>
    <div class="xtabs text-left mt-6">
      {#each tabs as tab}
        <a
          href="{baseUrl}/{tab.id ? tab.id : ''}"
          class={selectedTab === tab.id ? "active" : ""}
          on:click={changeTab(tab.id)}>{tab.name}</a
        >
      {/each}
    </div>
  </div>
</div>

<div
  class="h-0 border border-neutral border-t-0 border-l-0 border-r-0 mb-8"
></div>

<div class={width}>
  <slot />
</div>

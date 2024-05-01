<script>
  import {
    UserPlus,
    Check,
    ExclamationCircle,
    ArrowPath,
    Bolt,
    Key,
  } from "svelte-heros-v2";
  import { pkg } from "../../lib/config.js";
  import { xrpcCall } from "../../lib/api.js";
  import { writable } from "svelte/store";
  import { goto } from "$app/navigation";
  import { user, config } from "$lib/stores";
  import { generate } from "random-words";

  const registerData = writable({});
  const did = writable(null);
  let isProcessing = false;
  let validData = false;
  let validHandle = false;
  let customDomainMode = false;

  registerData.subscribe((d, p) => {
    validHandle =
      d.handle && d.handle.match(/^[0-9a-zA-Z-]+$/) && d.handle.length >= 3;
    validData =
      validHandle &&
      d.password &&
      d.passwordRepeat &&
      d.password === d.passwordRepeat;

    console.log({ d, p });
    if (d?.handle?.did !== $did?.handle) {
      did.set(null);
    }
  });
  async function submitRegistration() {
    const d = $registerData;
    isProcessing = true;
    let resp;
    try {
      resp = await xrpcCall(
        { fetch },
        "app.evermeet.auth.createAccount",
        null,
        {
          handle: d.handle + "." + $config.domain,
          password: d.password,
        },
      );
    } catch (e) {}
    isProcessing = false;
    console.log(resp);
    return null;
  }

  function generateHandle() {
    registerData.update((x) => {
      x.handle = generate({ exactly: 2 }).join("-");
      return x;
    });
    return true;
  }

  function toggleCustomDomainMode() {
    customDomainMode = !customDomainMode;
  }

  async function generateDid() {
    let resp;
    try {
      resp = await xrpcCall({ fetch }, "app.evermeet.auth.generateDid", null, {
        handle: $registerData.handle,
      });
    } catch (e) {}
    did.set({ did: resp.did, handle: $registerData.handle });
    return null;
  }
</script>

<div class="w-[28rem] m-auto my-24 itembox p-6 shadow-xl">
  <UserPlus size="50" />
  <div class="text-2xl mt-4">Create Personal Identity</div>
  <div class="mt-2 text-base-content/75">
    Here you can register new user account on <span class="font-mono"
      >{$config.domain}</span
    >.
  </div>
  <div class="mt-2 text-base-content/75 opacity-50">
    If you already have account, you can <a
      href="/login"
      class="underline hover:no-underline">sign in</a
    >.
  </div>

  <form class="mt-6" on:submit|preventDefault={submitRegistration}>
    <div class="flex items-center">
      <label for="username" class="grow">Handle</label>
      {#if !customDomainMode}
        <div class="align-right text-xs opacity-50 hover:opacity-100">
          <button
            class="hover:underline flex gap-1.5 items-center active:opacity-75"
            on:click|preventDefault={generateHandle}
            ><ArrowPath tabindex="-1" size="16" /> Random handle</button
          >
        </div>
      {/if}
    </div>
    <div class="mt-2 mb-2">
      <input
        id="username"
        type="text"
        placeholder="your-cool-{customDomainMode ? 'domain.com' : 'handle'}"
        class="input input-bordered {isProcessing
          ? 'input-disabled'
          : ''} w-full"
        bind:value={$registerData.handle}
      />
    </div>
    <div class="text-base-content/75 flex gap-3 items-center">
      <div class="grow">
        <div>Your handle will be</div>
      </div>
      <div>
        <button
          class="text-xs flex gap-1.5 items-center hover:underline {customDomainMode
            ? 'text-primary'
            : 'opacity-50'} hover:opacity-100"
          on:click|preventDefault={toggleCustomDomainMode}
          ><input
            type="checkbox"
            class="toggle toggle-xs toggle-primary"
            checked={customDomainMode ? "checked" : false}
            on:click|preventDefault={() => {
              this.checked = !this.checked;
            }}
          /> Custom domain</button
        >
      </div>
    </div>
    <div class="text-accent text-lg font-semibold mt-1 mb-3">
      @{$registerData.handle ||
        "%"}{#if !customDomainMode}.{$config.domain}{/if}
    </div>

    {#if customDomainMode}
      <div
        class="py-3 px-4 itembox border-primary/10 bg-primary/5 mb-4 text-sm text-base-content/75 grid grid-cols-1 gap-3"
      >
        <div class="text-lg">5-min DNS verification ✨</div>
        <div class="text-base-content/75">
          Add the following DNS record to your domain as <span
            class="inline-block py-0.5 px-1.5 bg-base-300 rounded font-mono"
            >_evermeet.{$registerData.handle || "%"}</span
          > subdomain:
        </div>
        <div class="py-3 px-4 bg-base-300 rounded-lg grid grid-cols-1 gap-2">
          <p>Host:<br /><span class="font-mono">_evermeet</span></p>
          <p>Type:<br /><span class="font-mono">TXT</span></p>
          <div>
            <div>Value:</div>
            <div class="relative mt-2">
              <div
                class="{$did ? '' : 'blur select-none h-10'} font-mono text-md"
              >
                {$did?.did || "did=did:plc:123456789_123456_8901234"}
              </div>
              {#if !$did}
                <div
                  class="absolute btn btn-sm btn-primary top-0 left-1 flex gap-1.5"
                  on:click|preventDefault={generateDid}
                >
                  <Key /> Generate new DID
                </div>
              {/if}
            </div>
          </div>
        </div>
      </div>
    {:else}
      <div
        class="py-2 px-3 itembox mb-4 text-sm text-base-content/75 grid grid-cols-1 gap-2"
      >
        <div class="flex gap-2 items-center">
          <div>
            {#if $registerData.handle?.match(/^[0-9a-zA-Z-]+$/)}<div
                class="text-success"
              >
                <Check tabindex="-1" />
              </div>{:else}<div class="text-error">
                <ExclamationCircle tabindex="-1" />
              </div>{/if}
          </div>
          <div>Only contains letters, numbers, and hyphens</div>
        </div>
        <div class="flex gap-2 items-center">
          <div>
            {#if $registerData.handle?.length >= 3}<div class="text-success">
                <Check tabindex="-1" />
              </div>{:else}<div class="text-error">
                <ExclamationCircle tabindex="-1" />
              </div>{/if}
          </div>
          <div>At least 3 characters</div>
        </div>
      </div>
    {/if}
    <div class={validHandle ? "" : "opacity-50"}>
      <label for="password" class="">Password</label>
      <div class="mt-2 mb-2">
        <input
          id="password"
          type="password"
          placeholder="Password"
          class="input input-bordered {isProcessing || !validHandle
            ? 'input-disabled'
            : ''} w-full"
          bind:value={$registerData.password}
        />
      </div>
      <label for="passwordRepeat" class="">Password (repeat)</label>
      <div class="mt-2">
        <input
          id="passwordRepeat"
          type="password"
          placeholder="Password"
          class="input input-bordered {isProcessing || !validHandle
            ? 'input-disabled'
            : ''} w-full"
          bind:value={$registerData.passwordRepeat}
        />
      </div>
    </div>
    <div class="mt-4">
      <button
        class="btn btn-primary {!validData || isProcessing
          ? 'btn-disabled'
          : ''} w-full"
      >
        {#if !isProcessing}
          Register
        {:else}
          <span class="loading loading-infinity loading-lg"></span>
        {/if}
      </button>
    </div>
  </form>
</div>

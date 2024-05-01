<script>
  import { writable } from "svelte/store";
  import ManagePage from "./ManagePage.svelte";
  import Form from "./Form.svelte";
  import { xrpcCall, blobUpload } from "$lib/api";
  import { onDestroy, onMount, getContext, setContext } from "svelte";
  import { t } from "$lib/i18n";

  export let selectedTab;

  const user = getContext("user");

  // update profile
  const profile = writable({
    name: user.name,
    description: user.description,
    avatarBlob: user.avatarBlob,
  });
  async function submitProfileForm(d) {
    let avatar = undefined;
    if (d.avatarBlob && typeof d.avatarBlob === "object") {
      const blob = await blobUpload({ fetch, user }, d.avatarBlob.data);
      avatar = { $cid: blob?.blob.cid };
    } else if (d.avatarBlob) {
      avatar = { $cid: d.avatarBlob };
    }
    const resp = await xrpcCall(
      { fetch, user },
      "app.evermeet.auth.updateAccount",
      null,
      {
        profile: {
          name: d.name,
          description: d.description,
          avatar,
        },
      },
    );
    if (resp.user) {
      setContext("user", Object.assign(resp.user, { token: user.token }));
      return true;
    }
  }

  // password change
  const passwordChange = writable({});

  // date preferenes
  const datePreferences = writable(user.preferences.date || {});
  async function submitDatePreferencesForm(d) {
    const resp = await xrpcCall(
      { fetch, user },
      "app.evermeet.auth.updateAccount",
      null,
      {
        preferences: {
          date: d,
        },
      },
    );
    if (resp.user) {
      user.set(resp.user);
      return true;
    }
  }
</script>

<ManagePage
  {selectedTab}
  baseUrl="/me/settings"
  title={$t`Settings`}
  tabs={[
    { id: "account", name: $t`Account` },
    { id: "preferences", name: $t`Preferences` },
    { id: "security", name: $t`Security` },
  ]}
>
  {#if selectedTab === "account"}
    <h2 class="manage-heading1">{$t`Handle`}</h2>
    <div>
      {$t`Your handle`}:
      <span class="font-semibold text-accent">@{user.handle}</span>
    </div>

    <h2 id="profile" class="manage-heading1">{$t`Public Profile`}</h2>
    <div class="itembox mt-4">
      <Form
        item={$profile}
        onSubmit={submitProfileForm}
        config={{
          bordered: false,
          user,
          submitButton: $t`Save profile`,
        }}
        schema={{
          type: "object",
          properties: {
            name: {
              title: $t`Display Name`,
              type: "string",
              placeholder: "name",
            },
            avatarBlob: {
              title: $t`Profile Image (avatar)`,
              type: "string",
              view: "image",
            },
            description: {
              title: $t`Description`,
              type: "string",
              view: "textarea-markdown",
            },
          },
        }}
        layout={["avatarBlob", "name", "description"]}
      />
    </div>

    <h2 id="password" class="manage-heading1">{$t`Password`}</h2>
    <div class="itembox mt-4">
      <Form
        item={$passwordChange}
        config={{
          bordered: false,
          user,
          submitButton: $t`Change password`,
        }}
        schema={{
          type: "object",
          properties: {
            currentPassword: {
              title: $t`Current Password`,
              type: "string",
              view: "password",
            },
            newPassword: {
              title: $t`New Password`,
              type: "string",
              view: "password",
            },
            newPasswordRepeat: {
              title: $t`New Password (repeat)`,
              type: "string",
              view: "password",
            },
          },
        }}
        layout={["currentPassword", "newPassword", "newPasswordRepeat"]}
      />
    </div>

    <h2 class="manage-heading1">{$t`Delete Account`}</h2>
    <div>
      <div class="mb-4">
        {$t`If you delete your account, you will lose forever access to your DID.`}
      </div>
      <button class="btn btn-error">{$t`Delete account`}</button>
    </div>
  {/if}

  {#if selectedTab === "preferences"}
    <h2 class="manage-heading1">{$t`Date & Time`}</h2>
    <div class="itembox mt-4">
      <Form
        item={$datePreferences}
        onSubmit={submitDatePreferencesForm}
        config={{
          bordered: false,
          user,
          submitButton: $t`Save`,
        }}
        schema={{
          type: "object",
          properties: {
            hoursFormat: {
              title: $t`Time format`,
              type: "string",
              view: "radio",
              enum: ["24-hour", "12-hour"],
            },
            timezone: {
              title: $t`Timezone`,
              type: "string",
            },
          },
        }}
        layout={[
          "hoursFormat",
          //'timezone',
        ]}
      />
    </div>
  {/if}
</ManagePage>

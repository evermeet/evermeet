<script>
    import { writable } from 'svelte/store'
    import ManagePage from './ManagePage.svelte'
    import { user } from '$lib/stores'
    import Form from './Form.svelte'
    import { xrpcCall, blobUpload } from '$lib/api';

    export let selectedTab;

    $: u = $user

    // update profile
    const profile = writable({
        name: $user.name,
        description: $user.description,
        avatarBlob: $user.avatarBlob,
    })
    async function submitProfileForm (d) {
        let avatar = undefined
        if (d.avatarBlob && typeof(d.avatarBlob) === 'object') {
            const blob = await blobUpload(fetch, d.avatarBlob.data)
            avatar = { $cid: blob?.blob.cid }
        } else if (d.avatarBlob) {
            avatar = { $cid: d.avatarBlob }
        }
        const resp = await xrpcCall(fetch, 'app.evermeet.auth.updateAccount', null, {
            profile: {
                name: d.name,
                description: d.description,
                avatar,
            }
        })
        if (resp.user) {
            user.set(resp.user)
            return true
        }
    }

    // password change
    const passwordChange = writable({})

</script>

<ManagePage {selectedTab}
    baseUrl="/me/settings"
    title="Settings"
    tabs={[
        { id: 'account', name: 'Account' },
        { id: 'preferences', name: 'Preferences' },
        { id: 'security', name: 'Security' },
    ]}
    >

    {#if selectedTab === 'account'}
        <h2 class="manage-heading1">Handle</h2>
        <div>
            Your handle: <span class="font-semibold text-accent">@{$user.handle}</span>
        </div>

        <h2 id="profile" class="manage-heading1">Public Profile</h2>
        <div class="itembox mt-4">
            <Form item={$profile} onSubmit={submitProfileForm} config={{ 
                bordered: false,
                user: $user,
                submitButton: 'Save profile'
            }} schema={{
                type: 'object',
                properties: {
                    name: {
                        title: 'Display Name',
                        type: 'string',
                        placeholder: 'name',
                    },
                    avatarBlob: {
                        title: 'Profile Image (avatar)',
                        type: 'string',
                        view: 'image',
                    },
                    description: {
                        title: 'Description',
                        type: 'string',
                        view: 'textarea',
                    }
                }
            }} layout={[
                'avatarBlob',
                'name',
                'description',
            ]} />
        </div>

        <h2 id="password" class="manage-heading1">Password</h2>
        <div class="itembox mt-4">
            <Form item={$passwordChange} config={{
                bordered: false,
                user: $user,
                submitButton: 'Update password',
            }} schema={{
                type: 'object',
                properties: {
                    currentPassword: {
                        title: 'Current Password',
                        type: 'string',
                        view: 'password'
                    },
                    newPassword: {
                        title: 'New Password',
                        type: 'string',
                        view: 'password'
                    },
                    newPasswordRepeat: {
                        title: 'New Password (repeat)',
                        type: 'string',
                        view: 'password'
                    }
                }
            }} layout={[
                'currentPassword',
                'newPassword',
                'newPasswordRepeat'
            ]} />
        </div>

        <h2 class="manage-heading1">Delete Account</h2>
        <div>
            <div class="mb-4">If you delete your account, you will lose forever access to your DID.</div>
            <button class="btn btn-error">Delete account</button>
        </div>

    {/if}


</ManagePage>
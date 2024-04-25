import { redirect } from '@sveltejs/kit';

export async function load ({ parent }) {
    const data = await parent()
    if (!data.session) {
        redirect(307, '/login')
        return null
    }
    return data
}
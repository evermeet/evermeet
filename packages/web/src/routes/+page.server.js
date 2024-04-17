import { redirect } from '@sveltejs/kit';

export async function load({ locals, parent }) {
    const data = await parent()
    //redirect(302, '/explore')
    return data
}
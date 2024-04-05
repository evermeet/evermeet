
export function load({ locals, cookies }) {
    cookies.delete('evermeet-session-id', { path: '/' });
}
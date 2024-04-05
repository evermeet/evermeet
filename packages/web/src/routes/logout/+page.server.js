
export function load({ locals, cookies }) {
    cookies.delete('deluma-session-id', { path: '/' });
}
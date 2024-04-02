
export async function apiCall (fetch, path) {
    const resp = await fetch("http://localhost:3000/" + path)
    return resp.json()
}
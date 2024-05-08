export async function handle({ event, resolve }) {
  const response = await resolve(event, {
    filterSerializedResponseHeaders: (name) => name.startsWith("content-"),
  });

  return response;
}

export async function handleFetch({ event, request, fetch }) {
  console.log(event.request.headers);

  return fetch(request);
}

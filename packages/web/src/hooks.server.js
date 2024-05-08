export async function handle({ event, resolve }) {
  const response = await resolve(event, {
    filterSerializedResponseHeaders: (name) => name.startsWith("content-"),
  });

  return response;
}

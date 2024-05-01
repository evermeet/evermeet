export async function load({ params, parent }) {
  const data = await parent();
  return {
    tab: params.tab,
  };
}

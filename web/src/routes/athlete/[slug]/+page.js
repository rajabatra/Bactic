export const load = async ({ fetch, params }) => {
  const athleteData = await fetch(`/api/stats/athlete/${params.slug}`, {
    method: 'GET',
  }).then((res) => {
    return res.json();
  });

  return athleteData;
};

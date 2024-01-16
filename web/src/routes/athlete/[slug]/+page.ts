import { Configuration, DefaultApi, type AthleteSummary } from "$lib/api";
import type { PageLoad } from "./$types";

let api = new DefaultApi();
export const load: PageLoad = async ({ params }) => {
  let athleteInfo = await api.statsAthleteIdGet({ id: params.slug });
  return athleteInfo;
};

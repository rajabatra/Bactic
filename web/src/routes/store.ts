import { Configuration, DefaultApi } from "$lib/api";

export const api = new DefaultApi(
  new Configuration({
    basePath: "/api",
  }),
);

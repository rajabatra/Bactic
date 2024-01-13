import type { Handle } from "@sveltejs/kit";
const PROXY_PATH = "/api";
import { env } from "$env/dynamic/private";
import { error } from "@sveltejs/kit";

const API_URL = env.API_URL;

const handleApiProxy: Handle = async ({ event }) => {
  const origin = event.request.headers.get("Origin");

  if (!origin || new URL(origin).origin !== event.url.origin) {
    throw error(403, "Request Forbidden.");
  }

  const strippedPath = event.url.pathname.substring(PROXY_PATH.length);

  const urlPath = `${API_URL}${strippedPath}${event.url.search}`;
  const proxiedURL = new URL(urlPath);

  // Strip header added by SveltKit yet forbidden by underlying HTTP request
  event.request.headers.delete("connection");

  return fetch(proxiedURL.toString(), {
    // propogate the request method and body
    body: event.request.body,
    method: event.request.method,
    headers: event.request.headers,
  }).catch((err) => {
    console.log("Could not proxy API request: ", err);
    throw err;
  });
};

export const handle: Handle = async ({ event, resolve }) => {
  if (event.url.pathname.startsWith(PROXY_PATH)) {
    return await handleApiProxy({ event, resolve });
  } else {
    return await resolve(event);
  }
};

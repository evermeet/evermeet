
import { writable } from "svelte/store";

export const config = writable({});

export const user = writable(null);

export const session = writable(null);

export const eventDetail = writable(null);

export const socket = writable();
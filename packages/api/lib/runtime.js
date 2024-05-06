import os from "node:os";

export function detectRuntime() {
  let rt;
  if (typeof Bun !== "undefined") {
    rt = "bun";
  } else if (typeof Deno !== "undefined") {
    rt = "deno";
  } else {
    rt = "node";
  }
  return {
    name: rt,
    version:
      rt === "bun"
        ? Bun.version
        : rt === "deno"
          ? Deno.version.deno
          : process.version,
    arch: os.arch(),
  };
}

export const runtime = detectRuntime();

export function env(key) {
  if (runtime.name === "deno") {
    return Deno.env.get(key);
  }
  return process.env[key];
}

export function exitRuntime() {
  if (runtime.name === "deno") {
    Deno.exit(1);
  } else {
    process.exit(1);
  }
}

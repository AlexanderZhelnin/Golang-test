import { build } from "./gr-kernel/kernel.mjs";
import { parentPort } from "node:worker_threads";

parentPort.on("message", (v) =>
  parentPort.postMessage(build(v.ls, v.pr, v.rect))
);

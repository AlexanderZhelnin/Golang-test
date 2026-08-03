import { build } from "./gr-kernel/kernel.mjs";
import { parentPort } from "node:worker_threads";
import * as fs from "fs";
const json = JSON.parse(fs.readFileSync("primitives.json", "utf8"));

const ls = json.sort((l1, l2) => {
  if (l1.priority < l2.priority) return -1;
  if (l1.priority > l2.priority) return 1;
  return 0;
});

parentPort.on("message", (v) =>
  parentPort.postMessage(build(ls, v.pr, v.rect))
);

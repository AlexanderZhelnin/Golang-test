import { parentPort } from "node:worker_threads";

parentPort.on("message", (message) => {
  let a = 0;
  let b = 1;
  let c = 0;
  for (let i = 2; i < 2000000; i++) {
    c = a + b;
    a = b;
    b = c;
  }

  parentPort.postMessage(a.toString());
});

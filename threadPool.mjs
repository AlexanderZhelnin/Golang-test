import {
  Worker,
  isMainThread,
  parentPort,
  workerData,
} from "node:worker_threads";

export class ThreadPool {
  constructor(file, max = 10) {
    this.workers = [...Array(max).keys()].map((_) => new Worker(file));
    this.active = [...Array(max).keys()].map((_) => false);
    this.query = [];
  }

  run(data) {
    let callback;
    const result = new Promise((resolve, reject) => {
      callback = resolve;
    });

    this.query.push({ data, callback });
    this.runQuery();
    return result;
  }

  runQuery() {
    if (this.query.length === 0) return;
    const i = this.getIndex();

    if (i < 0) return;

    const q = this.query.shift();

    const worker = this.workers[i];
    worker.once("message", (v) => {
      this.active[i] = false;
      setTimeout(() => this.runQuery());
      q.callback(v);
    });

    worker.postMessage(q.data);
  }

  getIndex() {
    for (let i = 0; i < this.active.length; i++) {
      if (!this.active[i]) {
        this.active[i] = true;
        return i;
      }
    }
    return -1;
  }
}

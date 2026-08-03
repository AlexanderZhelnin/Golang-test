import { build } from "./gr-kernel/kernel.mjs";
import { compareStrings } from "./strings.mjs";
import { ThreadPool } from "./threadPool.mjs";
import * as fs from "fs";
import * as http from "http";
import * as url from "url";

const hostname = "127.0.0.1";
const port = 3001;

const fibonacciPool = new ThreadPool("./fibonacci.js");
const mapPool = new ThreadPool("./map.js");
const mapPoolWithRead = new ThreadPool("./mapWithRead.js");

const ls = JSON.parse(fs.readFileSync("primitives.json", "utf8"));
//const json = JSON.parse(fs.readFileSync("primitives.json", "utf8"));
// const ls = json.sort((l1, l2) => {
//   if (l1.priority < l2.priority) return -1;
//   if (l1.priority > l2.priority) return 1;
//   return 0;
// });
const rect = { left: 1200, bottom: 50, right: 4000, top: 2850 };
const pr = {
  left: rect.left,
  top: rect.top,
  scale: 0.37037037037037035,
  mashtab: 100,
};
const STR1 = "asrgfsadf12421";
const STR2 = "asrgfsadf12321";

const server = http.createServer((req, res) => {
  if (req.url == "/") {
    res.statusCode = 200;
    res.setHeader("Content-Type", "text/plain");
    res.end("Hello World nodeJs!");
    return;
  }

  if (req.url === "/readfile") {
    res.statusCode = 200;
    res.setHeader("Content-Type", "text/plain");
    const data = fs.readFileSync("data.txt", "utf8");
    res.end(data);

    return;
  }

  if (req.url === "/fibonacci") {
    let a = 0;
    let b = 1;
    let c = 0;
    for (let i = 2; i < 2000000; i++) {
      c = a + b;
      a = b;
      b = c;
    }

    res.statusCode = 200;
    res.setHeader("Content-Type", "text/plain");
    res.end(a.toString());

    return;
  }

  if (req.url === "/fibonacci/thread") {
    fibonacciPool.run().then((r) => {
      res.statusCode = 200;
      res.setHeader("Content-Type", "text/plain");
      res.end(r);
    });

    return;
  }

  const q = url.parse(req.url, true);

  if (q.pathname === "/map")
  {
    res.statusCode = 200;
    res.setHeader("Content-Type", "application/json; charset=utf-8");

    res.end(JSON.stringify(build(ls, pr, rect).length.toString()));
    return;
  }

  if (q.pathname === "/mapJSON")
  {
    res.statusCode = 200;
    res.setHeader("Content-Type", "application/json; charset=utf-8");

    res.end(JSON.stringify(build(ls, pr, rect).slice(0, 5)));
    return;
  }


  if (q.pathname === "/map/thread")
  {
    mapPool.run({ ls, pr, rect }).then((r) => {
      res.statusCode = 200;
      res.setHeader("Content-Type", "text/plain");
      res.end(JSON.stringify(r).length.toString());
    });
    return;
  }

  if (q.pathname === "/map/thread/with_thread")
  {
    mapPoolWithRead.run({ pr, rect }).then((r) => {
      res.statusCode = 200;
      res.setHeader("Content-Type", "text/plain");
      res.end(JSON.stringify(r).length.toString());
    });
    return;
  }

  if (req.url === "/naturalsort") {
    res.statusCode = 200;
    res.setHeader("Content-Type", "application/json; charset=utf-8");

    let result = 0;
    for (let i = 0; i < 10000; i++)
      result += compareStrings(STR1 + i, STR2 + i);

    res.end(result.toString());
  }
});

server.listen(port, hostname, () => {
  console.log(`Server running at http://${hostname}:${port}/`);
});

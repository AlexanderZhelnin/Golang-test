import { build } from "./gr-kernel/kernel.mjs";
import { compareStrings } from "./strings.mjs";

import * as cluster from "cluster";
import * as os from "os";
import * as fs from "fs";
import * as http from "http";
import * as url from "url";

const numCPUs = os.cpus().length;
if (cluster.isMaster) {
  for (let i = 0; i < numCPUs; i++) cluster.fork();

  // cluster.on("exit", function (worker, code, signal) {
  //   console.log("worker ${worker.process.pid} died");
  //   cluster.fork();
  // });
} else {
  const hostname = "127.0.0.1";
  const port = 3000;

  const ls = JSON.parse(fs.readFileSync("primitives.json", "utf8"));
  // const json = JSON.parse(fs.readFileSync("primitives.json", "utf8"));
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
    const q = url.parse(req.url, true);

    if (req.url === "/") {
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

    if (q.pathname === "/map") {
      res.statusCode = 200;
      res.setHeader("Content-Type", "application/json; charset=utf-8");

      const x = +q.query.x / 100;
      const y = +q.query.y / 100;
      res.end(
        JSON.stringify(
          build(
            ls,
            { ...pr, let: pr.left + x, top: pr.top + y },
            {
              left: rect.left + x,
              top: rect.top + y,
              right: rect.right,
              bottom: rect.bottom,
            }
          )
        ).length.toString()
      );
      return;
    }

    if (q.pathname === "/mapJSON")
    {
      res.statusCode = 200;
      res.setHeader("Content-Type", "application/json; charset=utf-8");

      const x = +q.query.x / 100;
      const y = +q.query.y / 100;
      res.end(
        JSON.stringify(
          build(
            ls,
            { ...pr, let: pr.left + x, top: pr.top + y },
            {
              left: rect.left + x,
              top: rect.top + y,
              right: rect.right,
              bottom: rect.bottom,
            }
          ).slice(0, 5)
        )
      );
      return;
    }

    if (req.url === "/naturalsort") {
      res.statusCode = 200;
      res.setHeader("Content-Type", "application/json; charset=utf-8");

      let result = 0;
      for (let i = 0; i < 10000; i++)
        result += compareStrings(STR1 + i, STR2 + i);

      res.end(result.toString());
      return;
    }

    res.end("error");
  });

  server.listen(port, hostname, () => {
    console.log(`Server running at http://${hostname}:${port}/`);
  });
}

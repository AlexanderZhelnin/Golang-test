import { build } from "./gr-kernel/kernel.mjs";
import * as fs from "fs";
import express from "express";
import compression from "compression";
const port = 3000;

const json = JSON.parse(fs.readFileSync("primitives.json", "utf8"));
const ls = json.sort((l1, l2) => {
  if (l1.priority < l2.priority) return -1;
  if (l1.priority > l2.priority) return 1;
  return 0;
});
const rect = { left: 1200, bottom: 50, right: 4000, top: 2850 };

const app = express();

app.use(compression());

app.get("/readfile", (req, res) => {
  res.statusCode = 200;
  res.setHeader("Content-Type", "text/plain");
  const data = fs.readFileSync("data.txt", "utf8");
  res.end(data);
});

app.get("/fibonacci", (req, res) => {
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
});

app.get("/map", (req, res) => {
  res.statusCode = 200;
  res.setHeader("Content-Type", "application/json");
    
  const result = build(ls, { left: rect.left, top: rect.top, scale: 0.37037037037037035, mashtab: 100,}, rect);

  res.end(JSON.stringify(result));
});


app.listen(port, () => {
  console.log(`Server running at :${port}/`);
});

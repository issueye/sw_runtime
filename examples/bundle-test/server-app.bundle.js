var __getOwnPropNames = Object.getOwnPropertyNames;
var __commonJS = (cb, mod) => function __require() {
  return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
};

// utils.js
var require_utils = __commonJS({
  "utils.js"(exports2) {
    exports2.add = function(a, b) {
      return a + b;
    };
    exports2.multiply = function(a, b) {
      return a * b;
    };
    exports2.greet = function(name) {
      return `Hello, ${name}!`;
    };
  }
});

// server-app.js
var httpserver = require("httpserver");
var fs = require("fs");
var utils = require_utils();
console.log("=== Server Application ===\n");
console.log("Testing custom module:");
console.log("  add(10, 20) =", utils.add(10, 20));
var app = httpserver.createServer();
app.get("/hello", (req, res) => {
  res.send(utils.greet("Server"));
});
app.get("/math", (req, res) => {
  const result = {
    sum: utils.add(5, 10),
    product: utils.multiply(3, 4)
  };
  res.json(result);
});
console.log("Server configured with routes:");
console.log("  GET /hello");
console.log("  GET /math");
app.listen("38200", () => {
  console.log("\n\u2713 Server ready on http://localhost:38200");
});

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

// math-lib.ts
var require_math_lib = __commonJS({
  "math-lib.ts"(exports2) {
    function square(n) {
      return n * n;
    }
    function cube(n) {
      return n * n * n;
    }
    var PI = 3.14159;
    exports2.square = square;
    exports2.cube = cube;
    exports2.PI = PI;
  }
});

// app.js
var utils = require_utils();
var mathLib = require_math_lib();
console.log("=== Bundle Test Application ===\n");
console.log("1. Utils Module:");
console.log("   5 + 3 =", utils.add(5, 3));
console.log("   5 * 3 =", utils.multiply(5, 3));
console.log("   ", utils.greet("World"));
console.log("\n2. Math Library (TypeScript):");
console.log("   square(4) =", mathLib.square(4));
console.log("   cube(3) =", mathLib.cube(3));
console.log("   PI =", mathLib.PI);
console.log("\n3. Async Support:");
Promise.resolve(42).then((value) => {
  console.log("   Promise resolved with:", value);
});
setTimeout(() => {
  console.log("   Timeout executed after 100ms");
}, 100);
console.log("\n=== Bundle Test Complete ===");

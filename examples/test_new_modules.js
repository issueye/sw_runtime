const { os } = require('fs');
const { util } = require('utils');

console.log("--- OS Module Test ---");
console.log("Hostname:", os.hostname());
console.log("Platform:", os.platform());
console.log("Arch:", os.arch());
console.log("Tmpdir:", os.tmpdir());
console.log("Homedir:", os.homedir());
console.log("Uptime:", os.uptime());
console.log("Total Memory:", os.totalmem());
console.log("Free Memory:", os.freemem());
console.log("OS Type:", os.type());
console.log("Release:", os.release());

const cpus = os.cpus();
console.log("CPUs Count:", cpus.length);
if (cpus.length > 0) {
  console.log("First CPU Model:", cpus[0].model);
}

const interfaces = os.networkInterfaces();
console.log("Network Interfaces:", Object.keys(interfaces).join(", "));

const userInfo = os.userInfo();
console.log("User Info:", JSON.stringify(userInfo));

console.log("\n--- Util Module Test ---");
console.log("Format %s %d:", util.format("hello", 123));
console.log("Format %j:", util.format("obj: %j", { a: 1 }));

const obj1 = { a: [1, 2], b: { c: 3 } };
const obj2 = { a: [1, 2], b: { c: 3 } };
const obj3 = { a: [1, 2], b: { c: 4 } };

console.log("isDeepStrictEqual (1, 2):", util.isDeepStrictEqual(obj1, obj2));
console.log("isDeepStrictEqual (1, 3):", util.isDeepStrictEqual(obj1, obj3));

console.log("isDate:", util.types.isDate(new Date()));
console.log("isRegExp:", util.types.isRegExp(/abc/));

const promise = new Promise((resolve) => resolve());
console.log("isPromise:", util.types.isPromise(promise));

console.log("Inspect:", util.inspect({ foo: "bar", baz: 123 }));

interface EdgeRequest {
  method: string;
  path: string;
  query: Record<string, string>;
  body?: string;
}

interface EdgeResponse {
  status?: number;
  headers?: Record<string, string>;
  body?: string;
  json?: any;
}

// 这些声明是为了让 TypeScript 有类型信息，运行时由 Go 注入。
// 在 SW Runtime 中，全局对象可以通过 global 或 globalThis 访问。

declare const request: EdgeRequest;
declare let response: EdgeResponse | undefined;

type AnyGlobal = typeof globalThis & { request: EdgeRequest; response?: EdgeResponse };

const g = globalThis as AnyGlobal;

// 简单的 edge 逻辑：根据 query.name 返回问候语
const name = request.query["name"] || "World";

const now = new Date().toISOString();

g.response = {
  status: 200,
  headers: {
    "X-Edge-Service": "sw-runtime",
  },
  json: {
    message: `Hello, ${name}!` ,
    method: request.method,
    path: request.path,
    time: now,
  },
};

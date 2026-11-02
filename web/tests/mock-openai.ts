import { createServer, type IncomingMessage, type ServerResponse } from 'node:http';
import type { AddressInfo } from 'node:net';

export interface MockOpenAI {
  url: string;
  hits: Record<string, number>;
  close: () => Promise<void>;
}

const MODELS = [
  { id: 'alpha/fast', owned_by: 'alpha' },
  { id: 'alpha/quick', owned_by: 'alpha' },
  { id: 'beta/hang', owned_by: 'beta' },
  { id: 'beta/broken', owned_by: 'beta' },
];

function readBody(req: IncomingMessage): Promise<string> {
  return new Promise((resolve) => {
    let data = '';
    req.on('data', (c: Buffer) => (data += c.toString()));
    req.on('end', () => resolve(data));
  });
}

function json(res: ServerResponse, status: number, body: unknown) {
  res.writeHead(status, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify(body));
}

/** A tiny OpenAI-compatible server: two fast models, one that hangs and one that is missing. */
export async function startMockOpenAI(hangMs = 120_000): Promise<MockOpenAI> {
  const hits: Record<string, number> = {};
  const server = createServer(async (req, res) => {
    if (req.method === 'GET' && req.url === '/v1/models') {
      json(res, 200, { object: 'list', data: MODELS.map((m) => ({ ...m, object: 'model', created: 1700000000 })) });
      return;
    }
    if (req.method === 'POST' && req.url === '/v1/chat/completions') {
      const body = JSON.parse((await readBody(req)) || '{}') as {
        model?: string;
        stream?: boolean;
        max_tokens?: number;
      };
      const model = body.model ?? '';
      hits[model] = (hits[model] ?? 0) + 1;
      if (model === 'beta/broken' || !MODELS.some((m) => m.id === model)) {
        json(res, 404, {
          error: {
            message: `The model '${model}' does not exist`,
            type: 'invalid_request_error',
            code: 'model_not_found',
          },
        });
        return;
      }
      if (model === 'beta/hang') {
        const timer = setTimeout(() => json(res, 200, completion(model, 'late')), hangMs);
        req.on('close', () => clearTimeout(timer));
        return;
      }
      await new Promise((r) => setTimeout(r, 40 + Math.random() * 40));
      if (body.stream) {
        res.writeHead(200, { 'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache' });
        for (const word of ['one', ' two', ' three', ' four']) {
          res.write(
            `data: ${JSON.stringify({ id: 'c', object: 'chat.completion.chunk', model, choices: [{ index: 0, delta: { content: word } }] })}\n\n`,
          );
          await new Promise((r) => setTimeout(r, 15));
        }
        res.write(
          `data: ${JSON.stringify({ id: 'c', object: 'chat.completion.chunk', model, choices: [{ index: 0, delta: {}, finish_reason: 'stop' }], usage: { prompt_tokens: 12, completion_tokens: 4, total_tokens: 16 } })}\n\n`,
        );
        res.write('data: [DONE]\n\n');
        res.end();
        return;
      }
      json(res, 200, completion(model, 'one two three four'));
      return;
    }
    json(res, 404, { error: { message: 'not found' } });
  });
  await new Promise<void>((resolve) => server.listen(0, '127.0.0.1', resolve));
  const { port } = server.address() as AddressInfo;
  return {
    url: `http://127.0.0.1:${port}`,
    hits,
    close: () =>
      new Promise((resolve) => {
        server.closeAllConnections();
        server.close(() => resolve());
      }),
  };
}

function completion(model: string, content: string) {
  return {
    id: 'cmpl-e2e',
    object: 'chat.completion',
    created: 1700000000,
    model,
    choices: [{ index: 0, message: { role: 'assistant', content }, finish_reason: 'stop' }],
    usage: { prompt_tokens: 12, completion_tokens: 4, total_tokens: 16 },
  };
}

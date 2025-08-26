import { Hono } from "hono";
import { getConnInfo, upgradeWebSocket } from "hono/cloudflare-workers";
import { connect } from "cloudflare:sockets";

const app = new Hono();

app.get(
  "/ws/:host/:port",
  upgradeWebSocket((c) => {
    const { host, port } = c.req.param();
    let sock: Socket | undefined = undefined;
    let writer: WritableStreamDefaultWriter | undefined = undefined;
    const info = getConnInfo(c);
    return {
      onMessage(ev, ws) {
        if (!sock) {
          try {
            sock = connect({
              hostname: host,
              port: parseInt(port),
            });
            setTimeout(async () => {
              const opened = await sock?.opened;
              if (!opened) {
                ws.close();
                console.log({
                  status: "failed",
                  remote: { host, port },
                });
                sock?.close();
              }
            }, 10 * 1000);
            writer = sock.writable.getWriter();
            const data = new Uint8Array(ev.data);
            writer.write(data);
            (async () => {
              try {
                for await (const chunk of sock.readable) {
                  const data = new Uint8Array(chunk);
                  ws.send(data);
                }
                sock.close();
                sock = undefined;
              } catch {
                ws.close();
                sock?.close();
                sock = undefined;
              }
            })();
          } catch (e) {
            const res = JSON.stringify({
              status: "failed",
              error: (e as Error).message,
            });
            ws.send(res);
            ws.close();
          }
        } else {
          const data = new Uint8Array(ev.data);
          try {
            if (!writer) {
              writer = sock.writable.getWriter();
            }
            writer.write(data);
          } catch {
            ws.close();
          }
        }
      },
      onClose(ev) {
        console.log("close websocket case:", ev.reason);
        if (sock) {
          sock.close();
          sock = undefined;
        }
        writer = undefined;
      },
      onError() {
        try {
          if (sock) {
            sock.close();
            sock = undefined;
          }
          writer = undefined;
        } catch (e) {
          console.log(
            "close error:",
            String(e),
          );
        }
      },
    };
  }),
);

export default app;
